.PHONY: all proto build build-wasm lint clean run-engine run-gateway run-terminal help

PROTO_DIR = api/proto
PROTO_SRC = $(PROTO_DIR)/snake/v1/snake.proto
BIN_DIR = bin
BIN_ENGINE = $(BIN_DIR)/game-engine
BIN_GATEWAY = $(BIN_DIR)/gateway
BIN_TERMINAL = $(BIN_DIR)/terminal
WEB_DIR = web
WASM_CLIENT = $(WEB_DIR)/client.wasm

MODULE = GoSnakeGame

proto:
	mkdir -p api/proto/snake/v1
	protoc --proto_path=. \
		--go_out=. --go_opt=module=$(MODULE) \
		--go-grpc_out=. --go-grpc_opt=module=$(MODULE) \
		$(PROTO_SRC)
	@echo "ok"

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_ENGINE) ./services/game-engine
	go build -o $(BIN_GATEWAY) ./services/gateway
	go build -o $(BIN_TERMINAL) ./cmd/terminal
	@echo "ok"

build-wasm:
	mkdir -p $(WEB_DIR)
	GOOS=js GOARCH=wasm go build -o $(WASM_CLIENT) ./cmd/browser
	@echo "ok"

lint:
	golangci-lint run ./...
	@echo "ok"

run-engine:
	go run ./services/game-engine

run-gateway:
	go run ./services/gateway

run-terminal:
	go run ./cmd/terminal

clean:
	rm -rf $(BIN_DIR) $(WASM_CLIENT)
	@echo "ok"
	
test:
	go test -v -race ./...
	@echo "ok"
