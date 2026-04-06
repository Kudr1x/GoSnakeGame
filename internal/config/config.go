// Package config provides game configuration.
package config

import (
	"flag"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// ServerConfig defines the server-side configuration.
type ServerConfig struct {
	Addr            string        `envconfig:"ADDR" default:":50051"`
	WebAddr         string        `envconfig:"WEB_ADDR" default:":8080"`
	Width           int           `envconfig:"WIDTH" default:"20"`
	Height          int           `envconfig:"HEIGHT" default:"20"`
	UpdateInterval  time.Duration `envconfig:"UPDATE_INTERVAL" default:"150ms"`
	SendInterval    time.Duration `envconfig:"SEND_INTERVAL" default:"100ms"`
	DeathWaitTime   time.Duration `envconfig:"DEATH_WAIT_TIME" default:"200ms"`
	MaxFoodAttempts int           `envconfig:"MAX_FOOD_ATTEMPTS" default:"100"`
	TopPlayersLimit int           `envconfig:"TOP_PLAYERS_LIMIT" default:"10"`
	ScoreMultiplier int           `envconfig:"SCORE_MULTIPLIER" default:"10"`
	MetricsAddr     string        `envconfig:"METRICS_ADDR" default:":9090"`
}

// ClientConfig defines the client-side configuration.
type ClientConfig struct {
	ServerAddr         string        `envconfig:"SERVER_ADDR" default:"localhost:50051"`
	Width              int           `envconfig:"WIDTH" default:"20"`
	Height             int           `envconfig:"HEIGHT" default:"20"`
	CellSize           int           `envconfig:"CELL_SIZE" default:"20"`
	SidebarWidth       int           `envconfig:"SIDEBAR_WIDTH" default:"150"`
	TopUpdateInterval  time.Duration `envconfig:"TOP_UPDATE_INTERVAL" default:"5s"`
	RenderInterval     time.Duration `envconfig:"RENDER_INTERVAL" default:"50ms"`
	DirectionTimeout   time.Duration `envconfig:"DIRECTION_TIMEOUT" default:"100ms"`
	TopPlayersTimeout  time.Duration `envconfig:"TOP_PLAYERS_TIMEOUT" default:"2s"`
	ScoreMultiplier    int           `envconfig:"SCORE_MULTIPLIER" default:"10"`
	Margin             int           `envconfig:"MARGIN" default:"10"`
	SidebarScoreOffset int           `envconfig:"SIDEBAR_SCORE_OFFSET" default:"10"`
	SidebarTopOffset   int           `envconfig:"SIDEBAR_TOP_OFFSET" default:"50"`
}

// DefaultServerConfig returns the default server configuration.
func DefaultServerConfig() *ServerConfig {
	cfg := &ServerConfig{}

	// Load from environment variables with defaults
	_ = envconfig.Process("SNAKE", cfg)

	return cfg
}

// ParseFlags parses server flags.
func (c *ServerConfig) ParseFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Addr, "addr", c.Addr, "server address for raw gRPC")
	fs.StringVar(&c.WebAddr, "web-addr", c.WebAddr, "server address for gRPC-Web (HTTP)")
	fs.StringVar(&c.MetricsAddr, "metrics-addr", c.MetricsAddr, "metrics server address")
	fs.IntVar(&c.Width, "width", c.Width, "game width")
	fs.IntVar(&c.Height, "height", c.Height, "game height")
}

// DefaultClientConfig returns the default client configuration.
func DefaultClientConfig() *ClientConfig {
	cfg := &ClientConfig{}

	// Load from environment variables with defaults
	_ = envconfig.Process("SNAKE", cfg)

	return cfg
}

// ParseFlags parses client flags.
func (c *ClientConfig) ParseFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.ServerAddr, "server", c.ServerAddr, "server address")
}
