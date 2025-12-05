# Edge Video V2 - Makefile
# Professional build automation for Windows, Linux, and macOS

# Variables
BINARY_NAME=edge-video-v2
VERSION=2.2.0
BUILD_DIR=bin
SRC_DIR=src
CONFIG_FILE=config.yaml
GO=go
GOFMT=gofmt
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
GCFLAGS=-gcflags="all=-trimpath=$(shell pwd)"
ASMFLAGS=-asmflags="all=-trimpath=$(shell pwd)"

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: all build build-prod build-debug clean test coverage fmt lint run help install deps cross-compile

# Default target
all: fmt lint test build

# Help
help:
	@echo "$(COLOR_BOLD)Edge Video V2 - Build Commands$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)Building:$(COLOR_RESET)"
	@echo "  make build          - Build debug binary (with symbols)"
	@echo "  make build-prod     - Build production binary (optimized)"
	@echo "  make build-debug    - Build with race detector"
	@echo ""
	@echo "$(COLOR_GREEN)Testing:$(COLOR_RESET)"
	@echo "  make test           - Run all tests"
	@echo "  make coverage       - Run tests with coverage report"
	@echo "  make bench          - Run benchmarks"
	@echo ""
	@echo "$(COLOR_GREEN)Code Quality:$(COLOR_RESET)"
	@echo "  make fmt            - Format code (gofmt)"
	@echo "  make lint           - Lint code (golangci-lint)"
	@echo "  make vet            - Run go vet"
	@echo ""
	@echo "$(COLOR_GREEN)Cross-Compilation:$(COLOR_RESET)"
	@echo "  make cross-compile  - Build for Windows, Linux, macOS"
	@echo "  make build-linux    - Build for Linux (amd64)"
	@echo "  make build-darwin   - Build for macOS (amd64)"
	@echo "  make build-rpi      - Build for Raspberry Pi (ARM7)"
	@echo ""
	@echo "$(COLOR_GREEN)Utilities:$(COLOR_RESET)"
	@echo "  make run            - Build and run"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make deps           - Download dependencies"
	@echo "  make install        - Install to GOPATH/bin"
	@echo ""

# Build debug version (default)
build:
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME) (debug)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME).exe ./$(SRC_DIR)
	@echo "$(COLOR_GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME).exe$(COLOR_RESET)"

# Build production version (optimized, stripped)
build-prod:
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME) (production)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe ./$(SRC_DIR)
	@echo "$(COLOR_GREEN)✓ Production build complete: $(BUILD_DIR)/$(BINARY_NAME).exe$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)  Size: $(shell du -h $(BUILD_DIR)/$(BINARY_NAME).exe | cut -f1)$(COLOR_RESET)"

# Build with race detector
build-debug:
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME) (race detector)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -race -o $(BUILD_DIR)/$(BINARY_NAME)-race.exe ./$(SRC_DIR)
	@echo "$(COLOR_GREEN)✓ Race detector build complete: $(BUILD_DIR)/$(BINARY_NAME)-race.exe$(COLOR_RESET)"

# Cross-compile for all platforms
cross-compile: build-linux build-darwin build-rpi
	@echo "$(COLOR_GREEN)✓ Cross-compilation complete!$(COLOR_RESET)"

# Build for Linux
build-linux:
	@echo "$(COLOR_BLUE)Building for Linux (amd64)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(SRC_DIR)
	@echo "$(COLOR_GREEN)✓ Linux build: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64$(COLOR_RESET)"

# Build for macOS
build-darwin:
	@echo "$(COLOR_BLUE)Building for macOS (amd64)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(SRC_DIR)
	@echo "$(COLOR_GREEN)✓ macOS build: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64$(COLOR_RESET)"

# Build for Raspberry Pi
build-rpi:
	@echo "$(COLOR_BLUE)Building for Raspberry Pi (ARM7)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-rpi ./$(SRC_DIR)
	@echo "$(COLOR_GREEN)✓ Raspberry Pi build: $(BUILD_DIR)/$(BINARY_NAME)-rpi$(COLOR_RESET)"

# Run tests
test:
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	$(GO) test -v ./$(SRC_DIR)/...
	@echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

# Run tests with coverage
coverage:
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	$(GO) test -coverprofile=coverage.out ./$(SRC_DIR)/...
	$(GO) tool cover -func=coverage.out
	@echo "$(COLOR_YELLOW)Generating HTML coverage report...$(COLOR_RESET)"
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: coverage.html$(COLOR_RESET)"

# Run benchmarks
bench:
	@echo "$(COLOR_BLUE)Running benchmarks...$(COLOR_RESET)"
	$(GO) test -bench=. -benchmem ./$(SRC_DIR)/...

# Format code
fmt:
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	$(GOFMT) -s -w $(SRC_DIR)
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

# Lint code
lint:
	@echo "$(COLOR_BLUE)Linting code...$(COLOR_RESET)"
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./$(SRC_DIR)/...; \
		echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed, skipping...$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)  Install: https://golangci-lint.run/usage/install/$(COLOR_RESET)"; \
	fi

# Run go vet
vet:
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	$(GO) vet ./$(SRC_DIR)/...
	@echo "$(COLOR_GREEN)✓ Vet complete$(COLOR_RESET)"

# Download dependencies
deps:
	@echo "$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)"
	$(GO) mod download
	$(GO) mod verify
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

# Clean build artifacts
clean:
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	rm -rf $(BUILD_DIR)/*
	rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

# Build and run
run: build
	@echo "$(COLOR_BLUE)Running $(BINARY_NAME)...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Using config: $(CONFIG_FILE)$(COLOR_RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME).exe -config $(CONFIG_FILE)

# Install to GOPATH/bin
install:
	@echo "$(COLOR_BLUE)Installing $(BINARY_NAME) to GOPATH/bin...$(COLOR_RESET)"
	$(GO) install $(LDFLAGS) ./$(SRC_DIR)
	@echo "$(COLOR_GREEN)✓ Installed to: $(shell go env GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"

# Docker build (future)
docker-build:
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	docker build -t edge-video-v2:$(VERSION) .
	@echo "$(COLOR_GREEN)✓ Docker image built: edge-video-v2:$(VERSION)$(COLOR_RESET)"

# Show version
version:
	@echo "$(COLOR_BOLD)Edge Video V2$(COLOR_RESET)"
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(shell go version)"
	@echo "Build Dir: $(BUILD_DIR)"
	@echo "Source Dir: $(SRC_DIR)"
