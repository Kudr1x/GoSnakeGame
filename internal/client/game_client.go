package client

import (
	"context"
	"fmt"
	"sync"

	pb "GoSnakeGame/api/proto/snake/v1"
)

// GameClient manages game state and communication with server.
type GameClient struct {
	playerName string
	roomID     string
	transport  Transport

	mu           sync.RWMutex
	currentState *pb.JoinGameResponse
	currentDir   pb.Direction

	ctx    context.Context
	cancel context.CancelFunc
}

// NewGameClient creates a new game client.
func NewGameClient(transport Transport) *GameClient {
	ctx, cancel := context.WithCancel(context.Background())

	return &GameClient{
		transport: transport,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// CreateAndJoinRoom creates a new room and joins it.
func (gc *GameClient) CreateAndJoinRoom(playerName string, mode pb.GameMode) (string, error) {
	gc.playerName = playerName

	// Create room
	resp, err := gc.transport.CreateRoom(gc.ctx, &pb.CreateRoomRequest{
		PlayerName: playerName,
		Mode:       mode,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create room: %w", err)
	}

	gc.roomID = resp.RoomId

	// Join game
	err = gc.transport.JoinGame(gc.ctx, &pb.JoinGameRequest{
		PlayerName: playerName,
		RoomId:     gc.roomID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to join game: %w", err)
	}

	return gc.roomID, nil
}

// JoinRoom joins an existing room.
func (gc *GameClient) JoinRoom(playerName, roomID string) error {
	gc.playerName = playerName
	gc.roomID = roomID

	err := gc.transport.JoinGame(gc.ctx, &pb.JoinGameRequest{
		PlayerName: playerName,
		RoomId:     roomID,
	})
	if err != nil {
		return fmt.Errorf("failed to join game: %w", err)
	}

	return nil
}

// ReceiveState receives the next game state update.
func (gc *GameClient) ReceiveState() (*pb.JoinGameResponse, error) {
	state, err := gc.transport.ReceiveState()
	if err != nil {
		return nil, err
	}

	gc.mu.Lock()
	gc.currentState = state
	gc.mu.Unlock()

	return state, nil
}

// SendDirection sends a direction command to the server.
func (gc *GameClient) SendDirection(dir pb.Direction) error {
	gc.mu.Lock()
	gc.currentDir = dir
	gc.mu.Unlock()

	return gc.transport.SendDirection(gc.ctx, &pb.SendDirectionRequest{
		PlayerName: gc.playerName,
		RoomId:     gc.roomID,
		Direction:  dir,
	})
}

// GetCurrentState returns the current game state.
func (gc *GameClient) GetCurrentState() *pb.JoinGameResponse {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	return gc.currentState
}

// GetPlayerName returns the player name.
func (gc *GameClient) GetPlayerName() string {
	return gc.playerName
}

// GetRoomID returns the room ID.
func (gc *GameClient) GetRoomID() string {
	return gc.roomID
}

// Close closes the game client.
func (gc *GameClient) Close() error {
	gc.cancel()

	return gc.transport.Close()
}
