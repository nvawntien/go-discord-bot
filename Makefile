.PHONY: build run clean test

APP_NAME := go-discord-bot
BUILD_DIR := ./bin

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/bot

run: build
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

dev:
	@go run ./cmd/bot

clean:
	@rm -rf $(BUILD_DIR)
	@echo "Cleaned."

test:
	@go test -v ./...

lint:
	@golangci-lint run ./...
