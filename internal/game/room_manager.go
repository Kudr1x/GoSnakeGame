package game

import (
	"GoSnakeGame/internal/config"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"

	pb "GoSnakeGame/api/proto/snake/v1"
)

const (
	maxRooms    = 100
	roomIDBytes = 4
)

// ErrMaxRoomsReached is returned when the server cannot create more rooms.
var ErrMaxRoomsReached = errors.New("maximum number of rooms reached")

// RoomManager manages multiple game rooms.
type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*Engine
	cfg   *config.ServerConfig
}

// NewRoomManager creates a new room manager.
func NewRoomManager(cfg *config.ServerConfig) *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Engine),
		cfg:   cfg,
	}
}

// CreateRoom creates a new game room with the specified mode.
func (rm *RoomManager) CreateRoom(mode pb.GameMode) (string, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if len(rm.rooms) >= maxRooms {
		return "", ErrMaxRoomsReached
	}

	roomID := ""

	for range 10 {
		tempID := generateRoomID()

		if _, exists := rm.rooms[tempID]; !exists {
			roomID = tempID

			break
		}
	}

	if roomID == "" {
		return "", errors.New("failed to generate unique room ID")
	}

	engine := NewEngine(rm.cfg, roomID, mode)
	rm.rooms[roomID] = engine

	// Start the game loop for this room
	go engine.Run(nil)

	return roomID, nil
}

// GetRoom returns the engine for a specific room.
func (rm *RoomManager) GetRoom(roomID string) (*Engine, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	engine, ok := rm.rooms[roomID]

	return engine, ok
}

// CleanupEmptyRooms removes empty rooms to free resources.
func (rm *RoomManager) CleanupEmptyRooms() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for id, engine := range rm.rooms {
		if engine.IsEmpty() {
			engine.Stop()
			delete(rm.rooms, id)
		}
	}
}

// GetTopPlayers returns the top players across all rooms.
func (rm *RoomManager) GetTopPlayers() []*pb.PlayerScore {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var allScores []*pb.PlayerScore
	for _, engine := range rm.rooms {
		allScores = append(allScores, engine.GetTopPlayers()...)
	}

	return allScores
}

// generateRoomID generates a random short string for room ID.
func generateRoomID() string {
	b := make([]byte, roomIDBytes)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}
