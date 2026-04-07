package main

import (
	"GoSnakeGame/internal/client"
	"context"
	"fmt"

	pb "GoSnakeGame/api/proto/snake/v1"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) startGame(mode pb.GameMode) tea.Cmd {
	return func() tea.Msg {
		transport, err := client.NewGRPCTransport(context.Background(), m.cfg.ServerAddr)
		if err != nil {
			return gameUpdateMsg{err: fmt.Errorf("failed to connect: %w", err)}
		}

		gameClient := client.NewGameClient(transport)

		_, err = gameClient.CreateAndJoinRoom(m.playerName, mode)
		if err != nil {
			return gameUpdateMsg{err: err}
		}

		// Wait for first update
		state, err := gameClient.ReceiveState()
		if err != nil {
			return gameUpdateMsg{err: fmt.Errorf("failed to receive game state: %w", err)}
		}

		return gameUpdateMsg{
			state:      state,
			gameClient: gameClient,
		}
	}
}

func (m model) waitForGameUpdate() tea.Cmd {
	return func() tea.Msg {
		if m.gameClient == nil {
			return gameUpdateMsg{err: fmt.Errorf("game client is nil")}
		}

		state, err := m.gameClient.ReceiveState()
		if err != nil {
			return gameUpdateMsg{err: fmt.Errorf("stream error: %w", err)}
		}

		return gameUpdateMsg{state: state}
	}
}

func (m model) sendDirection(dir pb.Direction) tea.Cmd {
	return func() tea.Msg {
		if m.gameClient == nil {
			return nil
		}

		err := m.gameClient.SendDirection(dir)
		if err != nil {
			// Ignore direction send errors
			return nil
		}

		return nil
	}
}

func (m model) fetchTopPlayers() tea.Cmd {
	return func() tea.Msg {
		var transport client.Transport

		if m.transport == nil {
			var err error
			transport, err = client.NewGRPCTransport(context.Background(), m.cfg.ServerAddr)
			if err != nil {
				return topPlayersMsg{err: fmt.Errorf("failed to connect: %w", err)}
			}
		} else {
			transport = m.transport
		}

		ctx, cancel := context.WithTimeout(context.Background(), m.cfg.TopPlayersTimeout)
		defer cancel()

		resp, err := transport.GetTopPlayers(ctx, &pb.GetTopPlayersRequest{})
		if err != nil {
			return topPlayersMsg{err: fmt.Errorf("failed to get top players: %w", err)}
		}

		return topPlayersMsg{
			players:   resp.TopPlayers,
			transport: transport,
		}
	}
}
