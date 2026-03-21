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
	roomID  string
	mode    pb.GameMode
	players map[string]*PlayerInfo
	food    []*pb.Point
	mu      sync.RWMutex
	started bool

	spawnPoints []spawnPoint
}

type spawnPoint struct {
	pos *pb.Point
	dir pb.Direction
}

// PlayerInfo stores the internal state of a player.
type PlayerInfo struct {
	mu        sync.RWMutex
	ID        int32
	Name      string
	Body      []*pb.Point
	Alive     bool
	Dir       pb.Direction
	BestScore int
	SessionID int64
}

// IsAlive returns the alive status of the player.
func (p *PlayerInfo) IsAlive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.Alive
}

// SetAlive sets the alive status of the player.
func (p *PlayerInfo) SetAlive(alive bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Alive = alive
}

// GetDirection returns the direction of the player.
func (p *PlayerInfo) GetDirection() pb.Direction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.Dir
}

// SetDirection sets the direction of the player.
func (p *PlayerInfo) SetDirection(dir pb.Direction) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Dir = dir
}

// GetBody returns a copy of the player's body.
func (p *PlayerInfo) GetBody() []*pb.Point {
	p.mu.RLock()
	defer p.mu.RUnlock()

	bodyCopy := make([]*pb.Point, len(p.Body))
	for i, pt := range p.Body {
		bodyCopy[i] = &pb.Point{X: pt.X, Y: pt.Y}
	}

	return bodyCopy
}

// SetBody sets the body of the player.
func (p *PlayerInfo) SetBody(body []*pb.Point) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Body = body
}

// GetBestScore returns the best score of the player.
func (p *PlayerInfo) GetBestScore() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.BestScore
}

// SetBestScore sets the best score of the player.
func (p *PlayerInfo) SetBestScore(score int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.BestScore = score
}

// GetSessionID returns the session ID of the player.
func (p *PlayerInfo) GetSessionID() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.SessionID
}

// SetSessionID sets the session ID of the player.
func (p *PlayerInfo) SetSessionID(id int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.SessionID = id
}

const (
	spawnOffset = 3
)

// NewEngine creates a new game engine.
func NewEngine(cfg *config.ServerConfig, roomID string, mode pb.GameMode) *Engine {
	return &Engine{
		cfg:     cfg,
		roomID:  roomID,
		mode:    mode,
		players: make(map[string]*PlayerInfo),
		food:    []*pb.Point{{X: 5, Y: 5}},
		spawnPoints: []spawnPoint{
			{pos: &pb.Point{X: 2, Y: 2}, dir: pb.Direction_DIRECTION_RIGHT},
			// #nosec G115 - Dimensions are safe for int32
			{
				pos: &pb.Point{X: int32(cfg.Width - spawnOffset), Y: int32(cfg.Height - spawnOffset)},
				dir: pb.Direction_DIRECTION_LEFT,
			},
			// #nosec G115 - Dimensions are safe for int32
			{
				pos: &pb.Point{X: int32(cfg.Width - spawnOffset), Y: 2},
				dir: pb.Direction_DIRECTION_DOWN,
			},
			// #nosec G115 - Dimensions are safe for int32
			{
				pos: &pb.Point{X: 2, Y: int32(cfg.Height - spawnOffset)},
				dir: pb.Direction_DIRECTION_UP,
			},
		},
	}
}

const (
	maxPlayersSolo = 1
	maxPlayers1v1  = 2
	maxPlayersFFA  = 4
)

// MaxPlayers returns the maximum number of players allowed in the current mode.
func (e *Engine) MaxPlayers() int {
	switch e.mode {
	case pb.GameMode_MODE_UNSPECIFIED:
		return maxPlayersFFA
	case pb.GameMode_MODE_SOLO:
		return maxPlayersSolo
	case pb.GameMode_MODE_1V1:
		return maxPlayers1v1
	case pb.GameMode_MODE_FFA:
		return maxPlayersFFA
	default:
		return maxPlayersFFA
	}
}

// AddOrUpdatePlayer adds a new player or resets an existing one. Returns nil if the room is full.
func (e *Engine) AddOrUpdatePlayer(name string) *PlayerInfo {
	e.mu.Lock()
	defer e.mu.Unlock()

	p, exists := e.players[name]

	if !exists && len(e.players) >= e.MaxPlayers() {
		return nil
	}

	sessionID := time.Now().UnixNano()

	// Pick a spawn point based on the current number of players
	spawn := e.spawnPoints[len(e.players)%len(e.spawnPoints)]

	if !exists {
		p = &PlayerInfo{
			// #nosec G115 - Player count will not exceed int32 limits
			ID:        int32(len(e.players) + 1),
			Name:      name,
			Body:      []*pb.Point{{X: spawn.pos.X, Y: spawn.pos.Y}},
			Alive:     true,
			Dir:       spawn.dir,
			SessionID: sessionID,
		}
		e.players[name] = p

		if len(e.players) == e.MaxPlayers() {
			e.started = true
		}
	} else {
		p.SetAlive(true)
		p.SetDirection(spawn.dir)
		p.SetBody([]*pb.Point{{X: spawn.pos.X, Y: spawn.pos.Y}})
		p.SetSessionID(sessionID)
	}

	return p
}

// RemovePlayer removes a player from the game.
func (e *Engine) RemovePlayer(name string, sessionID int64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if p, ok := e.players[name]; ok && p.GetSessionID() == sessionID {
		delete(e.players, name)
	}
}

// SetDirection updates the direction for a player if it's a valid change.
func (e *Engine) SetDirection(name string, dir pb.Direction) {
	e.mu.Lock()
	defer e.mu.Unlock()

	p, ok := e.players[name]
	if ok && p.IsAlive() {
		if e.isValidDirectionChange(p.GetDirection(), dir) {
			p.SetDirection(dir)
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
		RoomId:  e.roomID,
		Mode:    e.mode,
		Players: make([]*pb.Player, 0, len(e.players)),
		Food:    make([]*pb.Point, len(e.food)),
	}

	copy(state.Food, e.food)

	for _, p := range e.players {
		state.Players = append(state.Players, &pb.Player{
			Id:        p.ID,
			Name:      p.Name,
			Body:      p.GetBody(),
			Alive:     p.IsAlive(),
			Direction: p.GetDirection(),
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
			Score:      int32(p.GetBestScore()), // #nosec G115 - score is unlikely to overflow int32
		})
	}

	return scores
}

// Run starts the game loop.
func (e *Engine) Run(onPlayerDie func(name string)) {
	ticker := time.NewTicker(e.cfg.UpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		e.mu.RLock()
		started := e.started
		e.mu.RUnlock()

		if started {
			e.update(onPlayerDie)
		}
	}
}

func (e *Engine) update(onPlayerDie func(name string)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, p := range e.players {
		if !p.IsAlive() {
			continue
		}

		p.SetBestScore(int(math.Max(float64(p.GetBestScore()), float64(len(p.GetBody())*e.cfg.ScoreMultiplier))))

		newHead := e.getNewHead(p)

		if newHead == nil || e.checkCollision(newHead, p) {
			p.SetAlive(false)

			if onPlayerDie != nil {
				onPlayerDie(p.Name)
			}

			continue
		}

		body := p.GetBody()
		p.SetBody(append([]*pb.Point{newHead}, body...))

		if e.checkFood(newHead) {
			e.generateFood()
		} else {
			body := p.GetBody()
			p.SetBody(body[:len(body)-1])
		}
	}
}

func (e *Engine) getNewHead(p *PlayerInfo) *pb.Point {
	body := p.GetBody()
	if len(body) == 0 {
		return nil
	}

	head := body[0]
	newHead := &pb.Point{X: head.X, Y: head.Y}

	//nolint:exhaustive // Only movement related directions are handled
	switch p.GetDirection() {
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
	for _, bodyPart := range p.GetBody() {
		if newHead.X == bodyPart.X && newHead.Y == bodyPart.Y {
			return true
		}
	}

	return e.checkOtherPlayersCollision(newHead, p.Name)
}

func (e *Engine) checkOtherPlayersCollision(newHead *pb.Point, playerName string) bool {
	for _, otherPlayer := range e.players {
		if otherPlayer.Name != playerName && otherPlayer.IsAlive() {
			for _, bodyPart := range otherPlayer.GetBody() {
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
		if !p.IsAlive() {
			continue
		}

		for _, bodyPart := range p.GetBody() {
			if pt.X == bodyPart.X && pt.Y == bodyPart.Y {
				return true
			}
		}
	}

	return false
}
