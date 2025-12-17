.PHONY: build clean test lint install uninstall release docker-build docker-run

# Variables
BINARY_NAME=go-to-run
VERSION=$(shell git describe --tags 2>/dev/null || echo "v0.0.1")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitHash=$(COMMIT_HASH) -s -w"

# Default target
all: build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/go-to-run

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux ./cmd/go-to-run

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME).exe ./cmd/go-to-run

# Build for macOS
build-macos:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-macos ./cmd/go-to-run

# Build all platforms
build-all: build-linux build-windows build-macos
	@echo "All builds completed!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)*
	rm -f *.exe
	rm -rf dist/
	rm -rf coverage.txt

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

# Run linters
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Install to system
install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)

# Uninstall from system
uninstall:
	@echo "Uninstalling from /usr/local/bin..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# Create release archive
release: clean build-all
	@echo "Creating release archives..."
	mkdir -p dist
	zip -j dist/$(BINARY_NAME)-linux-$(VERSION).zip $(BINARY_NAME)-linux
	zip -j dist/$(BINARY_NAME)-windows-$(VERSION).zip $(BINARY_NAME).exe
	zip -j dist/$(BINARY_NAME)-macos-$(VERSION).zip $(BINARY_NAME)-macos
	@echo "Release archives created in dist/"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .

# Docker run
docker-run:
	@echo "Running in Docker..."
	docker run --rm -it --privileged $(BINARY_NAME):latest

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build for current platform"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-all    - Build for all platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linters"
	@echo "  fmt          - Format code"
	@echo "  install      - Install to system"
	@echo "  uninstall    - Uninstall from system"
	@echo "  release      - Create release archives"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run in Docker"