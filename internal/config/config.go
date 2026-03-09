// Package config provides game configuration.
package config

import (
	"flag"
	"time"
)

// ServerConfig defines the server-side configuration.
type ServerConfig struct {
	Addr            string
	Width           int
	Height          int
	UpdateInterval  time.Duration
	SendInterval    time.Duration
	DeathWaitTime   time.Duration
	MaxFoodAttempts int
	TopPlayersLimit int
	ScoreMultiplier int
}

// ClientConfig defines the client-side configuration.
type ClientConfig struct {
	ServerAddr         string
	Width              int
	Height             int
	CellSize           int
	SidebarWidth       int
	TopUpdateInterval  time.Duration
	RenderInterval     time.Duration
	DirectionTimeout   time.Duration
	TopPlayersTimeout  time.Duration
	ScoreMultiplier    int
	Margin             int
	SidebarScoreOffset int
	SidebarTopOffset   int
}

// DefaultServerConfig returns the default server configuration.
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Addr:            ":50051",
		Width:           20,
		Height:          20,
		UpdateInterval:  150 * time.Millisecond,
		SendInterval:    100 * time.Millisecond,
		DeathWaitTime:   200 * time.Millisecond,
		MaxFoodAttempts: 100,
		TopPlayersLimit: 10,
		ScoreMultiplier: 10,
	}
}

// ParseFlags parses server flags.
func (c *ServerConfig) ParseFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Addr, "addr", c.Addr, "server address")
	fs.IntVar(&c.Width, "width", c.Width, "game width")
	fs.IntVar(&c.Height, "height", c.Height, "game height")
}

// DefaultClientConfig returns the default client configuration.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		ServerAddr:         "localhost:50051",
		Width:              20,
		Height:             20,
		CellSize:           20,
		SidebarWidth:       150,
		TopUpdateInterval:  5 * time.Second,
		RenderInterval:     50 * time.Millisecond,
		DirectionTimeout:   100 * time.Millisecond,
		TopPlayersTimeout:  2 * time.Second,
		ScoreMultiplier:    10,
		Margin:             10,
		SidebarScoreOffset: 10,
		SidebarTopOffset:   50,
	}
}

// ParseFlags parses client flags.
func (c *ClientConfig) ParseFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.ServerAddr, "server", c.ServerAddr, "server address")
}
