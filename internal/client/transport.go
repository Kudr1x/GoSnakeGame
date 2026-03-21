// Package client provides the shared game application logic and UI.
package client

import (
	pb "GoSnakeGame/api/proto/snake/v1"
	"context"
)

// Transport defines the interface for communication with the game server.
type Transport interface {
	CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error)
	JoinGame(ctx context.Context, req *pb.JoinGameRequest) error
	ReceiveState() (*pb.JoinGameResponse, error)
	SendDirection(ctx context.Context, req *pb.SendDirectionRequest) error
	GetTopPlayers(ctx context.Context, req *pb.GetTopPlayersRequest) (*pb.GetTopPlayersResponse, error)
	Close() error
}
