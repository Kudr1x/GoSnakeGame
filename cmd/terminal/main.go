// Package main implements the native client for the Snake Game.
package main

import (
	"GoSnakeGame/internal/client"
	"GoSnakeGame/internal/config"
	"flag"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.DefaultClientConfig()
	cfg.ParseFlags(flag.CommandLine)
	flag.Parse()

	conn, err := grpc.NewClient(cfg.ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to %s: %v", cfg.ServerAddr, err)
		os.Exit(1)
	}

	transport := NewGRPCTransport(conn)

	defer func() {
		if err := transport.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	app := client.NewApp(cfg, transport)
	app.Run()
}
