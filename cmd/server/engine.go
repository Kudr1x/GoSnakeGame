package main

import (
	pb "GoSnakeGame/api/proto/snake/v1"
	"GoSnakeGame/internal/game"
)

// Engine defines the interface for the game engine.
type Engine interface {
	AddOrUpdatePlayer(name string) *game.PlayerInfo
	RemovePlayer(name string, sessionID int64)
	SetDirection(name string, dir pb.Direction)
	GetSnapshot() *pb.JoinGameResponse
	GetTopPlayers() []*pb.PlayerScore
}
