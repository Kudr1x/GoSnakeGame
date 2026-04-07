package main

import (
	pb "GoSnakeGame/api/proto/snake/v1"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.view {
	case nameInputView:
		return m.handleNameInputKeys(msg)
	case menuView:
		return m.handleMenuKeys(msg)
	case gameView:
		return m.handleGameKeys(msg)
	case topPlayersView:
		return m.handleTopPlayersKeys(msg)
	}

	return m, nil
}

func (m model) handleNameInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "enter":
		if len(m.nameInput) > 0 {
			m.playerName = m.nameInput
			m.view = menuView
		}

	case "backspace":
		if len(m.nameInput) > 0 {
			m.nameInput = m.nameInput[:len(m.nameInput)-1]
		}

	default:
		if len(msg.String()) == 1 && len(m.nameInput) < 20 {
			m.nameInput += msg.String()
		}
	}

	return m, nil
}

func (m model) handleMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}

	case "down", "j":
		if m.menuCursor < len(m.menuChoices)-1 {
			m.menuCursor++
		}

	case "enter", " ":
		selected := m.menuChoices[m.menuCursor]

		switch selected {
		case "Quit":
			return m, tea.Quit
		case "Play Solo":
			return m, m.startGame(pb.GameMode_MODE_SOLO)
		case "Play 1v1":
			return m, m.startGame(pb.GameMode_MODE_1V1)
		case "Play FFA":
			return m, m.startGame(pb.GameMode_MODE_FFA)
		case "View Top Players":
			return m, m.fetchTopPlayers()
		}
	}

	return m, nil
}

func (m model) handleGameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.cleanup()

		return m, tea.Quit

	case "esc":
		m.cleanup()
		m.view = menuView
		m.err = nil

		return m, nil

	case "up", "w":
		return m, m.sendDirection(pb.Direction_DIRECTION_UP)

	case "down", "s":
		return m, m.sendDirection(pb.Direction_DIRECTION_DOWN)

	case "left", "a":
		return m, m.sendDirection(pb.Direction_DIRECTION_LEFT)

	case "right", "d":
		return m, m.sendDirection(pb.Direction_DIRECTION_RIGHT)
	}

	return m, nil
}

func (m model) handleTopPlayersKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "enter", " ":
		m.view = menuView
		m.topPlayers = nil

		return m, nil
	}

	return m, nil
}
