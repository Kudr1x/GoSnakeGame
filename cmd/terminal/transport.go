package main

import (
	"context"
	"fmt"

	pb "GoSnakeGame/api/proto/snake/v1"

	"google.golang.org/grpc"
)

// GRPCTransport implements the client.Transport interface using gRPC.
type GRPCTransport struct {
	conn   *grpc.ClientConn
	client pb.SnakeGameServiceClient
	stream pb.SnakeGameService_JoinGameClient
}

// NewGRPCTransport creates a new GRPCTransport.
func NewGRPCTransport(conn *grpc.ClientConn) *GRPCTransport {
	return &GRPCTransport{
		conn:   conn,
		client: pb.NewSnakeGameServiceClient(conn),
	}
}

// CreateRoom creates a new game room.
func (t *GRPCTransport) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	resp, err := t.client.CreateRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create room failed: %w", err)
	}

	return resp, nil
}

// JoinGame joins an existing game room.
func (t *GRPCTransport) JoinGame(ctx context.Context, req *pb.JoinGameRequest) error {
	stream, err := t.client.JoinGame(ctx, req)
	if err != nil {
		return fmt.Errorf("join game failed: %w", err)
	}

	t.stream = stream

	return nil
}

// ReceiveState receives the next game state from the stream.
func (t *GRPCTransport) ReceiveState() (*pb.JoinGameResponse, error) {
	if t.stream == nil {
		return nil, fmt.Errorf("stream is nil: %w", context.Canceled)
	}

	state, err := t.stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("receive state failed: %w", err)
	}

	return state, nil
}

// SendDirection sends a player's movement direction.
func (t *GRPCTransport) SendDirection(ctx context.Context, req *pb.SendDirectionRequest) error {
	_, err := t.client.SendDirection(ctx, req)
	if err != nil {
		return fmt.Errorf("send direction failed: %w", err)
	}

	return nil
}

// GetTopPlayers retrieves the top players list.
func (t *GRPCTransport) GetTopPlayers(
	ctx context.Context,
	req *pb.GetTopPlayersRequest,
) (*pb.GetTopPlayersResponse, error) {
	resp, err := t.client.GetTopPlayers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get top players failed: %w", err)
	}

	return resp, nil
}

// Close closes the underlying connection.
func (t *GRPCTransport) Close() error {
	err := t.conn.Close()
	if err != nil {
		return fmt.Errorf("close connection failed: %w", err)
	}

	return nil
}
