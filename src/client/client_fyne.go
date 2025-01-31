package main

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2/dialog"
	"image/color"
	"log"
	"sync"
	"time"

	pb "GoSnakeGame/proto"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/grpc"
)

const (
	width    = 20
	height   = 20
	cellSize = 20
)

var (
	playerName   string
	client       pb.SnakeGameClient
	stream       pb.SnakeGame_JoinGameClient
	dirCh        = make(chan pb.Direction)
	currentState *pb.GameState
	mu           sync.Mutex
	currentDir   pb.Direction
	mainWindow   fyne.Window // Сохраняем ссылку на главное окно
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client = pb.NewSnakeGameClient(conn)

	a := app.New()
	mainWindow = a.NewWindow("Snake Game")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Введите ваше имя")

	joinButton := widget.NewButton("Присоединиться к игре", func() {
		playerName = nameEntry.Text
		if playerName == "" {
			return
		}

		var err error
		stream, err = client.JoinGame(context.Background(), &pb.JoinRequest{PlayerName: playerName})
		if err != nil {
			log.Printf("Не удалось присоединиться к игре: %v", err)
			return
		}

		go receiveGameState()
		go sendDirection()

		mainWindow.SetContent(createGameUI())
	})

	mainWindow.SetContent(container.NewVBox(
		nameEntry,
		joinButton,
	))

	mainWindow.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		var direction pb.Direction
		switch event.Name {
		case fyne.KeyUp:
			if currentDir != pb.Direction_DOWN {
				direction = pb.Direction_UP
			} else {
				return
			}
		case fyne.KeyDown:
			if currentDir != pb.Direction_UP {
				direction = pb.Direction_DOWN
			} else {
				return
			}
		case fyne.KeyLeft:
			if currentDir != pb.Direction_RIGHT {
				direction = pb.Direction_LEFT
			} else {
				return
			}
		case fyne.KeyRight:
			if currentDir != pb.Direction_LEFT {
				direction = pb.Direction_RIGHT
			} else {
				return
			}
		default:
			return
		}

		currentDir = direction
		dirCh <- currentDir
	})

	mainWindow.Resize(fyne.NewSize(width*cellSize+150, height*cellSize))
	mainWindow.ShowAndRun()
}

func createGameUI() fyne.CanvasObject {
	gameContainer := container.NewWithoutLayout()
	objects := make(map[string]fyne.CanvasObject)
	prevFood := make(map[string]bool)

	board := canvas.NewRectangle(color.RGBA{R: 63, G: 66, B: 75, A: 255})
	board.Resize(fyne.NewSize(cellSize*height, cellSize*width))
	board.Move(fyne.NewPos(0, 0))
	gameContainer.Add(board)

	go func() {
		for {
			mu.Lock()
			if currentState != nil {
				currentFood := make(map[string]bool)
				for _, food := range currentState.Food {
					key := fmt.Sprintf("food-%d-%d", food.X, food.Y)
					currentFood[key] = true
				}

				for key := range prevFood {
					if !currentFood[key] {
						if obj, exists := objects[key]; exists {
							gameContainer.Remove(obj)
							delete(objects, key)
						}
					}
				}
				prevFood = currentFood

				for _, food := range currentState.Food {
					key := fmt.Sprintf("food-%d-%d", food.X, food.Y)
					if _, exists := objects[key]; !exists {
						foodRect := canvas.NewRectangle(color.RGBA{R: 255, G: 0, B: 0, A: 255})
						foodRect.Resize(fyne.NewSize(cellSize, cellSize))
						foodRect.Move(fyne.NewPos(float32(food.X)*cellSize, float32(food.Y)*cellSize))
						gameContainer.Add(foodRect)
						objects[key] = foodRect
					}
				}

				for _, player := range currentState.Players {
					snakeColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}
					if player.Name == playerName {
						snakeColor = color.RGBA{R: 0, G: 255, B: 0, A: 255}
					}

					for i, bodyPart := range player.Body {
						key := fmt.Sprintf("snake-%s-%d", player.Name, i)
						if obj, exists := objects[key]; exists {
							obj.Move(fyne.NewPos(float32(bodyPart.X)*cellSize, float32(bodyPart.Y)*cellSize))
						} else {
							bodyRect := canvas.NewRectangle(snakeColor)
							bodyRect.Resize(fyne.NewSize(cellSize+1, cellSize+1))
							bodyRect.Move(fyne.NewPos(float32(bodyPart.X)*cellSize, float32(bodyPart.Y)*cellSize))
							gameContainer.Add(bodyRect)
							objects[key] = bodyRect
						}
					}
				}
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)
		}
	}()

	return gameContainer
}

func receiveGameState() {
	for {
		state, err := stream.Recv()
		if err != nil {
			log.Printf("Ошибка получения состояния игры: %v", err)
			return
		}

		mu.Lock()
		currentState = state
		mu.Unlock()

		for _, player := range state.Players {
			if player.Name == playerName && !player.Alive {
				log.Println("Змейка умерла")
				showGameOverScreen()
				return
			}
		}
	}
}

func sendDirection() {
	for dir := range dirCh {
		_, err := client.SendDirection(context.Background(), &pb.DirectionRequest{
			PlayerName: playerName,
			Direction:  dir,
		})
		if err != nil {
			log.Printf("Ошибка отправки направления: %v", err)
		}
	}
}

func showGameOverScreen() {
	dialog.ShowCustomConfirm(
		"Игра окончена",
		"Перезапустить",
		"Выход",
		container.NewVBox(
			widget.NewLabel("Игра окончена!"),
		),
		func(restart bool) {
			if restart {
				//todo доделать
			} else {
				mainWindow.Close()
			}
		},
		mainWindow,
	)
}
