package game

import (
	"GoSnakeGame/internal/config"
	"crypto/rand"
	"encoding/hex"
	"sync"

	pb "GoSnakeGame/api/proto/snake/v1"
)

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
func (rm *RoomManager) CreateRoom(mode pb.GameMode) string {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	roomID := generateRoomID()
	engine := NewEngine(rm.cfg, roomID, mode)
	rm.rooms[roomID] = engine

	// Start the game loop for this room
	go engine.Run(nil)

	return roomID
}

// GetRoom returns the engine for a specific room.
func (rm *RoomManager) GetRoom(roomID string) (*Engine, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	engine, ok := rm.rooms[roomID]

	return engine, ok
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

const roomIDBytes = 4

// generateRoomID generates a random short string for room ID.
func generateRoomID() string {
	b := make([]byte, roomIDBytes)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}
