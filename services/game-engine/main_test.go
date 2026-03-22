package main

import (
	"GoSnakeGame/internal/config"
	"GoSnakeGame/internal/game"
	"context"
	"testing"
	"time"

	pb "GoSnakeGame/api/proto/snake/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockJoinGameServer struct {
	grpc.ServerStream
	ctx        context.Context
	sentStates []*pb.JoinGameResponse
}

func (m *mockJoinGameServer) Send(resp *pb.JoinGameResponse) error {
	m.sentStates = append(m.sentStates, resp)

	return nil
}

func (m *mockJoinGameServer) Context() context.Context {
	return m.ctx
}

func TestGameServer_CreateRoom(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultServerConfig()
	rm := game.NewRoomManager(cfg)
	server := &gameServer{
		roomManager: rm,
		cfg:         cfg,
	}

	req := &pb.CreateRoomRequest{
		PlayerName: "player1",
		Mode:       pb.GameMode_MODE_1V1,
	}

	resp, err := server.CreateRoom(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.RoomId)
	assert.Contains(t, resp.InviteLink, "/#"+resp.RoomId)

	_, ok := rm.GetRoom(resp.RoomId)
	assert.True(t, ok)
}

func TestGameServer_SendDirection(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultServerConfig()
	rm := game.NewRoomManager(cfg)
	server := &gameServer{
		roomManager: rm,
		cfg:         cfg,
	}

	roomID, _ := rm.CreateRoom(pb.GameMode_MODE_SOLO)
	engine, _ := rm.GetRoom(roomID)
	engine.AddOrUpdatePlayer("test_player")

	req := &pb.SendDirectionRequest{
		PlayerName: "test_player",
		RoomId:     roomID,
		Direction:  pb.Direction_DIRECTION_LEFT,
	}

	_, err := server.SendDirection(context.Background(), req)
	assert.NoError(t, err)
}

func TestGameServer_JoinGame(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultServerConfig()
	cfg.SendInterval = 10 * time.Millisecond
	cfg.DeathWaitTime = 5 * time.Millisecond
	cfg.Width = 20
	cfg.Height = 20
	rm := game.NewRoomManager(cfg)
	server := &gameServer{
		roomManager: rm,
		cfg:         cfg,
	}

	roomID, _ := rm.CreateRoom(pb.GameMode_MODE_SOLO)

	ctx, cancel := context.WithCancel(context.Background())
	mockStream := &mockJoinGameServer{ctx: ctx}

	go func() {
		time.Sleep(25 * time.Millisecond) // allow some sends
		cancel()
	}()

	err := server.JoinGame(&pb.JoinGameRequest{PlayerName: "test_player", RoomId: roomID}, mockStream)
	assert.NoError(t, err)
}
