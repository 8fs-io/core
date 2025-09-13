#!/bin/bash

# 8fs Docker Management Scripts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Build the Docker image
build() {
    print_status "Building 8fs Docker image..."
    docker build -t 8fs:latest .
    print_status "Build completed successfully!"
}

# Run the container
run() {
    print_status "Starting 8fs container..."
    docker-compose up -d
    print_status "Container started! Access at http://localhost:8080"
    print_status "Health check: http://localhost:8080/healthz"
    print_status "Metrics: http://localhost:8080/metrics"
}

# Run with monitoring
run_with_monitoring() {
    print_status "Starting 8fs with monitoring..."
    docker-compose --profile monitoring up -d
    print_status "Services started!"
    print_status "8fs: http://localhost:8080"
    print_status "Prometheus: http://localhost:9090"
}

# Stop containers
stop() {
    print_status "Stopping containers..."
    docker-compose down
    print_status "Containers stopped!"
}

# View logs
logs() {
    docker-compose logs -f 8fs
}

# Clean up
clean() {
    print_warning "Cleaning up Docker images and containers..."
    docker-compose down --rmi all --volumes --remove-orphans
    print_status "Cleanup completed!"
}

# Show container status
status() {
    print_status "Container status:"
    docker-compose ps
    echo ""
    print_status "Health status:"
    docker-compose exec 8fs wget -qO- http://localhost:8080/healthz || print_error "Health check failed"
}

# Test the S3 API
test_s3() {
    print_status "Testing S3 API..."
    
    # Test health endpoint
    echo "Testing health endpoint..."
    curl -f http://localhost:8080/healthz || { print_error "Health check failed"; exit 1; }
    
    # Test metrics endpoint
    echo -e "\nTesting metrics endpoint..."
    curl -f http://localhost:8080/metrics > /dev/null || { print_error "Metrics endpoint failed"; exit 1; }
    
    print_status "Basic API tests passed!"
}

# Show usage
usage() {
    echo "8fs Docker Management Script"
    echo ""
    echo "Usage: $0 {build|run|run-monitoring|stop|logs|clean|status|test|help}"
    echo ""
    echo "Commands:"
    echo "  build           - Build the Docker image"
    echo "  run             - Start the 8fs container"
    echo "  run-monitoring  - Start with Prometheus monitoring"
    echo "  stop            - Stop all containers"
    echo "  logs            - Show container logs"
    echo "  clean           - Remove all containers and images"
    echo "  status          - Show container status"
    echo "  test            - Test the S3 API endpoints"
    echo "  help            - Show this help message"
}

# Main script logic
case "${1:-help}" in
    build)
        build
        ;;
    run)
        run
        ;;
    run-monitoring)
        run_with_monitoring
        ;;
    stop)
        stop
        ;;
    logs)
        logs
        ;;
    clean)
        clean
        ;;
    status)
        status
        ;;
    test)
        test_s3
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        print_error "Unknown command: $1"
        usage
        exit 1
        ;;
esac
