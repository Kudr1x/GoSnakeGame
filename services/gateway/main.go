// Package main provides a WebSocket gateway for the Snake Game.
package main

import (
	pb "GoSnakeGame/api/proto/snake/v1"
	"GoSnakeGame/internal/logger"
	"GoSnakeGame/internal/metrics"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

type gatewayHandler struct {
	grpcClient pb.SnakeGameServiceClient
}

func main() {
	if err := logger.Init(false); err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	defer func() {
		_ = logger.Sync()
	}()

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Get().Fatal("failed to connect to engine", zap.Error(err))
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logger.Get().Error("failed to close grpc connection", zap.Error(err))
		}
	}()

	h := &gatewayHandler{
		grpcClient: pb.NewSnakeGameServiceClient(conn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWS)
	mux.Handle("/metrics", promhttp.Handler())

	mux.Handle("/", http.FileServer(http.Dir("web")))

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	logger.Get().Info("gateway listening on :8080")

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Get().Fatal("failed to start gateway", zap.Error(err))
	}
}

const (
	writeBufferSize = 10
)

func (h *gatewayHandler) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
		OriginPatterns:     []string{"*"},
	})
	if err != nil {
		logger.Get().Error("failed to accept websocket", zap.Error(err))

		return
	}

	metrics.WebSocketConnections.Inc()

	defer func() {
		metrics.WebSocketConnections.Dec()

		_ = c.Close(websocket.StatusInternalError, "internal error")
	}()

	ctx := r.Context()
	writeCh := make(chan []byte, writeBufferSize)

	go h.writeLoop(ctx, c, writeCh)

	for {
		mt, data, err := c.Read(ctx)
		if err != nil {
			return
		}

		if mt != websocket.MessageBinary {
			continue
		}

		var msg pb.ClientMessage
		if err := proto.Unmarshal(data, &msg); err != nil {
			continue
		}

		h.handleClientMessage(ctx, writeCh, &msg)
	}
}

func (h *gatewayHandler) writeLoop(ctx context.Context, c *websocket.Conn, writeCh <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-writeCh:
			if !ok {
				return
			}

			if err := c.Write(ctx, websocket.MessageBinary, data); err != nil {
				return
			}
		}
	}
}

func (h *gatewayHandler) handleClientMessage(ctx context.Context, writeCh chan<- []byte, msg *pb.ClientMessage) {
	switch payload := msg.Payload.(type) {
	case *pb.ClientMessage_Join:
		metrics.MessagesReceived.WithLabelValues("join").Inc()

		go h.proxyJoin(ctx, writeCh, payload.Join)
	case *pb.ClientMessage_Direction:
		metrics.MessagesReceived.WithLabelValues("direction").Inc()

		_, _ = h.grpcClient.SendDirection(ctx, payload.Direction)
	case *pb.ClientMessage_Top:
		metrics.MessagesReceived.WithLabelValues("top").Inc()

		go h.proxyTop(ctx, writeCh, payload.Top)
	case *pb.ClientMessage_CreateRoom:
		metrics.MessagesReceived.WithLabelValues("create_room").Inc()

		go h.proxyCreateRoom(ctx, writeCh, payload.CreateRoom)
	}
}

func (h *gatewayHandler) proxyCreateRoom(ctx context.Context, writeCh chan<- []byte, req *pb.CreateRoomRequest) {
	res, err := h.grpcClient.CreateRoom(ctx, req)
	if err != nil {
		return
	}

	out := &pb.ServerMessage{
		Payload: &pb.ServerMessage_RoomCreated{RoomCreated: res},
	}

	data, err := proto.Marshal(out)
	if err != nil {
		return
	}

	select {
	case writeCh <- data:
	case <-ctx.Done():
	}
}

func (h *gatewayHandler) proxyJoin(ctx context.Context, writeCh chan<- []byte, req *pb.JoinGameRequest) {
	stream, err := h.grpcClient.JoinGame(ctx, req)
	if err != nil {
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			return
		}

		out := &pb.ServerMessage{
			Payload: &pb.ServerMessage_Update{Update: resp},
		}

		data, err := proto.Marshal(out)
		if err != nil {
			continue
		}

		select {
		case writeCh <- data:
		case <-ctx.Done():
			return
		}
	}
}

func (h *gatewayHandler) proxyTop(ctx context.Context, writeCh chan<- []byte, req *pb.GetTopPlayersRequest) {
	res, err := h.grpcClient.GetTopPlayers(ctx, req)
	if err != nil {
		return
	}

	out := &pb.ServerMessage{
		Payload: &pb.ServerMessage_Top{Top: res},
	}

	data, err := proto.Marshal(out)
	if err != nil {
		return
	}

	select {
	case writeCh <- data:
	case <-ctx.Done():
	}
}
