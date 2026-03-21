// Package main implements the server for the Snake Game.
package main

import (
	"GoSnakeGame/internal/config"
	"GoSnakeGame/internal/game"
	"context"
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
)

// Engine defines the interface for the game engine.
type Engine interface {
	AddOrUpdatePlayer(name string) *game.PlayerInfo
	RemovePlayer(name string, sessionID int64)
	SetDirection(name string, dir pb.Direction)
	GetSnapshot() *pb.JoinGameResponse
	GetTopPlayers() []*pb.PlayerScore
	Run(onPlayerDie func(name string))
}

type gameServer struct {
	pb.UnimplementedSnakeGameServiceServer
	engine Engine
	cfg    *config.ServerConfig
}

// GetTopPlayers returns the list of top 10 players by their best score.
func (s *gameServer) GetTopPlayers(_ context.Context, _ *pb.GetTopPlayersRequest) (*pb.GetTopPlayersResponse, error) {
	playerScores := s.engine.GetTopPlayers()

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

// JoinGame handles a new player joining the game and streams game updates.
func (s *gameServer) JoinGame(req *pb.JoinGameRequest, stream pb.SnakeGameService_JoinGameServer) error {
	log.Printf("player %s joined", req.PlayerName)

	p := s.engine.AddOrUpdatePlayer(req.PlayerName)
	sessionID := p.SessionID

	defer func() {
		s.engine.RemovePlayer(req.PlayerName, sessionID)
		log.Printf("player %s disconnected", req.PlayerName)
	}()

	return s.gameLoop(p, stream)
}

func (s *gameServer) gameLoop(p *game.PlayerInfo, stream pb.SnakeGameService_JoinGameServer) error {
	ctx := stream.Context()
	ticker := time.NewTicker(s.cfg.SendInterval)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil && err != context.Canceled {
				return fmt.Errorf("context error: %w", err)
			}

			return nil
		case <-ticker.C:
			state := s.engine.GetSnapshot()

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
	s.engine.SetDirection(req.PlayerName, req.Direction)

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

	engine := game.NewEngine(cfg)

	go engine.Run(func(name string) {
		log.Printf("player %s died", name)
	})

	server := &gameServer{
		engine: engine,
		cfg:    cfg,
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
