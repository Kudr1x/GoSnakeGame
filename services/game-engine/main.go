// Package main implements the server for the Snake Game.
package main

import (
	"GoSnakeGame/internal/config"
	"GoSnakeGame/internal/game"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	pb "GoSnakeGame/api/proto/snake/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gameServer struct {
	pb.UnimplementedSnakeGameServiceServer
	roomManager *game.RoomManager
	cfg         *config.ServerConfig
}

// CreateRoom handles room creation.
func (s *gameServer) CreateRoom(_ context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	roomID := s.roomManager.CreateRoom(req.Mode)
	log.Printf("room %s created with mode %v", roomID, req.Mode)

	inviteLink := fmt.Sprintf("https://s.kudrix.com/%s", roomID)

	return &pb.CreateRoomResponse{
		RoomId:     roomID,
		InviteLink: inviteLink,
	}, nil
}

// GetTopPlayers returns the list of top players across all active rooms.
func (s *gameServer) GetTopPlayers(_ context.Context, _ *pb.GetTopPlayersRequest) (*pb.GetTopPlayersResponse, error) {
	playerScores := s.roomManager.GetTopPlayers()

	sort.SliceStable(playerScores, func(i, j int) bool {
		if playerScores[i].Score == playerScores[j].Score {
			return playerScores[i].PlayerName < playerScores[j].PlayerName
		}

		return playerScores[i].Score > playerScores[j].Score
	})

	if len(playerScores) > s.cfg.TopPlayersLimit {
		playerScores = playerScores[:s.cfg.TopPlayersLimit]
	}

	return &pb.GetTopPlayersResponse{
		TopPlayers: playerScores,
	}, nil
}

// JoinGame handles a new player joining a game room.
func (s *gameServer) JoinGame(req *pb.JoinGameRequest, stream pb.SnakeGameService_JoinGameServer) error {
	log.Printf("player %s joining room %s", req.PlayerName, req.RoomId)

	engine, ok := s.roomManager.GetRoom(req.RoomId)
	if !ok {
		return status.Errorf(codes.NotFound, "room not found")
	}

	p := engine.AddOrUpdatePlayer(req.PlayerName)
	if p == nil {
		return status.Errorf(codes.ResourceExhausted, "room is full")
	}

	sessionID := p.SessionID

	defer func() {
		engine.RemovePlayer(req.PlayerName, sessionID)
		log.Printf("player %s disconnected from room %s", req.PlayerName, req.RoomId)
	}()

	return s.gameLoop(engine, p, stream)
}

func (s *gameServer) gameLoop(
	engine *game.Engine,
	p *game.PlayerInfo,
	stream pb.SnakeGameService_JoinGameServer,
) error {
	ctx := stream.Context()
	ticker := time.NewTicker(s.cfg.SendInterval)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("context error: %w", err)
			}

			return nil
		case <-ticker.C:
			state := engine.GetSnapshot()

			if err := stream.Send(state); err != nil {
				return fmt.Errorf("failed to send game state: %w", err)
			}

			if !p.IsAlive() {
				time.Sleep(s.cfg.DeathWaitTime)

				return nil
			}
		}
	}
}

// SendDirection updates the direction of the player's snake.
func (s *gameServer) SendDirection(_ context.Context, req *pb.SendDirectionRequest) (*pb.SendDirectionResponse, error) {
	engine, ok := s.roomManager.GetRoom(req.RoomId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "room not found")
	}

	engine.SetDirection(req.PlayerName, req.Direction)

	return &pb.SendDirectionResponse{}, nil
}

func main() {
	cfg := config.DefaultServerConfig()
	cfg.ParseFlags(flag.CommandLine)
	flag.Parse()

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Printf("failed to listen on %s: %v", cfg.Addr, err)
		os.Exit(1)
	}

	rm := game.NewRoomManager(cfg)

	server := &gameServer{
		roomManager: rm,
		cfg:         cfg,
	}

	s := grpc.NewServer()
	pb.RegisterSnakeGameServiceServer(s, server)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("stopping server")
		s.GracefulStop()
	}()

	log.Printf("gRPC engine listening on %s", cfg.Addr)

	if err := s.Serve(lis); err != nil {
		log.Printf("gRPC server error: %v", err)
		os.Exit(1)
	}
}
