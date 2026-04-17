# ============================================================
# Digital Papyrus Backend — Makefile
# ============================================================

.PHONY: all build run test clean lint deps seed docker-build docker-run

# Variables
BINARY_NAME = digital-papyrus-api
BUILD_DIR = ./cmd/server
OUTPUT = ./$(BINARY_NAME)

# Default target
all: deps test build

# Install/download dependencies
deps:
	go mod download
	go mod tidy

# Build the application
build:
	go build -ldflags="-s -w" -trimpath -o $(OUTPUT) $(BUILD_DIR)

# Run the application in development mode
run:
	APP_ENV=development go run $(BUILD_DIR)/main.go

# Run all tests with verbose output and coverage
test:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run tests with HTML coverage report
test-cover: test
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(OUTPUT)
	rm -f coverage.out coverage.html
	rm -rf ./data/*.db

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

# Docker build
docker-build:
	docker build -t dp-backend:latest .

# Docker run
docker-run: docker-build
	docker run -p 8080:8080 --name dp-backend --rm dp-backend:latest
