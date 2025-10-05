.PHONY: build run dev clean test install-deps migrate-up migrate-down

# Build the server
build:
	go build -o bin/triply-server ./cmd/server

# Run the server
run: build
	./bin/triply-server

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f dev.db

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Install dependencies
install-deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Initialize database with migrations (placeholder for future)
migrate-up:
	@echo "Migrations not yet implemented"

migrate-down:
	@echo "Migrations not yet implemented"

# Generate mocks (requires mockery if needed)
generate:
	go generate ./...

# Run the server in production mode
prod: build
	GO_ENV=production ./bin/triply-server

# Help
help:
	@echo "Available commands:"
	@echo "  make build          - Build the server"
	@echo "  make run            - Build and run the server"
	@echo "  make dev            - Run with hot reload (requires air)"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make install-deps   - Install dependencies"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make prod           - Run in production mode"

