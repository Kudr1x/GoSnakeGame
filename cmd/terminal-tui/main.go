// Package main implements an improved TUI client for the Snake Game.
package main

import (
	"GoSnakeGame/internal/client"
	"GoSnakeGame/internal/config"
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var playerName string

	cfg := config.DefaultClientConfig()
	cfg.ParseFlags(flag.CommandLine)
	flag.StringVar(&playerName, "name", "", "player name")
	flag.Parse()

	// Test connection
	transport, err := client.NewGRPCTransport(context.Background(), cfg.ServerAddr)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Failed to connect to %s: %v", cfg.ServerAddr, err)))
		os.Exit(1)
	}

	_ = transport.Close()

	p := tea.NewProgram(initialModel(cfg, playerName), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}
}
