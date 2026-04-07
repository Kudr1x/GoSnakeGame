package client

import (
	"context"
	"fmt"

	pb "GoSnakeGame/api/proto/snake/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCTransport implements Transport interface using gRPC.
type GRPCTransport struct {
	conn   *grpc.ClientConn
	client pb.SnakeGameServiceClient
	stream pb.SnakeGameService_JoinGameClient
	ctx    context.Context
	cancel context.CancelFunc
}

// NewGRPCTransport creates a new gRPC transport.
func NewGRPCTransport(ctx context.Context, serverAddr string) (*GRPCTransport, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	return &GRPCTransport{
		conn:   conn,
		client: pb.NewSnakeGameServiceClient(conn),
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// CreateRoom creates a new game room.
func (t *GRPCTransport) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	return t.client.CreateRoom(ctx, req)
}

// JoinGame joins a game room and starts receiving updates.
func (t *GRPCTransport) JoinGame(ctx context.Context, req *pb.JoinGameRequest) error {
	stream, err := t.client.JoinGame(ctx, req)
	if err != nil {
		return err
	}

	t.stream = stream

	return nil
}

// ReceiveState receives the next game state update.
func (t *GRPCTransport) ReceiveState() (*pb.JoinGameResponse, error) {
	if t.stream == nil {
		return nil, fmt.Errorf("stream is nil")
	}

	return t.stream.Recv()
}

// SendDirection sends a direction command to the server.
func (t *GRPCTransport) SendDirection(ctx context.Context, req *pb.SendDirectionRequest) error {
	_, err := t.client.SendDirection(ctx, req)

	return err
}

// GetTopPlayers retrieves the top players leaderboard.
func (t *GRPCTransport) GetTopPlayers(
	ctx context.Context,
	req *pb.GetTopPlayersRequest,
) (*pb.GetTopPlayersResponse, error) {
	return t.client.GetTopPlayers(ctx, req)
}

// Close closes the transport connection.
func (t *GRPCTransport) Close() error {
	t.cancel()

	if t.stream != nil {
		_ = t.stream.CloseSend()
	}

	if t.conn != nil {
		return t.conn.Close()
	}

	return nil
}
