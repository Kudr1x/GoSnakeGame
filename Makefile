.PHONY: all proto build lint lint-go lint-proto clean run-server run-client help

PROTO_DIR = api/proto
PROTO_SRC = $(PROTO_DIR)/snake/v1/snake.proto
SERVER_MAIN = cmd/server/main.go
CLIENT_MAIN = cmd/client/main.go
BIN_DIR = bin
BIN_SERVER = $(BIN_DIR)/server
BIN_CLIENT = $(BIN_DIR)/client

MODULE = GoSnakeGame

all: proto build

proto:
	mkdir -p api/proto/snake/v1
	protoc --proto_path=. \
		--go_out=. --go_opt=module=$(MODULE) \
		--go-grpc_out=. --go-grpc_opt=module=$(MODULE) \
		$(PROTO_SRC)
	@echo "ok"

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_SERVER) $(SERVER_MAIN)
	go build -o $(BIN_CLIENT) $(CLIENT_MAIN)
	@echo "ok"

lint:
	golangci-lint run ./...
	@echo "ok"

run-server:
	go run $(SERVER_MAIN)

run-client:
	go run $(CLIENT_MAIN)

clean:
	rm -rf $(BIN_DIR)
	@echo "ok"
	
test:
	go test -v -race ./...
	@echo "ok"

help:
	@echo "Usage:"
	@echo "  make proto       - Generate Go code from .proto files"
	@echo "  make build       - Build server and client binaries"
	@echo "  make lint        - Run all linters (Go and Proto)"
	@echo "  make run-server  - Run server using 'go run'"
	@echo "  make run-client  - Run client using 'go run'"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make test        - Run tests"
