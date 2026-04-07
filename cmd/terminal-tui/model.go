package main

import (
	"GoSnakeGame/internal/client"
	"GoSnakeGame/internal/config"

	pb "GoSnakeGame/api/proto/snake/v1"

	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	nameInputView viewState = iota
	menuView
	gameView
	topPlayersView
)

type gameState struct {
	players []*pb.Player
	food    []*pb.Point
	roomID  string
	mode    pb.GameMode
	sysMsg  *pb.SystemMessage
}

type gameUpdateMsg struct {
	state      *pb.JoinGameResponse
	gameClient *client.GameClient
	err        error
}

type topPlayersMsg struct {
	players   []*pb.PlayerScore
	transport client.Transport
	err       error
}

type model struct {
	cfg         *config.ClientConfig
	view        viewState
	menuCursor  int
	menuChoices []string
	playerName  string
	nameInput   string
	game        *gameState
	gameClient  *client.GameClient
	transport   client.Transport
	err         error
	topPlayers  []*pb.PlayerScore
	lastSysMsg  string
	width       int
	height      int
}

func initialModel(cfg *config.ClientConfig, playerName string) model {
	view := menuView
	if playerName == "" {
		view = nameInputView
	}

	return model{
		cfg:         cfg,
		view:        view,
		playerName:  playerName,
		menuChoices: []string{"Play Solo", "Play 1v1", "Play FFA", "View Top Players", "Quit"},
		width:       cfg.Width,
		height:      cfg.Height,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case gameUpdateMsg:
		if msg.err != nil {
			m.err = msg.err
			m.view = menuView

			return m, nil
		}

		// Update game client from first message
		if msg.gameClient != nil {
			m.gameClient = msg.gameClient
			m.view = gameView
		}

		if msg.state != nil {
			m.game = &gameState{
				players: msg.state.Players,
				food:    msg.state.Food,
				roomID:  msg.state.RoomId,
				mode:    msg.state.Mode,
				sysMsg:  msg.state.SystemMessage,
			}

			if msg.state.SystemMessage != nil {
				m.lastSysMsg = msg.state.SystemMessage.Message
			}
		}

		return m, m.waitForGameUpdate()

	case topPlayersMsg:
		if msg.err != nil {
			m.err = msg.err
			m.view = menuView

			return m, nil
		}

		if msg.transport != nil {
			m.transport = msg.transport
		}

		m.topPlayers = msg.players
		m.view = topPlayersView

		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	switch m.view {
	case nameInputView:
		return m.renderNameInput()
	case menuView:
		return m.renderMenu()
	case gameView:
		return m.renderGame()
	case topPlayersView:
		return m.renderTopPlayers()
	}

	return ""
}

func (m *model) cleanup() {
	if m.gameClient != nil {
		_ = m.gameClient.Close()
		m.gameClient = nil
	}

	if m.transport != nil {
		_ = m.transport.Close()
		m.transport = nil
	}

	m.game = nil
}
