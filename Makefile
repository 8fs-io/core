# Makefile for 8fs S3-compatible storage server

.PHONY: build clean test run docker help cross-platform install dev benchmark

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
	@echo ""
	@echo "Build & Run:"
	@echo "  build            Build the 8fs binary"
	@echo "  clean            Clean build artifacts" 
	@echo "  run              Build and run the server"
	@echo "  dev              Run in development mode"
	@echo ""
	@echo "Testing:"
	@echo "  test             Run tests"
	@echo "  test-integration Run integration tests"
	@echo "  coverage         Generate coverage report"
	@echo "  race             Build with race detector"
	@echo ""
	@echo "Performance & Benchmarking:"
	@echo "  benchmark-quick     Quick benchmark (100 vectors, 3D)"
	@echo "  benchmark-realistic Realistic benchmark (1000 vectors, 384D)" 
	@echo "  benchmark-large     Large benchmark (5000 vectors, 384D)"
	@echo "  benchmark-compare   Comparative benchmarks across configs"
	@echo ""
	@echo "Data Generation:"
	@echo "  generate-sample     Generate sample data (1000 realistic vectors)"
	@echo "  generate-clustered  Generate clustered dataset (1000 vectors)"
	@echo "  generate-random     Generate random dataset (1000 vectors)"
	@echo ""
	@echo "Docker:"
	@echo "  docker           Build Docker image"
	@echo "  compose-up       Start with docker-compose"
	@echo "  compose-down     Stop docker-compose"
	@echo ""
	@echo "Utilities:"
	@echo "  install          Install to GOPATH/bin"
	@echo "  fmt              Format code"
	@echo "  lint             Lint code"
	@echo "  info             Show binary information"

# Build the binary
build:
	@echo "🚀 Building 8fs..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build \
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

# Performance Benchmarking Targets
# =================================

# Build benchmark tool
benchmark-build:
	@echo "🔨 Building benchmark tool..."
	@go build -o $(BUILD_DIR)/benchmark ./cmd/benchmark/

# Build data generator tool
generate-data-build:
	@echo "🔨 Building data generator tool..."
	@go build -o $(BUILD_DIR)/generate-data ./cmd/generate-data/

# Quick benchmark for development (small dataset, fast)
benchmark-quick: benchmark-build
	@echo "⚡ Running quick benchmark..."
	@mkdir -p ./data
	@$(BUILD_DIR)/benchmark \
		-db ./data/benchmark_quick.db \
		-output ./data/benchmark_quick_results.json \
		-vectors 100 \
		-queries 25 \
		-dims 3 \
		-dataset random \
		-topk 5 \
		-verbose=false \
		-cleanup=true
	@echo "✅ Quick benchmark complete! Results in ./data/benchmark_quick_results.json"

# Realistic benchmark (production-like dataset)
benchmark-realistic: benchmark-build
	@echo "🎯 Running realistic benchmark..."
	@mkdir -p ./data
	@$(BUILD_DIR)/benchmark \
		-db ./data/benchmark_realistic.db \
		-output ./data/benchmark_realistic_results.json \
		-vectors 1000 \
		-queries 100 \
		-dims 384 \
		-dataset realistic \
		-topk 10 \
		-verbose=true \
		-cleanup=true
	@echo "✅ Realistic benchmark complete! Results in ./data/benchmark_realistic_results.json"

# Large benchmark (performance testing)
benchmark-large: benchmark-build
	@echo "🚀 Running large benchmark..."
	@mkdir -p ./data
	@$(BUILD_DIR)/benchmark \
		-db ./data/benchmark_large.db \
		-output ./data/benchmark_large_results.json \
		-vectors 5000 \
		-queries 200 \
		-dims 384 \
		-dataset realistic \
		-topk 10 \
		-verbose=true \
		-cleanup=true
	@echo "✅ Large benchmark complete! Results in ./data/benchmark_large_results.json"

# Comparative benchmarks (multiple configurations)
benchmark-compare: benchmark-build
	@echo "📊 Running comparative benchmarks..."
	@mkdir -p ./data
	@$(BUILD_DIR)/benchmark \
		-db ./data/benchmark_compare.db \
		-output ./data/benchmark_compare_results.json \
		-compare=true \
		-cleanup=true
	@echo "✅ Comparative benchmarks complete! Results in ./data/benchmark_compare_results.json"

# Generate sample data for testing
generate-sample: generate-data-build
	@echo "🎯 Generating sample data..."
	@mkdir -p ./data
	@$(BUILD_DIR)/generate-data -db ./data/sample_vectors.db -count 1000 -dims 384 -type realistic
	@echo "✅ Sample data generated in ./data/sample_vectors.db"

# Generate different dataset types
generate-clustered: generate-data-build
	@echo "🎯 Generating clustered dataset..."
	@mkdir -p ./data
	@$(BUILD_DIR)/generate-data -db ./data/clustered_vectors.db -count 1000 -dims 384 -type clustered
	@echo "✅ Clustered data generated in ./data/clustered_vectors.db"

generate-random: generate-data-build
	@echo "🎯 Generating random dataset..."
	@mkdir -p ./data  
	@$(BUILD_DIR)/generate-data -db ./data/random_vectors.db -count 1000 -dims 384 -type random
	@echo "✅ Random data generated in ./data/random_vectors.db"

# Clean benchmark data
benchmark-clean:
	@echo "🧹 Cleaning benchmark data..."
	@rm -f ./data/benchmark_*.db ./data/benchmark_*.json
	@rm -f ./data/sample_*.db ./data/clustered_*.db ./data/random_*.db
	@echo "✅ Benchmark data cleaned"

# Show benchmark results
benchmark-results:
	@echo "📈 Recent benchmark results:"
	@find ./data -name "benchmark_*_results.json" -exec echo "📄 {}" \; -exec jq -r '.config | "Vectors: \(.vector_count), Queries: \(.query_count), Dims: \(.dimensions), Dataset: \(.dataset_type)"' {} \; -exec jq -r '"Insert: \(.insert_metrics.throughput_per_sec | floor) vec/sec, Search: \(.search_metrics.throughput_per_sec | floor) query/sec"' {} \; -exec echo "" \; 2>/dev/null || echo "No results found or jq not installed"
