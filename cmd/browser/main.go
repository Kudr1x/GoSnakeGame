//go:build js && wasm

package main

import (
	"GoSnakeGame/internal/config"
	"context"
	"fmt"
	"image/color"
	"log"
	"strings"
	"sync"
	"syscall/js"
	"time"

	pb "GoSnakeGame/api/proto/snake/v1"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"
)

type gameClient struct {
	playerName   string
	conn         *websocket.Conn
	writeCh      chan []byte
	dirCh        chan pb.Direction
	currentState *pb.JoinGameResponse
	mu           sync.RWMutex
	currentDir   pb.Direction
	currentScore int
	mainWindow   fyne.Window
	stopGameCh   chan struct{}
	ctx          context.Context
	cancel       context.CancelFunc
	cfg          *config.ClientConfig

	rectPool []*canvas.Rectangle
	active   []*canvas.Rectangle

	topPlayersLabel *widget.Label

	gameOver func()
}

func main() {
	cfg := config.DefaultClientConfig()
	// No flag parsing in WASM usually, or simplified
	
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

	ctx, cancel := context.WithCancel(context.Background())
	conn, _, err := websocket.Dial(ctx, addr, nil)
	if err != nil {
		log.Printf("failed to connect to gateway %s: %v", addr, err)
		return
	}

	gc := &gameClient{
		conn:    conn,
		writeCh: make(chan []byte, 10),
		dirCh:   make(chan pb.Direction, 1),
		cfg:     cfg,
		ctx:     ctx,
		cancel:  cancel,
	}
	gc.gameOver = gc.showGameOverScreen

	go gc.writeLoop()

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
	if gc.stopGameCh != nil {
		close(gc.stopGameCh)
	}

	gc.stopGameCh = make(chan struct{})

	gc.mu.Lock()
	gc.rectPool = nil
	gc.active = nil
	gc.mu.Unlock()

	// Send Join message via WebSocket
	joinMsg := &pb.ClientMessage{
		Payload: &pb.ClientMessage_Join{
			Join: &pb.JoinGameRequest{PlayerName: gc.playerName},
		},
	}
	data, _ := proto.Marshal(joinMsg)
	select {
	case gc.writeCh <- data:
	case <-gc.ctx.Done():
	}

	gc.currentDir = pb.Direction_DIRECTION_UNSPECIFIED

	go gc.readLoop(gc.stopGameCh)
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

	gc.topPlayersLabel = widget.NewLabel("Loading top...")
	topX := float32(gc.cfg.Width*gc.cfg.CellSize + gc.cfg.SidebarScoreOffset)
	gc.topPlayersLabel.Move(fyne.NewPos(topX, float32(gc.cfg.SidebarTopOffset)))
	gameContainer.Add(gc.topPlayersLabel)

	go gc.updateTopPlayersLoop(stopCh, gc.topPlayersLabel)
	go gc.renderLoop(stopCh, gameContainer, scoreLabel)

	return gameContainer
}

func (gc *gameClient) updateTopPlayersLoop(stopCh chan struct{}, label *widget.Label) {
	ticker := time.NewTicker(gc.cfg.TopUpdateInterval)
	defer ticker.Stop()

	updateTop := func() {
		msg := &pb.ClientMessage{
			Payload: &pb.ClientMessage_Top{
				Top: &pb.GetTopPlayersRequest{},
			},
		}
		data, _ := proto.Marshal(msg)
		select {
		case gc.writeCh <- data:
		case <-gc.ctx.Done():
		}
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

func (gc *gameClient) drawPlayer(
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
	posX := float32(x * int32(gc.cfg.CellSize))
	posY := float32(y * int32(gc.cfg.CellSize))
	r.Move(fyne.NewPos(posX, posY))
	r.Show()

	gc.active = append(gc.active, r)
}

func (gc *gameClient) readLoop(stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
			_, data, err := gc.conn.Read(gc.ctx)
			if err != nil {
				log.Printf("ws read error: %v", err)
				return
			}

			var msg pb.ServerMessage
			if err := proto.Unmarshal(data, &msg); err != nil {
				continue
			}

			switch payload := msg.Payload.(type) {
			case *pb.ServerMessage_Update:
				state := payload.Update
				gc.mu.Lock()
				gc.currentState = state
				for _, p := range state.Players {
					if p.Name == gc.playerName {
						if p.Direction != pb.Direction_DIRECTION_UNSPECIFIED {
							gc.currentDir = p.Direction
						}
						if !p.Alive {
							gc.mu.Unlock()
							gc.gameOver()
							return
						}
					}
				}
				gc.mu.Unlock()
			case *pb.ServerMessage_Top:
				top := payload.Top
				playerList := make([]string, 0, len(top.TopPlayers)+1)
				playerList = append(playerList, "Top Players:")
				for _, p := range top.TopPlayers {
					playerList = append(playerList, fmt.Sprintf("%s: %d", p.PlayerName, p.Score))
				}
				gc.topPlayersLabel.SetText(strings.Join(playerList, "\n"))
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
			msg := &pb.ClientMessage{
				Payload: &pb.ClientMessage_Direction{
					Direction: &pb.SendDirectionRequest{
						PlayerName: gc.playerName,
						Direction:  dir,
					},
				},
			}
			data, _ := proto.Marshal(msg)
			select {
			case gc.writeCh <- data:
			case <-gc.ctx.Done():
			}

			gc.mu.Lock()
			gc.currentDir = dir
			gc.mu.Unlock()
		}
	}
}

func (gc *gameClient) writeLoop() {
	for {
		select {
		case <-gc.ctx.Done():
			return
		case data, ok := <-gc.writeCh:
			if !ok {
				return
			}
			if err := gc.conn.Write(gc.ctx, websocket.MessageBinary, data); err != nil {
				return
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
