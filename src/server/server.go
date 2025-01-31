package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "GoSnakeGame/proto"
	"google.golang.org/grpc"
)

const (
	width  = 20
	height = 20
)

type gameServer struct {
	pb.UnimplementedSnakeGameServer
	players map[string]*Player
	food    []*pb.Point
	mu      sync.Mutex
}

type Player struct {
	name  string
	body  []*pb.Point
	alive bool
	dir   pb.Direction
}

func (s *gameServer) JoinGame(req *pb.JoinRequest, stream pb.SnakeGame_JoinGameServer) error {
	s.mu.Lock()
	player := &Player{
		name:  req.PlayerName,
		body:  []*pb.Point{{X: 10, Y: 10}},
		alive: true,
		dir:   pb.Direction_UP,
	}
	s.players[req.PlayerName] = player
	s.mu.Unlock()

	for {
		s.mu.Lock()
		state := &pb.GameState{
			Players: make([]*pb.Player, 0, len(s.players)),
			Food:    s.food,
		}
		for _, p := range s.players {
			state.Players = append(state.Players, &pb.Player{
				Name:  p.name,
				Body:  p.body,
				Alive: p.alive,
			})
		}
		s.mu.Unlock()

		if err := stream.Send(state); err != nil {
			return err
		}

		s.mu.Lock()
		if !player.alive {
			s.mu.Unlock()
			return nil
		}
		s.mu.Unlock()

		time.Sleep(50 * time.Millisecond)
	}
}

func (s *gameServer) SendDirection(ctx context.Context, req *pb.DirectionRequest) (*pb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if player, ok := s.players[req.PlayerName]; ok && player.alive {
		player.dir = req.Direction
	}
	return &pb.Empty{}, nil
}

func (s *gameServer) updateGame() {
	for {
		s.mu.Lock()
		for name, player := range s.players {
			if !player.alive {
				delete(s.players, name)
				log.Printf("Игрок %s удален из игры", player.name)
				continue
			}

			newHead := s.getNewHead(player)
			if newHead == nil || s.checkCollision(newHead, player) {
				player.alive = false
				log.Printf("Игрок %s умер", player.name)
				continue
			}

			player.body = append([]*pb.Point{newHead}, player.body...)
			if s.checkFood(newHead) {
				s.generateFood()
			} else {
				player.body = player.body[:len(player.body)-1]
			}
		}
		s.mu.Unlock()

		time.Sleep(100 * time.Millisecond)
	}
}

func (s *gameServer) getNewHead(player *Player) *pb.Point {
	head := player.body[0]
	newHead := &pb.Point{X: head.X, Y: head.Y}

	switch player.dir {
	case pb.Direction_UP:
		newHead.Y--
	case pb.Direction_DOWN:
		newHead.Y++
	case pb.Direction_LEFT:
		newHead.X--
	case pb.Direction_RIGHT:
		newHead.X++
	}

	if newHead.X < 0 || newHead.X >= width || newHead.Y < 0 || newHead.Y >= height {
		return nil
	}

	return newHead
}

func (s *gameServer) checkCollision(newHead *pb.Point, player *Player) bool {
	for _, bodyPart := range player.body {
		if newHead.X == bodyPart.X && newHead.Y == bodyPart.Y {
			return true
		}
	}

	for _, otherPlayer := range s.players {
		if otherPlayer.name != player.name && otherPlayer.alive {
			for _, bodyPart := range otherPlayer.body {
				if newHead.X == bodyPart.X && newHead.Y == bodyPart.Y {
					return true
				}
			}
		}
	}

	return false
}

func (s *gameServer) checkFood(head *pb.Point) bool {
	for i, food := range s.food {
		if head.X == food.X && head.Y == food.Y {
			s.food = append(s.food[:i], s.food[i+1:]...)
			return true
		}
	}
	return false
}

func (s *gameServer) generateFood() {
	for {
		food := &pb.Point{
			X: int32(rand.Intn(width)),
			Y: int32(rand.Intn(height)),
		}

		collision := false
		for _, player := range s.players {
			for _, bodyPart := range player.body {
				if food.X == bodyPart.X && food.Y == bodyPart.Y {
					collision = true
					break
				}
			}
			if collision {
				break
			}
		}

		if !collision {
			s.food = append(s.food, food)
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}

	server := &gameServer{
		players: make(map[string]*Player),
		food:    []*pb.Point{{X: 5, Y: 5}},
	}

	go server.updateGame()

	s := grpc.NewServer()
	pb.RegisterSnakeGameServer(s, server)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("Остановка сервера...")
		s.GracefulStop()
	}()

	log.Println("Сервер запущен на порту 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
