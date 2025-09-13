#!/bin/bash

# 8fs Build Script
# Builds the 8fs S3-compatible storage server from the new clean architecture

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="8fs"
SOURCE_DIR="./cmd/server"
BUILD_DIR="./bin"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
GO_VERSION=$(go version | cut -d' ' -f3)

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo "=================================="
    echo "🚀 Building 8fs Storage Server"
    echo "=================================="
    echo "Version: $VERSION"
    echo "Go Version: $GO_VERSION"
    echo "Build Time: $BUILD_TIME"
    echo "Source: $SOURCE_DIR"
    echo "Output: $BUILD_DIR/$BINARY_NAME"
    echo "=================================="
    echo
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.20 or later."
        exit 1
    fi
    
    # Check Go version (require 1.20+)
    GO_VERSION_NUM=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | cut -d'o' -f2)
    MAJOR=$(echo $GO_VERSION_NUM | cut -d'.' -f1)
    MINOR=$(echo $GO_VERSION_NUM | cut -d'.' -f2)
    
    if [ "$MAJOR" -lt 1 ] || ([ "$MAJOR" -eq 1 ] && [ "$MINOR" -lt 20 ]); then
        log_error "Go 1.20 or later is required. Current version: go$GO_VERSION_NUM"
        exit 1
    fi
    
    # Check if source directory exists
    if [ ! -d "$SOURCE_DIR" ]; then
        log_error "Source directory '$SOURCE_DIR' not found!"
        log_error "Make sure you're running this script from the project root."
        exit 1
    fi
    
    # Check if go.mod exists
    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found. Make sure you're in the project root."
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

clean_build() {
    log_info "Cleaning previous builds..."
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    log_success "Build directory cleaned"
}

download_deps() {
    log_info "Downloading dependencies..."
    go mod download
    go mod tidy
    log_success "Dependencies downloaded"
}

run_tests() {
    log_info "Running tests..."
    if go test -v ./... 2>/dev/null; then
        log_success "All tests passed"
    else
        log_warn "Some tests failed, but continuing with build..."
    fi
}

build_binary() {
    log_info "Building 8fs binary..."
    
    # Build flags
    BUILD_FLAGS=(
        -trimpath
        -ldflags "-s -w -X 'main.version=$VERSION' -X 'main.buildTime=$BUILD_TIME' -X 'main.goVersion=$GO_VERSION'"
    )
    
    # Set CGO_ENABLED=0 for static binary
    export CGO_ENABLED=0
    
    # Build for current platform
    if go build "${BUILD_FLAGS[@]}" -o "$BUILD_DIR/$BINARY_NAME" "$SOURCE_DIR"; then
        log_success "Binary built successfully: $BUILD_DIR/$BINARY_NAME"
    else
        log_error "Build failed!"
        exit 1
    fi
}

build_cross_platform() {
    log_info "Building cross-platform binaries..."
    
    # Platforms to build for
    platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64" 
        "darwin/arm64"
        "windows/amd64"
    )
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r -a platform_split <<< "$platform"
        GOOS="${platform_split[0]}"
        GOARCH="${platform_split[1]}"
        
        output_name="$BUILD_DIR/${BINARY_NAME}-${GOOS}-${GOARCH}"
        if [ "$GOOS" = "windows" ]; then
            output_name+=".exe"
        fi
        
        log_info "Building for $GOOS/$GOARCH..."
        
        if env GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build \
            -trimpath \
            -ldflags "-s -w -X 'main.version=$VERSION' -X 'main.buildTime=$BUILD_TIME'" \
            -o "$output_name" \
            "$SOURCE_DIR"; then
            log_success "Built $output_name"
        else
            log_error "Failed to build for $GOOS/$GOARCH"
        fi
    done
}

show_info() {
    echo
    log_info "Binary information:"
    if [ -f "$BUILD_DIR/$BINARY_NAME" ]; then
        ls -lh "$BUILD_DIR/$BINARY_NAME"
        echo
        log_info "To run the server:"
        echo "  ./$BUILD_DIR/$BINARY_NAME"
        echo
        log_info "To run with custom config:"
        echo "  DEFAULT_ACCESS_KEY=your-key DEFAULT_SECRET_KEY=your-secret ./$BUILD_DIR/$BINARY_NAME"
        echo
        log_info "To build Docker image:"
        echo "  docker build -t 8fs:latest ."
    fi
}

print_usage() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  --cross-platform  Build for multiple platforms"
    echo "  --no-tests       Skip running tests"
    echo "  --no-clean       Skip cleaning build directory"
    echo "  --help           Show this help message"
}

# Main execution
main() {
    local cross_platform=false
    local run_tests_flag=true
    local clean_flag=true
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --cross-platform)
                cross_platform=true
                shift
                ;;
            --no-tests)
                run_tests_flag=false
                shift
                ;;
            --no-clean)
                clean_flag=false
                shift
                ;;
            --help)
                print_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done
    
    print_header
    check_dependencies
    
    if [ "$clean_flag" = true ]; then
        clean_build
    fi
    
    download_deps
    
    if [ "$run_tests_flag" = true ]; then
        run_tests
    fi
    
    build_binary
    
    if [ "$cross_platform" = true ]; then
        build_cross_platform
    fi
    
    show_info
    
    log_success "Build completed! 🎉"
}

# Run main function with all arguments
main "$@"
