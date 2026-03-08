// Package main implements the client for the Snake Game.
package main

import (
	"GoSnakeGame/internal/config"
	"context"
	"fmt"
	"image/color"
	"log"
	"os"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gameClient struct {
	playerName   string
	client       pb.SnakeGameServiceClient
	stream       pb.SnakeGameService_JoinGameClient
	dirCh        chan pb.Direction
	currentState *pb.JoinGameResponse
	mu           sync.RWMutex
	currentDir   pb.Direction
	currentScore int
	mainWindow   fyne.Window
	stopGameCh   chan struct{}
	cancelStream context.CancelFunc
	cfg          *config.ClientConfig

	// UI object pool
	rectPool []*canvas.Rectangle
	active   []*canvas.Rectangle
}

func main() {
	cfg := config.DefaultClientConfig()
	cfg.ParseFlags()

	conn, err := grpc.NewClient(cfg.ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("failed to connect to %s: %v", cfg.ServerAddr, err)
		os.Exit(1)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	gc := &gameClient{
		client: pb.NewSnakeGameServiceClient(conn),
		dirCh:  make(chan pb.Direction, 1),
		cfg:    cfg,
	}

	a := app.New()
	gc.mainWindow = a.NewWindow("Snake Game")
	gc.mainWindow.SetFixedSize(true)

	gc.showJoinScreen()

	gc.mainWindow.Canvas().SetOnTypedKey(gc.handleKeyPress)

	w := float32(cfg.Width*cfg.CellSize + cfg.SidebarWidth)
	h := float32(cfg.Height * cfg.CellSize)
	gc.mainWindow.Resize(fyne.NewSize(w, h))
	gc.mainWindow.ShowAndRun()
}

func (gc *gameClient) showJoinScreen() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter your name")

	joinButton := widget.NewButton("Join the game", func() {
		gc.playerName = nameEntry.Text
		if gc.playerName == "" {
			return
		}

		gc.joinGame()
	})

	gc.mainWindow.SetContent(container.NewVBox(nameEntry, joinButton))
}

func (gc *gameClient) handleKeyPress(event *fyne.KeyEvent) {
	gc.mu.RLock()
	prevDir := gc.currentDir
	gc.mu.RUnlock()

	direction := gc.getDirectionFromKey(event.Name, prevDir)

	if direction != pb.Direction_DIRECTION_UNSPECIFIED {
		select {
		case gc.dirCh <- direction:
		default:
		}
	}
}

func (gc *gameClient) getDirectionFromKey(key fyne.KeyName, prevDir pb.Direction) pb.Direction {
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

func (gc *gameClient) joinGame() {
	if gc.cancelStream != nil {
		gc.cancelStream()
	}

	if gc.stopGameCh != nil {
		close(gc.stopGameCh)
	}

	gc.stopGameCh = make(chan struct{})

	// Clear pools because we are creating a new UI container
	gc.mu.Lock()
	gc.rectPool = nil
	gc.active = nil
	gc.mu.Unlock()

	var (
		err error
		ctx context.Context
	)

	ctx, gc.cancelStream = context.WithCancel(context.Background())

	gc.stream, err = gc.client.JoinGame(ctx, &pb.JoinGameRequest{PlayerName: gc.playerName})
	if err != nil {
		log.Printf("failed to join game: %v", err)

		return
	}

	gc.currentDir = pb.Direction_DIRECTION_UP

	go gc.receiveGameState(gc.stopGameCh)
	go gc.sendDirection(gc.stopGameCh)

	gc.mainWindow.SetContent(gc.createGameUI(gc.stopGameCh))
}

func (gc *gameClient) createGameUI(stopCh chan struct{}) fyne.CanvasObject {
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

func (gc *gameClient) updateTopPlayersLoop(stopCh chan struct{}, label *widget.Label) {
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

func (gc *gameClient) renderLoop(
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

func (gc *gameClient) drawFrame(
	container *fyne.Container,
	scoreLabel *widget.Label,
	state *pb.JoinGameResponse,
) {
	// Put all currently active rects back to pool
	for _, r := range gc.active {
		r.Hide()
	}

	gc.rectPool = append(gc.rectPool, gc.active...)
	gc.active = gc.active[:0]

	// Draw Food
	foodColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for _, f := range state.Food {
		gc.drawRect(container, f.X, f.Y, foodColor)
	}

	// Draw Players
	for _, p := range state.Players {
		if !p.Alive && p.Name != gc.playerName {
			continue
		}

		snakeColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}
		if p.Name == gc.playerName {
			snakeColor = color.RGBA{R: 0, G: 255, B: 0, A: 255}
			gc.currentScore = len(p.Body) * gc.cfg.ScoreMultiplier
			scoreLabel.SetText(fmt.Sprintf("Score: %d", gc.currentScore))
		}

		for _, bodyPart := range p.Body {
			gc.drawRect(container, bodyPart.X, bodyPart.Y, snakeColor)
		}
	}

	container.Refresh()
}

func (gc *gameClient) drawRect(container *fyne.Container, x, y int32, fillColor color.Color) {
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

func (gc *gameClient) getStringTop() string {
	ctx, cancel := context.WithTimeout(context.Background(), gc.cfg.TopPlayersTimeout)
	defer cancel()

	topPlayers, err := gc.client.GetTopPlayers(ctx, &pb.GetTopPlayersRequest{})
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

func (gc *gameClient) receiveGameState(stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
			state, err := gc.stream.Recv()
			if err != nil {
				log.Printf("connection lost: %v", err)

				return
			}

			gc.mu.Lock()
			gc.currentState = state
			gc.mu.Unlock()

			for _, p := range state.Players {
				if p.Name == gc.playerName && !p.Alive {
					gc.showGameOverScreen()

					return
				}
			}
		}
	}
}

func (gc *gameClient) sendDirection(stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case dir := <-gc.dirCh:
			ctx, cancel := context.WithTimeout(context.Background(), gc.cfg.DirectionTimeout)

			_, err := gc.client.SendDirection(ctx, &pb.SendDirectionRequest{
				PlayerName: gc.playerName,
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

func (gc *gameClient) showGameOverScreen() {
	dialog.ShowCustomConfirm(
		"Game over",
		"Reboot",
		"Exit",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Your score %d", gc.currentScore)),
		),
		func(restart bool) {
			if restart {
				gc.joinGame()
			} else {
				gc.mainWindow.Close()
			}
		},
		gc.mainWindow,
	)
}
