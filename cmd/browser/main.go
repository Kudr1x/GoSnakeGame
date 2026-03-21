//go:build js && wasm

// Package main implements the web client for the Snake Game.
package main

import (
	"GoSnakeGame/internal/client"
	"GoSnakeGame/internal/config"
	"context"
	"log"
	"strings"
	"syscall/js"

	"github.com/coder/websocket"
)

func main() {
	cfg := config.DefaultClientConfig()
	
	addr := cfg.ServerAddr
	if addr == "localhost:50051" || addr == ":50051" {
		window := js.Global().Get("window")
		if !window.IsUndefined() {
			location := window.Get("location")
			host := location.Get("hostname").String()
			port := location.Get("port").String()
			if port == "" {
				addr = host
			} else {
				addr = host + ":" + port
			}
		}
	}

	if !strings.HasPrefix(addr, "ws://") && !strings.HasPrefix(addr, "wss://") {
		addr = "ws://" + addr + "/ws"
	}

	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, addr, nil)
	if err != nil {
		log.Printf("failed to connect to gateway %s: %v", addr, err)
		return
	}

	transport := NewWSTransport(ctx, conn)

	defer func() {
		if err := transport.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	app := client.NewApp(cfg, transport)
	app.Run()
}
