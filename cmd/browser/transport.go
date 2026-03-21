//go:build js && wasm

package main

import (
	"context"
	"log"

	pb "GoSnakeGame/api/proto/snake/v1"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"
)

type WSTransport struct {
	conn          *websocket.Conn
	writeCh       chan []byte
	roomCreatedCh chan *pb.CreateRoomResponse
	topPlayersCh  chan *pb.GetTopPlayersResponse
	updateCh      chan *pb.JoinGameResponse
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewWSTransport(ctx context.Context, conn *websocket.Conn) *WSTransport {
	ctx, cancel := context.WithCancel(ctx)
	
	t := &WSTransport{
		conn:          conn,
		writeCh:       make(chan []byte, 10),
		roomCreatedCh: make(chan *pb.CreateRoomResponse, 1),
		topPlayersCh:  make(chan *pb.GetTopPlayersResponse, 1),
		updateCh:      make(chan *pb.JoinGameResponse, 10),
		ctx:           ctx,
		cancel:        cancel,
	}

	go t.writeLoop()
	go t.readLoop()

	return t
}

func (t *WSTransport) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	msg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_CreateRoom{
			CreateRoom: req,
		},
	}
	
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	select {
	case t.writeCh <- data:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case resp := <-t.roomCreatedCh:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (t *WSTransport) JoinGame(ctx context.Context, req *pb.JoinGameRequest) error {
	msg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_Join{
			Join: req,
		},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case t.writeCh <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *WSTransport) ReceiveState() (*pb.JoinGameResponse, error) {
	select {
	case state := <-t.updateCh:
		return state, nil
	case <-t.ctx.Done():
		return nil, t.ctx.Err()
	}
}

func (t *WSTransport) SendDirection(ctx context.Context, req *pb.SendDirectionRequest) error {
	msg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_Direction{
			Direction: req,
		},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case t.writeCh <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *WSTransport) GetTopPlayers(ctx context.Context, req *pb.GetTopPlayersRequest) (*pb.GetTopPlayersResponse, error) {
	msg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_Top{
			Top: req,
		},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	select {
	case t.writeCh <- data:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case resp := <-t.topPlayersCh:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (t *WSTransport) Close() error {
	t.cancel()
	return t.conn.Close(websocket.StatusNormalClosure, "")
}

func (t *WSTransport) writeLoop() {
	for {
		select {
		case <-t.ctx.Done():
			return
		case data, ok := <-t.writeCh:
			if !ok {
				return
			}
			if err := t.conn.Write(t.ctx, websocket.MessageBinary, data); err != nil {
				return
			}
		}
	}
}

func (t *WSTransport) readLoop() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			_, data, err := t.conn.Read(t.ctx)
			if err != nil {
				log.Printf("ws read error: %v", err)
				t.cancel()
				return
			}

			var msg pb.ServerMessage
			if err := proto.Unmarshal(data, &msg); err != nil {
				continue
			}

			switch payload := msg.Payload.(type) {
			case *pb.ServerMessage_RoomCreated:
				select {
				case t.roomCreatedCh <- payload.RoomCreated:
				default:
				}
			case *pb.ServerMessage_Update:
				select {
				case t.updateCh <- payload.Update:
				default:
				}
			case *pb.ServerMessage_Top:
				select {
				case t.topPlayersCh <- payload.Top:
				default:
				}
			}
		}
	}
}
