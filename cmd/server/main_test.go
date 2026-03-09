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

type mockEngine struct {
	getTopPlayersFunc     func() []*pb.PlayerScore
	setDirectionFunc      func(name string, dir pb.Direction)
	addOrUpdatePlayerFunc func(name string) *game.PlayerInfo
	removePlayerFunc      func(name string, sessionID int64)
	getSnapshotFunc       func() *pb.JoinGameResponse
}

func (m *mockEngine) AddOrUpdatePlayer(name string) *game.PlayerInfo {
	if m.addOrUpdatePlayerFunc != nil {
		return m.addOrUpdatePlayerFunc(name)
	}

	return nil
}

func (m *mockEngine) RemovePlayer(name string, sessionID int64) {
	if m.removePlayerFunc != nil {
		m.removePlayerFunc(name, sessionID)
	}
}

func (m *mockEngine) SetDirection(name string, dir pb.Direction) {
	if m.setDirectionFunc != nil {
		m.setDirectionFunc(name, dir)
	}
}

func (m *mockEngine) GetSnapshot() *pb.JoinGameResponse {
	if m.getSnapshotFunc != nil {
		return m.getSnapshotFunc()
	}

	return nil
}

func (m *mockEngine) GetTopPlayers() []*pb.PlayerScore {
	if m.getTopPlayersFunc != nil {
		return m.getTopPlayersFunc()
	}

	return nil
}

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

func TestGameServer_GetTopPlayers(t *testing.T) {
	t.Parallel()

	mockEngine := &mockEngine{
		getTopPlayersFunc: func() []*pb.PlayerScore {
			return []*pb.PlayerScore{
				{PlayerName: "player2", Score: 100},
				{PlayerName: "player1", Score: 200},
			}
		},
	}
	server := &gameServer{
		engine: mockEngine,
		cfg:    config.DefaultServerConfig(),
	}

	resp, err := server.GetTopPlayers(context.Background(), &pb.GetTopPlayersRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.TopPlayers, 2)
	assert.Equal(t, "player1", resp.TopPlayers[0].PlayerName)
	assert.Equal(t, "player2", resp.TopPlayers[1].PlayerName)
}

func TestGameServer_SendDirection(t *testing.T) {
	t.Parallel()

	var called bool

	var playerName string

	var direction pb.Direction

	mockEngine := &mockEngine{
		setDirectionFunc: func(name string, dir pb.Direction) {
			called = true
			playerName = name
			direction = dir
		},
	}
	server := &gameServer{engine: mockEngine}

	req := &pb.SendDirectionRequest{
		PlayerName: "test_player",
		Direction:  pb.Direction_DIRECTION_LEFT,
	}

	_, err := server.SendDirection(context.Background(), req)
	assert.NoError(t, err)

	assert.True(t, called)
	assert.Equal(t, "test_player", playerName)
	assert.Equal(t, pb.Direction_DIRECTION_LEFT, direction)
}

func TestGameServer_JoinGame(t *testing.T) {
	t.Parallel()

	player := &game.PlayerInfo{Name: "test_player"}
	player.SetAlive(true)
	player.SetSessionID(123)

	mockEngine := &mockEngine{
		addOrUpdatePlayerFunc: func(_ string) *game.PlayerInfo {
			return player
		},
		getSnapshotFunc: func() *pb.JoinGameResponse {
			return &pb.JoinGameResponse{
				Players: []*pb.Player{{Name: "test_player", Alive: player.IsAlive()}},
			}
		},
	}
	server := &gameServer{
		engine: mockEngine,
		cfg:    &config.ServerConfig{SendInterval: 10 * time.Millisecond, DeathWaitTime: 5 * time.Millisecond},
	}
	mockStream := &mockJoinGameServer{ctx: context.Background()}

	go func() {
		time.Sleep(25 * time.Millisecond) // allow 2 sends
		player.SetAlive(false)
	}()

	err := server.JoinGame(&pb.JoinGameRequest{PlayerName: "test_player"}, mockStream)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(mockStream.sentStates), 2)
	assert.False(t, player.IsAlive())
}
