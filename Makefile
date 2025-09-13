# Makefile for 8fs S3-compatible storage server

.PHONY: build clean test run docker help cross-platform install dev

# Variables
BINARY_NAME := 8fs
SOURCE_DIR := ./cmd/server
BUILD_DIR := ./bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Default target
all: build

# Help target
help:
	@echo "8fs S3-compatible Storage Server"
	@echo "================================"
	@echo "Available targets:"
	@echo "  build            Build the 8fs binary"
	@echo "  clean            Clean build artifacts"
	@echo "  test             Run tests"
	@echo "  run              Build and run the server"
	@echo "  docker           Build Docker image"
	@echo "  cross-platform   Build for multiple platforms"
	@echo "  install          Build and install to GOPATH/bin"
	@echo "  dev              Run in development mode with file watching"
	@echo "  help             Show this help message"

# Build the binary
build:
	@echo "🚀 Building 8fs..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.buildTime=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')'" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		$(SOURCE_DIR)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build using the build script (with more features)
build-script:
	@./build.sh

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean -cache
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v -race ./...

# Run integration tests specifically
test-integration:
	@echo "🔗 Running integration tests..."
	@go test -v -run TestIntegration

# Build and run
run: build
	@echo "🏃 Starting 8fs server..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run in development mode with default config
dev:
	@echo "🔧 Starting development server..."
	@DEFAULT_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE DEFAULT_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY go run $(SOURCE_DIR)

# Build Docker image
docker:
	@echo "🐳 Building Docker image..."
	@docker build -t 8fs:latest -t 8fs:$(VERSION) .
	@echo "✅ Docker image built: 8fs:latest"

# Build for multiple platforms
cross-platform:
	@./build.sh --cross-platform

# Install to GOPATH/bin
install:
	@echo "📦 Installing 8fs..."
	@go install $(SOURCE_DIR)
	@echo "✅ 8fs installed to $(shell go env GOPATH)/bin/server"

# Download dependencies
deps:
	@echo "📥 Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Generate code coverage report
coverage:
	@echo "📊 Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run benchmarks
bench:
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Run with race detector
race:
	@echo "🏁 Building with race detector..."
	CGO_ENABLED=1 go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-race $(SOURCE_DIR)
	@echo "✅ Race-enabled binary: $(BUILD_DIR)/$(BINARY_NAME)-race"

# Docker compose up
compose-up:
	@echo "🐳 Starting with docker-compose..."
	@docker-compose up -d

# Docker compose with monitoring
compose-monitoring:
	@echo "🐳 Starting with monitoring stack..."
	@docker-compose --profile monitoring up -d

# Docker compose down
compose-down:
	@echo "🐳 Stopping docker-compose..."
	@docker-compose down

# Show binary info
info:
	@echo "8fs Binary Information"
	@echo "====================="
	@echo "Binary: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(GO_VERSION)"
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo "Size: $$(du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1)"; \
		echo "Modified: $$(stat -f "%Sm" $(BUILD_DIR)/$(BINARY_NAME))"; \
	fi

# Quick development workflow
quick: clean build test

# Full workflow with cross-platform build
full: clean test cross-platform docker
