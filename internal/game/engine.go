// Package game implements the core snake game logic.
package game

import (
	"GoSnakeGame/internal/config"
	"math"
	"math/rand"
	"sync"
	"time"

	pb "GoSnakeGame/api/proto/snake/v1"
)

// Engine manages the game state and rules.
type Engine struct {
	cfg     *config.ServerConfig
	players map[string]*PlayerInfo
	food    []*pb.Point
	mu      sync.RWMutex
}

// PlayerInfo stores the internal state of a player.
type PlayerInfo struct {
	Name      string
	Body      []*pb.Point
	Alive     bool
	Dir       pb.Direction
	BestScore int
	SessionID int64
}

// NewEngine creates a new game engine.
func NewEngine(cfg *config.ServerConfig) *Engine {
	return &Engine{
		cfg:     cfg,
		players: make(map[string]*PlayerInfo),
		food:    []*pb.Point{{X: 5, Y: 5}},
	}
}

// AddOrUpdatePlayer adds a new player or resets an existing one.
func (e *Engine) AddOrUpdatePlayer(name string) *PlayerInfo {
	e.mu.Lock()
	defer e.mu.Unlock()

	p, exists := e.players[name]
	sessionID := time.Now().UnixNano()

	if !exists {
		p = &PlayerInfo{
			Name:      name,
			Body:      []*pb.Point{{X: 10, Y: 10}},
			Alive:     true,
			Dir:       pb.Direction_DIRECTION_UP,
			SessionID: sessionID,
		}
		e.players[name] = p
	} else {
		p.Alive = true
		p.Dir = pb.Direction_DIRECTION_UP
		p.Body = []*pb.Point{{X: 10, Y: 10}}
		p.SessionID = sessionID
	}

	return p
}

// RemovePlayer removes a player from the game.
func (e *Engine) RemovePlayer(name string, sessionID int64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if p, ok := e.players[name]; ok && p.SessionID == sessionID {
		delete(e.players, name)
	}
}

// SetDirection updates the direction for a player if it's a valid change.
func (e *Engine) SetDirection(name string, dir pb.Direction) {
	e.mu.Lock()
	defer e.mu.Unlock()

	p, ok := e.players[name]
	if ok && p.Alive {
		if e.isValidDirectionChange(p.Dir, dir) {
			p.Dir = dir
		}
	}
}

func (e *Engine) isValidDirectionChange(current, next pb.Direction) bool {
	//nolint:exhaustive // Only movement related directions are handled
	switch next {
	case pb.Direction_DIRECTION_UP:
		return current != pb.Direction_DIRECTION_DOWN
	case pb.Direction_DIRECTION_DOWN:
		return current != pb.Direction_DIRECTION_UP
	case pb.Direction_DIRECTION_LEFT:
		return current != pb.Direction_DIRECTION_RIGHT
	case pb.Direction_DIRECTION_RIGHT:
		return current != pb.Direction_DIRECTION_LEFT
	default:
		return false
	}
}

// GetSnapshot returns a snapshot of the current game state.
func (e *Engine) GetSnapshot() *pb.JoinGameResponse {
	e.mu.RLock()
	defer e.mu.RUnlock()

	state := &pb.JoinGameResponse{
		Players: make([]*pb.Player, 0, len(e.players)),
		Food:    make([]*pb.Point, len(e.food)),
	}

	copy(state.Food, e.food)

	for _, p := range e.players {
		bodyCopy := make([]*pb.Point, len(p.Body))

		for i, pt := range p.Body {
			bodyCopy[i] = &pb.Point{X: pt.X, Y: pt.Y}
		}

		state.Players = append(state.Players, &pb.Player{
			Name:  p.Name,
			Body:  bodyCopy,
			Alive: p.Alive,
		})
	}

	return state
}

// GetTopPlayers returns the top scores.
func (e *Engine) GetTopPlayers() []*pb.PlayerScore {
	e.mu.RLock()
	defer e.mu.RUnlock()

	scores := make([]*pb.PlayerScore, 0, len(e.players))

	for _, p := range e.players {
		scores = append(scores, &pb.PlayerScore{
			PlayerName: p.Name,
			Score:      int32(p.BestScore), // #nosec G115
		})
	}

	return scores
}

// Run starts the game loop.
func (e *Engine) Run(onPlayerDie func(name string)) {
	ticker := time.NewTicker(e.cfg.UpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		e.update(onPlayerDie)
	}
}

func (e *Engine) update(onPlayerDie func(name string)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, p := range e.players {
		if !p.Alive {
			continue
		}

		p.BestScore = int(math.Max(float64(p.BestScore), float64(len(p.Body)*e.cfg.ScoreMultiplier)))

		newHead := e.getNewHead(p)

		if newHead == nil || e.checkCollision(newHead, p) {
			p.Alive = false

			if onPlayerDie != nil {
				onPlayerDie(p.Name)
			}

			continue
		}

		p.Body = append([]*pb.Point{newHead}, p.Body...)

		if e.checkFood(newHead) {
			e.generateFood()
		} else {
			p.Body = p.Body[:len(p.Body)-1]
		}
	}
}

func (e *Engine) getNewHead(p *PlayerInfo) *pb.Point {
	if len(p.Body) == 0 {
		return nil
	}

	head := p.Body[0]
	newHead := &pb.Point{X: head.X, Y: head.Y}

	//nolint:exhaustive // Only movement related directions are handled
	switch p.Dir {
	case pb.Direction_DIRECTION_UP:
		newHead.Y--
	case pb.Direction_DIRECTION_DOWN:
		newHead.Y++
	case pb.Direction_DIRECTION_LEFT:
		newHead.X--
	case pb.Direction_DIRECTION_RIGHT:
		newHead.X++
	}

	if e.isOutOfBounds(newHead) {
		return nil
	}

	return newHead
}

func (e *Engine) isOutOfBounds(pt *pb.Point) bool {
	// #nosec G115 - conversion is safe as width/height are reasonable
	return pt.X < 0 || pt.X >= int32(e.cfg.Width) || pt.Y < 0 || pt.Y >= int32(e.cfg.Height)
}

func (e *Engine) checkCollision(newHead *pb.Point, p *PlayerInfo) bool {
	for _, bodyPart := range p.Body {
		if newHead.X == bodyPart.X && newHead.Y == bodyPart.Y {
			return true
		}
	}

	return e.checkOtherPlayersCollision(newHead, p.Name)
}

func (e *Engine) checkOtherPlayersCollision(newHead *pb.Point, playerName string) bool {
	for _, otherPlayer := range e.players {
		if otherPlayer.Name != playerName && otherPlayer.Alive {
			for _, bodyPart := range otherPlayer.Body {
				if newHead.X == bodyPart.X && newHead.Y == bodyPart.Y {
					return true
				}
			}
		}
	}

	return false
}

func (e *Engine) checkFood(head *pb.Point) bool {
	for i, f := range e.food {
		if head.X == f.X && head.Y == f.Y {
			e.food = append(e.food[:i], e.food[i+1:]...)

			return true
		}
	}

	return false
}

func (e *Engine) generateFood() {
	for attempts := 0; attempts < e.cfg.MaxFoodAttempts; attempts++ {
		// #nosec G404 - weak random is ok for game food position
		// #nosec G115 - conversion of width/height is safe
		f := &pb.Point{
			X: int32(rand.Intn(e.cfg.Width)),
			Y: int32(rand.Intn(e.cfg.Height)),
		}

		if !e.isCellOccupied(f) {
			e.food = append(e.food, f)

			return
		}
	}
}

func (e *Engine) isCellOccupied(pt *pb.Point) bool {
	for _, p := range e.players {
		if !p.Alive {
			continue
		}

		for _, bodyPart := range p.Body {
			if pt.X == bodyPart.X && pt.Y == bodyPart.Y {
				return true
			}
		}
	}

	return false
}
