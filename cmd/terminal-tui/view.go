package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) renderNameInput() string {
	s := titleStyle.Render("Snake Game - Terminal Client") + "\n\n"
	s += menuStyle.Render("Enter your name:") + "\n\n"
	s += selectedStyle.Render("> "+m.nameInput+"_") + "\n\n"
	s += lipgloss.NewStyle().Faint(true).Render("Press Enter to continue, Ctrl+C to quit")

	return s
}

func (m model) renderMenu() string {
	s := titleStyle.Render(fmt.Sprintf("Snake Game - Welcome, %s!", m.playerName)) + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n"
	}

	for i, choice := range m.menuChoices {
		if m.menuCursor == i {
			s += selectedStyle.Render("> "+choice) + "\n"
		} else {
			s += menuStyle.Render("  "+choice) + "\n"
		}
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Use ↑/↓ or j/k to navigate, Enter to select, q to quit")

	return s
}

func (m model) renderGame() string {
	if m.game == nil {
		return "Loading game..."
	}

	var sb strings.Builder

	sb.WriteString(titleStyle.Render(fmt.Sprintf("Room: %s | Mode: %s", m.game.roomID, m.game.mode.String())) + "\n\n")

	if m.lastSysMsg != "" {
		sb.WriteString(systemMsgStyle.Render(m.lastSysMsg) + "\n\n")
	}

	// Render game board
	board := make([][]string, m.height)
	for i := range board {
		board[i] = make([]string, m.width)

		for j := range board[i] {
			board[i][j] = "·"
		}
	}

	// Place food
	for _, f := range m.game.food {
		// #nosec G115 - dimensions are safe for int32
		if f.Y >= 0 && f.Y < int32(m.height) && f.X >= 0 && f.X < int32(m.width) {
			board[f.Y][f.X] = foodStyle.Render("●")
		}
	}

	// Place players
	for i, p := range m.game.players {
		color := playerColors[i%len(playerColors)]
		style := lipgloss.NewStyle().Foreground(color)

		if !p.Alive {
			style = deadStyle
		}

		for j, pt := range p.Body {
			// #nosec G115 - dimensions are safe for int32
			if pt.Y >= 0 && pt.Y < int32(m.height) && pt.X >= 0 && pt.X < int32(m.width) {
				if j == 0 {
					board[pt.Y][pt.X] = style.Render("█")
				} else {
					board[pt.Y][pt.X] = style.Render("▓")
				}
			}
		}
	}

	// Draw board
	for _, row := range board {
		sb.WriteString(strings.Join(row, " ") + "\n")
	}

	sb.WriteString("\n")

	// Player info
	for i, p := range m.game.players {
		color := playerColors[i%len(playerColors)]
		style := lipgloss.NewStyle().Foreground(color)

		if !p.Alive {
			style = deadStyle
		}

		status := "Alive"
		if !p.Alive {
			status = "Dead"
		}

		const scoreMultiplier = 10
		sb.WriteString(style.Render(fmt.Sprintf("%s: %s | Score: %d", p.Name, status, len(p.Body)*scoreMultiplier)) + "\n")
	}

	sb.WriteString("\n" + lipgloss.NewStyle().Faint(true).Render("Use ↑↓←→ or WASD to move, ESC to return, q to quit"))

	return sb.String()
}

func (m model) renderTopPlayers() string {
	s := titleStyle.Render("Top Players") + "\n\n"

	if len(m.topPlayers) == 0 {
		s += menuStyle.Render("No players yet") + "\n"
	} else {
		for i, p := range m.topPlayers {
			s += menuStyle.Render(fmt.Sprintf("%d. %s - Score: %d | Rating: %d", i+1, p.PlayerName, p.Score, p.Rating)) + "\n"
		}
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Press ESC or Enter to return")

	return s
}
