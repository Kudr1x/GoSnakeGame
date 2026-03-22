package game

import (
	"GoSnakeGame/internal/config"
	"testing"

	pb "GoSnakeGame/api/proto/snake/v1"

	"github.com/stretchr/testify/assert"
)

func TestRoomManager_CreateRoom(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultServerConfig()
	rm := NewRoomManager(cfg)

	roomID, err := rm.CreateRoom(pb.GameMode_MODE_SOLO)
	assert.NoError(t, err)
	assert.NotEmpty(t, roomID)

	engine, ok := rm.GetRoom(roomID)
	assert.True(t, ok)
	assert.NotNil(t, engine)
	assert.Equal(t, pb.GameMode_MODE_SOLO, engine.mode)
}

func TestRoomManager_CreateRoom_Limit(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultServerConfig()
	rm := NewRoomManager(cfg)

	for i := 0; i < 100; i++ {
		_, err := rm.CreateRoom(pb.GameMode_MODE_SOLO)
		assert.NoError(t, err)
	}

	_, err := rm.CreateRoom(pb.GameMode_MODE_SOLO)
	assert.ErrorIs(t, err, ErrMaxRoomsReached)
}

func TestRoomManager_GetTopPlayers(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultServerConfig()
	rm := NewRoomManager(cfg)

	roomID1, _ := rm.CreateRoom(pb.GameMode_MODE_SOLO)
	engine1, _ := rm.GetRoom(roomID1)
	p1 := engine1.AddOrUpdatePlayer("p1")
	p1.SetBestScore(100)

	roomID2, _ := rm.CreateRoom(pb.GameMode_MODE_SOLO)
	engine2, _ := rm.GetRoom(roomID2)
	p2 := engine2.AddOrUpdatePlayer("p2")
	p2.SetBestScore(200)

	topPlayers := rm.GetTopPlayers()
	assert.Len(t, topPlayers, 2)
}
