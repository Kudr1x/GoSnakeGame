// Package main implements the server for the Snake Game.
package main

import (
	"GoSnakeGame/internal/config"
	"GoSnakeGame/internal/game"
	"GoSnakeGame/internal/logger"
	"GoSnakeGame/internal/metrics"
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	pb "GoSnakeGame/api/proto/snake/v1"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
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
	start := time.Now()
	defer func() {
		metrics.RequestDuration.WithLabelValues("CreateRoom").Observe(time.Since(start).Seconds())
	}()

	roomID, err := s.roomManager.CreateRoom(req.Mode)
	if err != nil {
		metrics.ErrorsTotal.WithLabelValues("create_room").Inc()

		if errors.Is(err, game.ErrMaxRoomsReached) {
			return nil, status.Errorf(codes.ResourceExhausted, "server is full")
		}

		return nil, status.Errorf(codes.Internal, "failed to create room: %v", err)
	}

	logger.Get().Info("room created",
		zap.String("room_id", roomID),
		zap.String("mode", req.Mode.String()))

	inviteLink := fmt.Sprintf("https://s.kudrix.com/#%s", roomID)

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
	logger.Get().Info("player joining room",
		zap.String("player", req.PlayerName),
		zap.String("room_id", req.RoomId))

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
		logger.Get().Info("player disconnected",
			zap.String("player", req.PlayerName),
			zap.String("room_id", req.RoomId))
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
	if err := logger.Init(false); err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	defer func() {
		_ = logger.Sync()
	}()

	cfg := config.DefaultServerConfig()
	cfg.ParseFlags(flag.CommandLine)
	flag.Parse()

	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		logger.Get().Fatal("failed to listen",
			zap.String("addr", cfg.Addr),
			zap.Error(err))
	}

	rm := game.NewRoomManager(cfg)

	go func() {
		const cleanupInterval = 30 * time.Second
		ticker := time.NewTicker(cleanupInterval)

		defer ticker.Stop()

		for range ticker.C {
			rm.CleanupEmptyRooms()
		}
	}()

	// Start metrics server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		metricsServer := &http.Server{
			Addr:              ":9090",
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		}

		logger.Get().Info("metrics server listening on :9090")

		if err := metricsServer.ListenAndServe(); err != nil {
			logger.Get().Error("metrics server error", zap.Error(err))
		}
	}()

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
		logger.Get().Info("stopping server")
		s.GracefulStop()
	}()

	logger.Get().Info("gRPC engine listening", zap.String("addr", cfg.Addr))

	if err := s.Serve(lis); err != nil {
		logger.Get().Fatal("gRPC server error", zap.Error(err))
	}
}
