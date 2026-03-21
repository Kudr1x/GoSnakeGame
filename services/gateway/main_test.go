package main

import (
	pb "GoSnakeGame/api/proto/snake/v1"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type mockSnakeGameClient struct {
	mock.Mock
}

func (m *mockSnakeGameClient) JoinGame(
	ctx context.Context,
	in *pb.JoinGameRequest,
	opts ...grpc.CallOption,
) (pb.SnakeGameService_JoinGameClient, error) {
	args := m.Called(ctx, in, opts)

	err := args.Error(1)
	if err != nil {
		return args.Get(0).(pb.SnakeGameService_JoinGameClient), fmt.Errorf("mock error: %w", err)
	}

	return args.Get(0).(pb.SnakeGameService_JoinGameClient), nil
}

func (m *mockSnakeGameClient) SendDirection(
	ctx context.Context,
	in *pb.SendDirectionRequest,
	opts ...grpc.CallOption,
) (*pb.SendDirectionResponse, error) {
	args := m.Called(ctx, in, opts)

	err := args.Error(1)
	if err != nil {
		return args.Get(0).(*pb.SendDirectionResponse), fmt.Errorf("mock error: %w", err)
	}

	return args.Get(0).(*pb.SendDirectionResponse), nil
}

func (m *mockSnakeGameClient) GetTopPlayers(
	ctx context.Context,
	in *pb.GetTopPlayersRequest,
	opts ...grpc.CallOption,
) (*pb.GetTopPlayersResponse, error) {
	args := m.Called(ctx, in, opts)

	err := args.Error(1)
	if err != nil {
		return args.Get(0).(*pb.GetTopPlayersResponse), fmt.Errorf("mock error: %w", err)
	}

	return args.Get(0).(*pb.GetTopPlayersResponse), nil
}

type mockJoinGameClient struct {
	grpc.ClientStream
	mock.Mock
}

func (m *mockJoinGameClient) Recv() (*pb.JoinGameResponse, error) {
	args := m.Called()

	err := args.Error(1)
	if err != nil {
		return args.Get(0).(*pb.JoinGameResponse), fmt.Errorf("mock error: %w", err)
	}

	return args.Get(0).(*pb.JoinGameResponse), nil
}

func TestGatewayHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	mockClient := new(mockSnakeGameClient)
	h := &gatewayHandler{grpcClient: mockClient}

	s := httptest.NewServer(http.HandlerFunc(h.handleWS))
	defer s.Close()

	wsURL := "ws" + s.URL[4:] + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c, resp, err := websocket.Dial(ctx, wsURL, nil)
	assert.NoError(t, err)

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	defer func() {
		_ = c.Close(websocket.StatusNormalClosure, "")
	}()

	// Test GetTopPlayers
	topResp := &pb.GetTopPlayersResponse{
		TopPlayers: []*pb.PlayerScore{{PlayerName: "test", Score: 100}},
	}
	mockClient.On("GetTopPlayers", mock.Anything, mock.Anything, mock.Anything).Return(topResp, nil)

	msg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_Top{Top: &pb.GetTopPlayersRequest{}},
	}

	data, err := proto.Marshal(msg)
	assert.NoError(t, err)

	err = c.Write(ctx, websocket.MessageBinary, data)
	assert.NoError(t, err)

	_, respData, err := c.Read(ctx)
	assert.NoError(t, err)

	var res pb.ServerMessage
	err = proto.Unmarshal(respData, &res)
	assert.NoError(t, err)
	assert.IsType(t, &pb.ServerMessage_Top{}, res.Payload)
	assert.Equal(t, "test", res.Payload.(*pb.ServerMessage_Top).Top.TopPlayers[0].PlayerName)

	// Test SendDirection
	mockClient.On("SendDirection", mock.Anything, mock.Anything, mock.Anything).Return(&pb.SendDirectionResponse{}, nil)

	msg = &pb.ClientMessage{
		Payload: &pb.ClientMessage_Direction{
			Direction: &pb.SendDirectionRequest{
				PlayerName: "p1",
				Direction:  pb.Direction_DIRECTION_UP,
			},
		},
	}

	data, err = proto.Marshal(msg)
	assert.NoError(t, err)

	err = c.Write(ctx, websocket.MessageBinary, data)
	assert.NoError(t, err)

	// Allow some time for processing
	time.Sleep(100 * time.Millisecond)
	mockClient.AssertExpectations(t)
}

func TestGatewayHandler_ProxyJoin(t *testing.T) {
	t.Parallel()

	mockClient := new(mockSnakeGameClient)
	mockStream := new(mockJoinGameClient)

	h := &gatewayHandler{grpcClient: mockClient}

	mockClient.On("JoinGame", mock.Anything, mock.Anything, mock.Anything).Return(mockStream, nil)

	update := &pb.JoinGameResponse{Players: []*pb.Player{{Name: "p1", Alive: true}}}
	mockStream.On("Recv").Return(update, nil).Once()
	mockStream.On("Recv").Return((*pb.JoinGameResponse)(nil), io.EOF)

	s := httptest.NewServer(http.HandlerFunc(h.handleWS))
	defer s.Close()

	wsURL := "ws" + s.URL[4:] + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c, resp, err := websocket.Dial(ctx, wsURL, nil)
	assert.NoError(t, err)

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	defer func() {
		_ = c.Close(websocket.StatusNormalClosure, "")
	}()

	msg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_Join{Join: &pb.JoinGameRequest{PlayerName: "p1"}},
	}

	data, err := proto.Marshal(msg)
	assert.NoError(t, err)

	err = c.Write(ctx, websocket.MessageBinary, data)
	assert.NoError(t, err)

	_, respData, err := c.Read(ctx)
	assert.NoError(t, err)

	var res pb.ServerMessage
	err = proto.Unmarshal(respData, &res)
	assert.NoError(t, err)
	assert.IsType(t, &pb.ServerMessage_Update{}, res.Payload)
	assert.Equal(t, "p1", res.Payload.(*pb.ServerMessage_Update).Update.Players[0].Name)
}
