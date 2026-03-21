package client

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
)

type mockTransport struct {
	joinGameFunc      func(ctx context.Context, req *pb.JoinGameRequest) error
	getTopPlayersFunc func(ctx context.Context, req *pb.GetTopPlayersRequest) (*pb.GetTopPlayersResponse, error)
	sendDirectionFunc func(ctx context.Context, req *pb.SendDirectionRequest) error
	createRoomFunc    func(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error)
	receiveStateFunc  func() (*pb.JoinGameResponse, error)
}

func (m *mockTransport) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	if m.createRoomFunc != nil {
		return m.createRoomFunc(ctx, req)
	}

	return &pb.CreateRoomResponse{}, nil
}

func (m *mockTransport) JoinGame(ctx context.Context, req *pb.JoinGameRequest) error {
	if m.joinGameFunc != nil {
		return m.joinGameFunc(ctx, req)
	}

	return nil
}

func (m *mockTransport) GetTopPlayers(
	ctx context.Context,
	req *pb.GetTopPlayersRequest,
) (*pb.GetTopPlayersResponse, error) {
	if m.getTopPlayersFunc != nil {
		return m.getTopPlayersFunc(ctx, req)
	}

	return &pb.GetTopPlayersResponse{}, nil
}

func (m *mockTransport) SendDirection(ctx context.Context, req *pb.SendDirectionRequest) error {
	if m.sendDirectionFunc != nil {
		return m.sendDirectionFunc(ctx, req)
	}

	return nil
}

func (m *mockTransport) ReceiveState() (*pb.JoinGameResponse, error) {
	if m.receiveStateFunc != nil {
		return m.receiveStateFunc()
	}

	return &pb.JoinGameResponse{}, nil
}

func (m *mockTransport) Close() error {
	return nil
}

func TestGetDirectionFromKey(t *testing.T) {
	t.Parallel()

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

			dir := GetDirectionFromKey(tc.key, tc.prevDir)
			assert.Equal(t, tc.expectedDir, dir)
		})
	}
}

func TestGetStringTop(t *testing.T) {
	t.Parallel()

	mockTrans := &mockTransport{
		getTopPlayersFunc: func(_ context.Context, _ *pb.GetTopPlayersRequest) (*pb.GetTopPlayersResponse, error) {
			return &pb.GetTopPlayersResponse{
				TopPlayers: []*pb.PlayerScore{
					{PlayerName: "p1", Score: 100},
					{PlayerName: "p2", Score: 90},
				},
			}, nil
		},
	}
	gc := &App{
		transport: mockTrans,
		cfg:       &config.ClientConfig{TopPlayersTimeout: 1 * time.Second},
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

	mockTrans := &mockTransport{
		sendDirectionFunc: func(_ context.Context, receivedReq *pb.SendDirectionRequest) error {
			mu.Lock()
			defer mu.Unlock()
			called = true
			capturedReq = receivedReq

			return nil
		},
	}
	gc := &App{
		playerName: "test_player",
		transport:  mockTrans,
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

	mockTrans := &mockTransport{
		receiveStateFunc: func() (*pb.JoinGameResponse, error) {
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

	gc := &App{
		playerName: "test_player",
		transport:  mockTrans,
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
