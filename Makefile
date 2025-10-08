.PHONY: build test install clean run help

# Variables
BINARY_NAME=docbrown
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) main.go
	@echo "✓ Built bin/$(BINARY_NAME)"

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "✓ Built all platform binaries"

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v -cover

# Install locally
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	go clean
	@echo "✓ Cleaned"

# Run the application
run: build
	./bin/$(BINARY_NAME)

# Run with auto command
demo: build
	./bin/$(BINARY_NAME) auto

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Formatted"

# Lint code
lint:
	@echo "Linting..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run
	@echo "✓ Linted"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✓ Dependencies updated"

# Show help
help:
	@echo "DocBrown - Makefile commands:"
	@echo ""
	@echo "  make build       - Build the binary"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make test        - Run tests"
	@echo "  make install     - Install binary to GOPATH/bin"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make run         - Build and run"
	@echo "  make demo        - Build and run 'auto' command"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Lint code"
	@echo "  make deps        - Download and tidy dependencies"
	@echo "  make help        - Show this help"
