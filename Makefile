# Makefile for Image Recognition Web Application

.PHONY: build run test clean deps lint fmt dev docker docker-build docker-run help

# Variables
BINARY_NAME=image-recognition-webapp
BUILD_DIR=bin
MAIN_PATH=./cmd/server
DOCKER_IMAGE=image-recognition-webapp
VERSION?=latest

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for current OS
build-local:
	@echo "Building $(BINARY_NAME) for local OS..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the application locally
run: build-local
	@echo "Starting $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run in development mode with auto-reload
dev:
	@echo "Starting development server..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Lint code
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@which goimports > /dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	goimports -w .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean

# Security scan
security:
	@echo "Running security scan..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Run with Docker
docker-run:
	@echo "Running with Docker..."
	docker run -it --rm -p 8080:8080 $(DOCKER_IMAGE):latest

# Docker compose up
docker-up:
	@echo "Starting with Docker Compose..."
	docker-compose up -d

# Docker compose down
docker-down:
	@echo "Stopping Docker Compose..."
	docker-compose down

# Generate Go documentation
docs:
	@echo "Generating documentation..."
	@which godoc > /dev/null || (echo "Installing godoc..." && go install golang.org/x/tools/cmd/godoc@latest)
	godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/tools/cmd/godoc@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Create necessary directories
setup-dirs:
	@echo "Creating directories..."
	mkdir -p uploads temp models cache/models logs

# Setup development environment
setup: deps install-tools setup-dirs
	@echo "Development environment setup complete!"

# Deploy to staging
deploy-staging: build
	@echo "Deploying to staging..."
	# Add staging deployment commands here

# Deploy to production
deploy-prod: build test lint security
	@echo "Deploying to production..."
	# Add production deployment commands here

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application for Linux"
	@echo "  build-local    - Build the application for current OS"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Run in development mode with auto-reload"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  bench          - Run benchmarks"
	@echo "  deps           - Install dependencies"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  clean          - Clean build artifacts"
	@echo "  security       - Run security scan"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker"
	@echo "  docker-up      - Start with Docker Compose"
	@echo "  docker-down    - Stop Docker Compose"
	@echo "  docs           - Generate and serve documentation"
	@echo "  install-tools  - Install development tools"
	@echo "  setup          - Setup development environment"
	@echo "  setup-dirs     - Create necessary directories"
	@echo "  deploy-staging - Deploy to staging environment"
	@echo "  deploy-prod    - Deploy to production environment"
	@echo "  help           - Show this help message"