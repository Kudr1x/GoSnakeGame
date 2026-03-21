// Package main provides a WebSocket gateway for the Snake Game.
package main

import (
	pb "GoSnakeGame/api/proto/snake/v1"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

type gatewayHandler struct {
	grpcClient pb.SnakeGameServiceClient
}

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to engine: %v", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close grpc connection: %v", err)
		}
	}()

	h := &gatewayHandler{
		grpcClient: pb.NewSnakeGameServiceClient(conn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWS)

	mux.Handle("/", http.FileServer(http.Dir("web")))

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Println("Gateway listening on :8080...")

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("failed to start gateway: %v", err)
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
		log.Printf("failed to accept websocket: %v", err)

		return
	}

	defer func() {
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
		go h.proxyJoin(ctx, writeCh, payload.Join)
	case *pb.ClientMessage_Direction:
		_, _ = h.grpcClient.SendDirection(ctx, payload.Direction)
	case *pb.ClientMessage_Top:
		go h.proxyTop(ctx, writeCh, payload.Top)
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
