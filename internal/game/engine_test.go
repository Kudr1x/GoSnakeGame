package game

import (
	"GoSnakeGame/internal/config"
	"testing"

	pb "GoSnakeGame/api/proto/snake/v1"

	"github.com/stretchr/testify/assert"
)

func TestEngine_NewEngine(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultServerConfig()
	e := NewEngine(cfg, "test_room", pb.GameMode_MODE_FFA)

	assert.NotNil(t, e)
	assert.NotNil(t, e.players)
	assert.NotNil(t, e.food)
	assert.Empty(t, e.players)
	assert.NotEmpty(t, e.food)
	assert.Equal(t, cfg, e.cfg)
}

func TestEngine_AddOrUpdatePlayer(t *testing.T) {
	t.Parallel()

	e := NewEngine(config.DefaultServerConfig(), "test_room", pb.GameMode_MODE_FFA)
	p := e.AddOrUpdatePlayer("player1")

	assert.NotNil(t, p)
	assert.Len(t, e.players, 1)
	assert.Equal(t, "player1", p.Name)
	assert.Contains(t, e.players, p.Name)

	p2 := e.AddOrUpdatePlayer("player1")
	assert.Equal(t, p, p2)
	assert.Len(t, e.players, 1)
}

func TestEngine_RemovePlayer(t *testing.T) {
	t.Parallel()

	e := NewEngine(config.DefaultServerConfig(), "test_room", pb.GameMode_MODE_FFA)
	p := e.AddOrUpdatePlayer("player1")

	assert.Len(t, e.players, 1)

	e.RemovePlayer(p.Name, p.GetSessionID())

	assert.Empty(t, e.players)
}

func TestEngine_SetDirection(t *testing.T) {
	t.Parallel()

	e := NewEngine(config.DefaultServerConfig(), "test_room", pb.GameMode_MODE_FFA)
	p := e.AddOrUpdatePlayer("player1")

	assert.Equal(t, pb.Direction_DIRECTION_RIGHT, p.GetDirection())

	e.SetDirection(p.Name, pb.Direction_DIRECTION_LEFT)
	assert.Equal(t, pb.Direction_DIRECTION_RIGHT, e.players[p.Name].GetDirection())

	e.SetDirection(p.Name, pb.Direction_DIRECTION_UP)
	assert.Equal(t, pb.Direction_DIRECTION_UP, e.players[p.Name].GetDirection())

	e.SetDirection(p.Name, pb.Direction_DIRECTION_DOWN)
	assert.Equal(t, pb.Direction_DIRECTION_UP, e.players[p.Name].GetDirection())

	e.SetDirection(p.Name, pb.Direction_DIRECTION_LEFT)
	assert.Equal(t, pb.Direction_DIRECTION_LEFT, e.players[p.Name].GetDirection())
}

func TestEngine_Update_PlayerMoves(t *testing.T) {
	t.Parallel()

	e := NewEngine(config.DefaultServerConfig(), "test_room", pb.GameMode_MODE_FFA)

	p := e.AddOrUpdatePlayer("player1")
	p.SetBody([]*pb.Point{{X: 10, Y: 10}})
	p.SetDirection(pb.Direction_DIRECTION_UP)

	e.update(nil)

	assert.True(t, p.IsAlive())
	body := p.GetBody()
	assert.Len(t, body, 1)
	assert.Equal(t, int32(10), body[0].X)
	assert.Equal(t, int32(9), body[0].Y)
}

func TestEngine_Update_PlayerEatsFood(t *testing.T) {
	t.Parallel()

	e := NewEngine(config.DefaultServerConfig(), "test_room", pb.GameMode_MODE_FFA)
	p := e.AddOrUpdatePlayer("player1")
	p.SetBody([]*pb.Point{{X: 5, Y: 6}})
	p.SetDirection(pb.Direction_DIRECTION_UP)

	e.food = []*pb.Point{{X: 5, Y: 5}}
	foodCountBefore := len(e.food)

	e.update(nil)

	assert.True(t, p.IsAlive())
	body := p.GetBody()
	assert.Len(t, body, 2)
	assert.Equal(t, int32(5), body[0].X)
	assert.Equal(t, int32(5), body[0].Y)
	assert.Len(t, e.food, foodCountBefore)
	assert.NotEqual(t, int32(5), e.food[0].X)
	assert.NotEqual(t, int32(5), e.food[0].Y)
}

func TestEngine_Update_PlayerHitsWall(t *testing.T) {
	t.Parallel()

	e := NewEngine(&config.ServerConfig{Width: 20, Height: 20}, "test_room", pb.GameMode_MODE_FFA)
	p := e.AddOrUpdatePlayer("player1")
	p.SetBody([]*pb.Point{{X: 10, Y: 0}})
	p.SetDirection(pb.Direction_DIRECTION_UP)

	var deadPlayerName string

	e.update(func(name string) {
		deadPlayerName = name
	})

	assert.False(t, p.IsAlive())
	assert.Equal(t, "player1", deadPlayerName)
}

func TestEngine_Update_PlayerHitsSelf(t *testing.T) {
	t.Parallel()

	e := NewEngine(config.DefaultServerConfig(), "test_room", pb.GameMode_MODE_FFA)
	p := e.AddOrUpdatePlayer("player1")
	p.SetBody([]*pb.Point{{X: 10, Y: 10}, {X: 10, Y: 11}, {X: 11, Y: 11}, {X: 11, Y: 10}})
	p.SetDirection(pb.Direction_DIRECTION_DOWN)

	var deadPlayerName string

	e.update(func(name string) {
		deadPlayerName = name
	})

	assert.False(t, p.IsAlive())
	assert.Equal(t, "player1", deadPlayerName)
}

// This test is deterministic now.
func TestEngine_Update_PlayerHitsOtherPlayer(t *testing.T) {
	t.Parallel()

	e := NewEngine(config.DefaultServerConfig(), "test_room", pb.GameMode_MODE_FFA)

	p1 := e.AddOrUpdatePlayer("player1")
	p2 := e.AddOrUpdatePlayer("player2")

	p1.SetBody([]*pb.Point{{X: 10, Y: 10}})
	p1.SetDirection(pb.Direction_DIRECTION_UP)

	// p2 is long enough so that even if p2 moves first, its body still covers {10, 9}
	p2.SetBody([]*pb.Point{{X: 10, Y: 9}, {X: 11, Y: 9}, {X: 12, Y: 9}})
	p2.SetDirection(pb.Direction_DIRECTION_LEFT)

	var deadPlayerName string

	e.update(func(name string) {
		if deadPlayerName == "" {
			deadPlayerName = name
		}
	})

	assert.False(t, p1.IsAlive())
	assert.True(t, p2.IsAlive())
	assert.Equal(t, "player1", deadPlayerName)
}
