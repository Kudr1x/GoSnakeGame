package main

import (
	"context"

	pb "GoSnakeGame/api/proto/snake/v1"

	"google.golang.org/grpc"
)

// SnakeGameClient defines the interface for the gRPC client.
type SnakeGameClient interface {
	JoinGame(ctx context.Context, in *pb.JoinGameRequest, opts ...grpc.CallOption) (
		pb.SnakeGameService_JoinGameClient, error)
	GetTopPlayers(ctx context.Context, in *pb.GetTopPlayersRequest, opts ...grpc.CallOption) (
		*pb.GetTopPlayersResponse, error)
	SendDirection(ctx context.Context, in *pb.SendDirectionRequest, opts ...grpc.CallOption) (
		*pb.SendDirectionResponse, error)
}
