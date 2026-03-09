package main

import (
	"GoSnakeGame/internal/config"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	pb "GoSnakeGame/api/proto/snake/v1"

	"fyne.io/fyne/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockSnakeGameClient struct {
	joinGameFunc func(ctx context.Context, in *pb.JoinGameRequest, opts ...grpc.CallOption) (
		pb.SnakeGameService_JoinGameClient, error)
	getTopPlayersFunc func(ctx context.Context, in *pb.GetTopPlayersRequest, opts ...grpc.CallOption) (
		*pb.GetTopPlayersResponse, error)
	sendDirectionFunc func(ctx context.Context, in *pb.SendDirectionRequest, opts ...grpc.CallOption) (
		*pb.SendDirectionResponse, error)
}

func (m *mockSnakeGameClient) JoinGame(ctx context.Context, in *pb.JoinGameRequest, opts ...grpc.CallOption) (
	pb.SnakeGameService_JoinGameClient, error,
) {
	if m.joinGameFunc != nil {
		return m.joinGameFunc(ctx, in, opts...)
	}

	return &mockJoinGameClient{}, nil
}

func (m *mockSnakeGameClient) GetTopPlayers(ctx context.Context, in *pb.GetTopPlayersRequest, opts ...grpc.CallOption) (
	*pb.GetTopPlayersResponse, error,
) {
	if m.getTopPlayersFunc != nil {
		return m.getTopPlayersFunc(ctx, in, opts...)
	}

	return &pb.GetTopPlayersResponse{}, nil
}

func (m *mockSnakeGameClient) SendDirection(ctx context.Context, in *pb.SendDirectionRequest, opts ...grpc.CallOption) (
	*pb.SendDirectionResponse, error,
) {
	if m.sendDirectionFunc != nil {
		return m.sendDirectionFunc(ctx, in, opts...)
	}

	return &pb.SendDirectionResponse{}, nil
}

type mockJoinGameClient struct {
	grpc.ClientStream
	recvFunc func() (*pb.JoinGameResponse, error)
}

func (m *mockJoinGameClient) Recv() (*pb.JoinGameResponse, error) {
	if m.recvFunc != nil {
		return m.recvFunc()
	}

	return &pb.JoinGameResponse{}, nil
}

func (m *mockJoinGameClient) Header() (metadata.MD, error) { return metadata.MD{}, nil }
func (m *mockJoinGameClient) Trailer() metadata.MD         { return nil }
func (m *mockJoinGameClient) CloseSend() error             { return nil }
func (m *mockJoinGameClient) Context() context.Context     { return context.Background() }
func (m *mockJoinGameClient) SendMsg(_ any) error          { return nil }
func (m *mockJoinGameClient) RecvMsg(_ any) error          { return nil }

func TestGetDirectionFromKey(t *testing.T) {
	t.Parallel()

	gc := &gameClient{}

	testCases := []struct {
		name        string
		key         fyne.KeyName
		prevDir     pb.Direction
		expectedDir pb.Direction
	}{
		{"Up", fyne.KeyUp, pb.Direction_DIRECTION_LEFT, pb.Direction_DIRECTION_UP},
		{"Down", fyne.KeyDown, pb.Direction_DIRECTION_LEFT, pb.Direction_DIRECTION_DOWN},
		{"Left", fyne.KeyLeft, pb.Direction_DIRECTION_UP, pb.Direction_DIRECTION_LEFT},
		{"Right", fyne.KeyRight, pb.Direction_DIRECTION_UP, pb.Direction_DIRECTION_RIGHT},
		{"Up_NoChange", fyne.KeyUp, pb.Direction_DIRECTION_DOWN, pb.Direction_DIRECTION_UNSPECIFIED},
		{"Down_NoChange", fyne.KeyDown, pb.Direction_DIRECTION_UP, pb.Direction_DIRECTION_UNSPECIFIED},
		{"Left_NoChange", fyne.KeyLeft, pb.Direction_DIRECTION_RIGHT, pb.Direction_DIRECTION_UNSPECIFIED},
		{"Right_NoChange", fyne.KeyRight, pb.Direction_DIRECTION_LEFT, pb.Direction_DIRECTION_UNSPECIFIED},
		{"OtherKey", "a", pb.Direction_DIRECTION_UP, pb.Direction_DIRECTION_UNSPECIFIED},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := gc.getDirectionFromKey(tc.key, tc.prevDir)
			assert.Equal(t, tc.expectedDir, dir)
		})
	}
}

func TestGetStringTop(t *testing.T) {
	t.Parallel()

	mockClient := &mockSnakeGameClient{
		getTopPlayersFunc: func(_ context.Context, _ *pb.GetTopPlayersRequest, _ ...grpc.CallOption) (
			*pb.GetTopPlayersResponse, error,
		) {
			return &pb.GetTopPlayersResponse{
				TopPlayers: []*pb.PlayerScore{
					{PlayerName: "p1", Score: 100},
					{PlayerName: "p2", Score: 90},
				},
			}, nil
		},
	}
	gc := &gameClient{
		client: mockClient,
		cfg:    &config.ClientConfig{TopPlayersTimeout: 1 * time.Second},
	}

	topStr := gc.getStringTop()
	assert.Contains(t, topStr, "p1: 100")
	assert.Contains(t, topStr, "p2: 90")
}

func TestSendDirection(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var called bool

	var capturedReq *pb.SendDirectionRequest

	mockClient := &mockSnakeGameClient{
		sendDirectionFunc: func(_ context.Context, receivedReq *pb.SendDirectionRequest, _ ...grpc.CallOption) (
			*pb.SendDirectionResponse, error,
		) {
			mu.Lock()
			defer mu.Unlock()
			called = true
			capturedReq = receivedReq

			return &pb.SendDirectionResponse{}, nil
		},
	}
	gc := &gameClient{
		playerName: "test_player",
		client:     mockClient,
		dirCh:      make(chan pb.Direction, 1),
		cfg:        &config.ClientConfig{DirectionTimeout: 1 * time.Second},
	}

	stopCh := make(chan struct{})
	go gc.sendDirection(stopCh)

	gc.dirCh <- pb.Direction_DIRECTION_RIGHT

	time.Sleep(50 * time.Millisecond)

	close(stopCh)

	mu.Lock()
	defer mu.Unlock()
	assert.True(t, called)
	assert.Equal(t, "test_player", capturedReq.PlayerName)
	assert.Equal(t, pb.Direction_DIRECTION_RIGHT, capturedReq.Direction)
}

func TestReceiveGameState(t *testing.T) {
	t.Parallel()

	states := []*pb.JoinGameResponse{
		{Players: []*pb.Player{{Name: "test_player", Alive: true}}},
		{Players: []*pb.Player{{Name: "test_player", Alive: false}}},
	}

	var mu sync.Mutex

	i := 0

	mockStream := &mockJoinGameClient{
		recvFunc: func() (*pb.JoinGameResponse, error) {
			mu.Lock()
			defer mu.Unlock()

			if i < len(states) {
				state := states[i]
				i++

				return state, nil
			}

			return nil, io.EOF
		},
	}

	var gameOverMu sync.Mutex

	gameOverCalled := false

	gc := &gameClient{
		playerName: "test_player",
		stream:     mockStream,
		gameOver: func() {
			gameOverMu.Lock()
			defer gameOverMu.Unlock()
			gameOverCalled = true
		},
	}

	stopCh := make(chan struct{})
	go gc.receiveGameState(stopCh)

	time.Sleep(100 * time.Millisecond)

	close(stopCh)

	gameOverMu.Lock()
	defer gameOverMu.Unlock()
	assert.True(t, gameOverCalled)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 2, i)

	gc.mu.RLock()
	defer gc.mu.RUnlock()
	assert.False(t, gc.currentState.Players[0].Alive)
}
