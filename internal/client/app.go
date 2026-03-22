// Package client provides the shared game application logic and UI.
package client

import (
	"GoSnakeGame/internal/config"
	"context"
	"fmt"
	"image/color"
	"log"
	"strings"
	"sync"
	"time"

	pb "GoSnakeGame/api/proto/snake/v1"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// App encapsulates the game client logic and UI.
type App struct {
	playerName   string
	roomID       string
	transport    Transport
	dirCh        chan pb.Direction
	currentState *pb.JoinGameResponse
	mu           sync.RWMutex
	currentDir   pb.Direction
	currentScore int
	mainWindow   fyne.Window
	stopGameCh   chan struct{}
	cancelStream context.CancelFunc
	cfg          *config.ClientConfig

	rectPool []*canvas.Rectangle
	active   []*canvas.Rectangle

	gameOver func(bool)
}

const maxNameLength = 12

// NewApp creates a new game application.
func NewApp(cfg *config.ClientConfig, transport Transport) *App {
	gc := &App{
		transport: transport,
		dirCh:     make(chan pb.Direction, 1),
		cfg:       cfg,
	}
	gc.gameOver = gc.showGameOverScreen

	return gc
}

// Run initializes the UI and starts the application.
func (gc *App) Run() {
	a := app.New()
	gc.mainWindow = a.NewWindow("Snake Game")
	gc.mainWindow.SetFixedSize(true)

	gc.checkInviteLink()
	gc.showJoinScreen()

	gc.mainWindow.Canvas().SetOnTypedKey(gc.handleKeyPress)

	w := float32(gc.cfg.Width*gc.cfg.CellSize + gc.cfg.SidebarWidth)
	h := float32(gc.cfg.Height * gc.cfg.CellSize)
	gc.mainWindow.Resize(fyne.NewSize(w, h))
	gc.mainWindow.ShowAndRun()
}

func (gc *App) showJoinScreen() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter your name (max 12 chars)")
	nameEntry.Text = gc.playerName

	modeSelect := widget.NewSelect([]string{"Solo", "1v1", "4-Player"}, nil)
	modeSelect.SetSelected("Solo")

	createButton := widget.NewButton("Create Room", func() {
		gc.playerName = strings.TrimSpace(nameEntry.Text)
		if gc.playerName == "" {
			dialog.ShowError(fmt.Errorf("name cannot be empty"), gc.mainWindow)

			return
		}

		if len(gc.playerName) > maxNameLength {
			dialog.ShowError(fmt.Errorf("name too long"), gc.mainWindow)

			return
		}

		var mode pb.GameMode

		switch modeSelect.Selected {
		case "Solo":
			mode = pb.GameMode_MODE_SOLO
		case "1v1":
			mode = pb.GameMode_MODE_1V1
		case "4-Player":
			mode = pb.GameMode_MODE_FFA
		}

		const createTimeout = 2 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), createTimeout)
		defer cancel()

		resp, err := gc.transport.CreateRoom(ctx, &pb.CreateRoomRequest{
			PlayerName: gc.playerName,
			Mode:       mode,
		})
		if err != nil {
			dialog.ShowError(err, gc.mainWindow)

			return
		}

		gc.roomID = resp.RoomId

		if mode != pb.GameMode_MODE_SOLO {
			d := dialog.NewInformation(
				"Room Created",
				fmt.Sprintf("Room ID: %s\nInvite Link: %s", resp.RoomId, resp.InviteLink),
				gc.mainWindow,
			)
			d.SetOnClosed(func() {
				gc.joinGame()
			})
			d.Show()
		} else {
			gc.joinGame()
		}
	})

	roomIDEntry := widget.NewEntry()
	roomIDEntry.SetPlaceHolder("Or enter Room ID to join")
	roomIDEntry.Text = gc.roomID

	joinButton := widget.NewButton("Join Room", func() {
		gc.playerName = strings.TrimSpace(nameEntry.Text)
		gc.roomID = strings.TrimSpace(roomIDEntry.Text)

		if gc.playerName == "" || gc.roomID == "" {
			dialog.ShowError(fmt.Errorf("name and room ID are required"), gc.mainWindow)

			return
		}

		gc.joinGame()
	})

	gc.mainWindow.SetContent(container.NewVBox(
		widget.NewLabel("Player Name:"),
		nameEntry,
		widget.NewSeparator(),
		widget.NewLabel("Create New Room:"),
		modeSelect,
		createButton,
		widget.NewSeparator(),
		widget.NewLabel("Join Existing Room:"),
		roomIDEntry,
		joinButton,
	))
}

func (gc *App) handleKeyPress(event *fyne.KeyEvent) {
	gc.mu.RLock()
	prevDir := gc.currentDir
	gc.mu.RUnlock()

	direction := GetDirectionFromKey(event.Name, prevDir)

	if direction != pb.Direction_DIRECTION_UNSPECIFIED {
		select {
		case gc.dirCh <- direction:
		default:
		}
	}
}

// GetDirectionFromKey maps Fyne keys to game directions.
func GetDirectionFromKey(key fyne.KeyName, prevDir pb.Direction) pb.Direction {
	//nolint:exhaustive // Only movement-related keys are handled
	switch key {
	case fyne.KeyUp:
		if prevDir != pb.Direction_DIRECTION_DOWN {
			return pb.Direction_DIRECTION_UP
		}
	case fyne.KeyDown:
		if prevDir != pb.Direction_DIRECTION_UP {
			return pb.Direction_DIRECTION_DOWN
		}
	case fyne.KeyLeft:
		if prevDir != pb.Direction_DIRECTION_RIGHT {
			return pb.Direction_DIRECTION_LEFT
		}
	case fyne.KeyRight:
		if prevDir != pb.Direction_DIRECTION_LEFT {
			return pb.Direction_DIRECTION_RIGHT
		}
	default:
		return pb.Direction_DIRECTION_UNSPECIFIED
	}

	return pb.Direction_DIRECTION_UNSPECIFIED
}

func (gc *App) joinGame() {
	if gc.cancelStream != nil {
		gc.cancelStream()
	}

	if gc.stopGameCh != nil {
		close(gc.stopGameCh)
	}

	gc.stopGameCh = make(chan struct{})

	gc.mu.Lock()
	gc.rectPool = nil
	gc.active = nil
	gc.mu.Unlock()

	var (
		err error
		ctx context.Context
	)

	ctx, gc.cancelStream = context.WithCancel(context.Background())

	err = gc.transport.JoinGame(ctx, &pb.JoinGameRequest{
		PlayerName: gc.playerName,
		RoomId:     gc.roomID,
	})
	if err != nil {
		log.Printf("failed to join game: %v", err)
		dialog.ShowError(fmt.Errorf("failed to join game: %w", err), gc.mainWindow)
		gc.showJoinScreen()

		return
	}

	gc.currentDir = pb.Direction_DIRECTION_UNSPECIFIED

	go gc.receiveGameState(gc.stopGameCh)
	go gc.sendDirection(gc.stopGameCh)

	gc.mainWindow.SetContent(gc.createGameUI(gc.stopGameCh))
}

func (gc *App) createGameUI(stopCh chan struct{}) fyne.CanvasObject {
	gameContainer := container.NewWithoutLayout()

	board := canvas.NewRectangle(color.RGBA{R: 63, G: 66, B: 75, A: 255})
	board.Resize(fyne.NewSize(float32(gc.cfg.CellSize*gc.cfg.Width), float32(gc.cfg.CellSize*gc.cfg.Height)))
	gameContainer.Add(board)

	scoreLabel := widget.NewLabel("Score: 0")
	scoreX := float32(gc.cfg.Width*gc.cfg.CellSize + gc.cfg.SidebarScoreOffset)
	scoreLabel.Move(fyne.NewPos(scoreX, float32(gc.cfg.Margin)))
	gameContainer.Add(scoreLabel)

	topPlayersLabel := widget.NewLabel("Loading top...")
	topX := float32(gc.cfg.Width*gc.cfg.CellSize + gc.cfg.SidebarScoreOffset)
	topPlayersLabel.Move(fyne.NewPos(topX, float32(gc.cfg.SidebarTopOffset)))
	gameContainer.Add(topPlayersLabel)

	go gc.updateTopPlayersLoop(stopCh, topPlayersLabel)
	go gc.renderLoop(stopCh, gameContainer, scoreLabel)

	return gameContainer
}

func (gc *App) updateTopPlayersLoop(stopCh chan struct{}, label *widget.Label) {
	ticker := time.NewTicker(gc.cfg.TopUpdateInterval)
	defer ticker.Stop()

	updateTop := func() {
		topText := gc.getStringTop()
		label.SetText(topText)
	}

	updateTop()

	for {
		select {
		case <-ticker.C:
			updateTop()
		case <-stopCh:
			return
		}
	}
}

func (gc *App) renderLoop(
	stopCh chan struct{},
	container *fyne.Container,
	scoreLabel *widget.Label,
) {
	ticker := time.NewTicker(gc.cfg.RenderInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			gc.mu.RLock()
			state := gc.currentState
			gc.mu.RUnlock()

			if state == nil {
				continue
			}

			gc.drawFrame(container, scoreLabel, state)
		}
	}
}

func (gc *App) drawFrame(
	container *fyne.Container,
	scoreLabel *widget.Label,
	state *pb.JoinGameResponse,
) {
	for _, r := range gc.active {
		r.Hide()
	}

	gc.rectPool = append(gc.rectPool, gc.active...)
	gc.active = gc.active[:0]

	foodColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for _, f := range state.Food {
		gc.drawRect(container, f.X, f.Y, foodColor)
	}

	for _, p := range state.Players {
		gc.drawPlayer(container, scoreLabel, p)
	}

	container.Refresh()
}

func (gc *App) drawPlayer(
	container *fyne.Container,
	scoreLabel *widget.Label,
	p *pb.Player,
) {
	if !p.Alive && p.Name != gc.playerName {
		return
	}

	var snakeColor color.RGBA
	if p.Name == gc.playerName {
		snakeColor = color.RGBA{R: 0, G: 255, B: 0, A: 255}
		gc.currentScore = len(p.Body) * gc.cfg.ScoreMultiplier
		scoreLabel.SetText(fmt.Sprintf("Score: %d", gc.currentScore))
	} else {
		//nolint:mnd // Magic numbers are used for color variety based on ID
		switch (p.Id - 1) % 4 {
		case 0:
			snakeColor = color.RGBA{R: 0, G: 191, B: 255, A: 255}
		case 1:
			snakeColor = color.RGBA{R: 255, G: 165, B: 0, A: 255}
		case 2:
			snakeColor = color.RGBA{R: 138, G: 43, B: 226, A: 255}
		default:
			snakeColor = color.RGBA{R: 255, G: 20, B: 147, A: 255}
		}
	}

	for _, bodyPart := range p.Body {
		gc.drawRect(container, bodyPart.X, bodyPart.Y, snakeColor)
	}
}

func (gc *App) drawRect(container *fyne.Container, x, y int32, fillColor color.Color) {
	var r *canvas.Rectangle

	if len(gc.rectPool) > 0 {
		r = gc.rectPool[len(gc.rectPool)-1]
		gc.rectPool = gc.rectPool[:len(gc.rectPool)-1]
	} else {
		r = canvas.NewRectangle(fillColor)
		r.Resize(fyne.NewSize(float32(gc.cfg.CellSize), float32(gc.cfg.CellSize)))
		container.Add(r)
	}

	r.FillColor = fillColor
	// #nosec G115 - conversion is safe
	posX := float32(x * int32(gc.cfg.CellSize))
	// #nosec G115 - conversion is safe
	posY := float32(y * int32(gc.cfg.CellSize))
	r.Move(fyne.NewPos(posX, posY))
	r.Show()

	gc.active = append(gc.active, r)
}

func (gc *App) getStringTop() string {
	ctx, cancel := context.WithTimeout(context.Background(), gc.cfg.TopPlayersTimeout)
	defer cancel()

	topPlayers, err := gc.transport.GetTopPlayers(ctx, &pb.GetTopPlayersRequest{})
	if err != nil {
		return "Top Players:\n(Error loading)"
	}

	playerList := make([]string, 0, len(topPlayers.TopPlayers)+1)
	playerList = append(playerList, "Top Players:")

	for _, p := range topPlayers.TopPlayers {
		playerList = append(playerList, fmt.Sprintf("%s: %d", p.PlayerName, p.Score))
	}

	return strings.Join(playerList, "\n")
}

func (gc *App) receiveGameState(stopCh chan struct{}) {
	var lastAliveCount int

	for {
		select {
		case <-stopCh:
			return
		default:
			state, err := gc.transport.ReceiveState()
			if err != nil {
				log.Printf("connection lost: %v", err)

				return
			}

			gc.updateState(state)

			aliveCount, myPlayer := gc.analyzePlayers(state)

			if myPlayer != nil && !myPlayer.Alive {
				// If I'm dead, check if I was the last one alive or if it was a solo game
				winner := (state.Mode != pb.GameMode_MODE_SOLO && lastAliveCount == 1 && aliveCount == 0)

				gc.gameOver(winner)

				return
			}

			lastAliveCount = aliveCount
		}
	}
}

func (gc *App) updateState(state *pb.JoinGameResponse) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.currentState = state
}

func (gc *App) analyzePlayers(state *pb.JoinGameResponse) (int, *pb.Player) {
	aliveCount := 0

	var myPlayer *pb.Player

	for _, p := range state.Players {
		if p.Alive {
			aliveCount++
		}

		if p.Name == gc.playerName {
			myPlayer = p
		}
	}

	return aliveCount, myPlayer
}

func (gc *App) sendDirection(stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case dir := <-gc.dirCh:
			ctx, cancel := context.WithTimeout(context.Background(), gc.cfg.DirectionTimeout)

			err := gc.transport.SendDirection(ctx, &pb.SendDirectionRequest{
				PlayerName: gc.playerName,
				RoomId:     gc.roomID,
				Direction:  dir,
			})

			cancel()

			if err != nil {
				log.Printf("failed to send direction: %v", err)
			} else {
				gc.mu.Lock()
				gc.currentDir = dir
				gc.mu.Unlock()
			}
		}
	}
}

func (gc *App) showGameOverScreen(winner bool) {
	title := "Game over"
	if winner {
		title = "YOU WIN!"
	}

	dialog.ShowCustomConfirm(
		title,
		"Retry",
		"Menu",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Your score %d", gc.currentScore)),
		),
		func(restart bool) {
			if restart {
				gc.joinGame()
			} else {
				gc.showJoinScreen()
			}
		},
		gc.mainWindow,
	)
}
