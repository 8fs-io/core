#!/bin/bash

# 8fs Docker Management Scripts with AI Provider Support

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_ai() {
    echo -e "${PURPLE}[AI]${NC} $1"
}

# Default configuration
DEFAULT_COMPOSE_FILE="docker-compose.yml"
DEFAULT_ENV_FILE=".env"

# Build the Docker image
build() {
    print_status "Building 8fs Docker image..."
    docker build -t 8fs:latest .
    print_status "Build completed successfully!"
}

# Run with different AI providers
run() {
    local provider=${1:-"ollama"}
    local compose_file
    local env_file
    
    case "$provider" in
        "ollama"|"local")
            compose_file="docker-compose.yml"
            env_file=".env"
            print_ai "Starting 8fs with Ollama (local AI)..."
            ;;
        "openai")
            compose_file="docker-compose.openai.yml"
            env_file=".env"
            print_ai "Starting 8fs with OpenAI embeddings..."
            if [ ! -f "$env_file" ] && [ -f ".env.openai.example" ]; then
                print_warning "No .env file found. Please copy .env.openai.example to .env and set your OpenAI API key"
                print_info "Example: cp .env.openai.example .env && nano .env"
                exit 1
            fi
            ;;
        "bedrock")
            compose_file="docker-compose.bedrock.yml"
            env_file=".env"
            print_ai "Starting 8fs with AWS Bedrock embeddings..."
            if [ ! -f "$env_file" ] && [ -f ".env.bedrock.example" ]; then
                print_warning "No .env file found. Please copy .env.bedrock.example to .env and set your AWS credentials"
                print_info "Example: cp .env.bedrock.example .env && nano .env"
                exit 1
            fi
            ;;
        *)
            print_error "Unknown AI provider: $provider"
            print_info "Supported providers: ollama (default), openai, bedrock"
            exit 1
            ;;
    esac
    
    # Check if compose file exists
    if [ ! -f "$compose_file" ]; then
        print_error "Compose file not found: $compose_file"
        exit 1
    fi
    
    print_status "Using compose file: $compose_file"
    if [ -f "$env_file" ]; then
        print_status "Using environment file: $env_file"
        docker-compose --env-file "$env_file" -f "$compose_file" up -d
    else
        docker-compose -f "$compose_file" up -d
    fi
    
    print_status "Container started! Access at http://localhost:8080"
    print_status "Health check: http://localhost:8080/healthz"
    print_status "Metrics: http://localhost:8080/metrics"
    
    # Provider-specific info
    case "$provider" in
        "ollama"|"local")
            print_ai "Ollama available at: http://localhost:11434"
            print_ai "Model: all-minilm:latest (automatically pulled)"
            ;;
        "openai")
            print_ai "Using OpenAI API for embeddings"
            print_ai "Model: text-embedding-3-small (or as configured)"
            ;;
        "bedrock")
            print_ai "Using AWS Bedrock for embeddings"
            print_ai "Model: amazon.titan-embed-text-v1 (or as configured)"
            ;;
    esac
}

# Run with monitoring
run_with_monitoring() {
    local provider=${1:-"ollama"}
    local compose_file
    
    case "$provider" in
        "ollama"|"local")
            compose_file="docker-compose.yml"
            ;;
        "openai")
            compose_file="docker-compose.openai.yml"
            ;;
        "bedrock")
            compose_file="docker-compose.bedrock.yml"
            ;;
        *)
            print_error "Unknown AI provider: $provider"
            exit 1
            ;;
    esac
    
    print_status "Starting 8fs with monitoring..."
    docker-compose --profile monitoring -f "$compose_file" up -d
    print_status "Services started!"
    print_status "8fs: http://localhost:8080"
    print_status "Prometheus: http://localhost:9090"
}

# Stop containers
stop() {
    local provider=${1:-"all"}
    
    if [ "$provider" = "all" ]; then
        print_status "Stopping all 8fs containers..."
        docker-compose -f docker-compose.yml down 2>/dev/null || true
        docker-compose -f docker-compose.openai.yml down 2>/dev/null || true
        docker-compose -f docker-compose.bedrock.yml down 2>/dev/null || true
    else
        case "$provider" in
            "ollama"|"local")
                compose_file="docker-compose.yml"
                ;;
            "openai")
                compose_file="docker-compose.openai.yml"
                ;;
            "bedrock")
                compose_file="docker-compose.bedrock.yml"
                ;;
            *)
                print_error "Unknown AI provider: $provider"
                exit 1
                ;;
        esac
        print_status "Stopping 8fs with $provider provider..."
        docker-compose -f "$compose_file" down
    fi
    print_status "Containers stopped!"
}

# Ollama management functions
ollama_status() {
    print_ai "Checking Ollama status..."
    if docker-compose ps ollama &>/dev/null; then
        docker-compose exec ollama ollama list
    else
        print_error "Ollama container not running"
    fi
}

ollama_pull() {
    local model=${1:-"all-minilm:latest"}
    print_ai "Pulling Ollama model: $model"
    docker-compose exec ollama ollama pull "$model"
}

ollama_models() {
    print_ai "Available Ollama models:"
    docker-compose exec ollama ollama list
}

ollama_logs() {
    print_ai "Showing Ollama logs..."
    docker-compose logs -f ollama
}

# View logs
logs() {
    local service=${1:-"8fs"}
    local provider=${2:-"ollama"}
    
    case "$provider" in
        "ollama"|"local")
            compose_file="docker-compose.yml"
            ;;
        "openai")
            compose_file="docker-compose.openai.yml"
            ;;
        "bedrock")
            compose_file="docker-compose.bedrock.yml"
            ;;
        *)
            compose_file="docker-compose.yml"
            ;;
    esac
    
    docker-compose -f "$compose_file" logs -f "$service"
}

# Clean up
clean() {
    local provider=${1:-"all"}
    
    print_warning "Cleaning up Docker images and containers..."
    
    if [ "$provider" = "all" ]; then
        docker-compose -f docker-compose.yml down --rmi all --volumes --remove-orphans 2>/dev/null || true
        docker-compose -f docker-compose.openai.yml down --rmi all --volumes --remove-orphans 2>/dev/null || true
        docker-compose -f docker-compose.bedrock.yml down --rmi all --volumes --remove-orphans 2>/dev/null || true
    else
        case "$provider" in
            "ollama"|"local")
                compose_file="docker-compose.yml"
                ;;
            "openai")
                compose_file="docker-compose.openai.yml"
                ;;
            "bedrock")
                compose_file="docker-compose.bedrock.yml"
                ;;
            *)
                print_error "Unknown AI provider: $provider"
                exit 1
                ;;
        esac
        docker-compose -f "$compose_file" down --rmi all --volumes --remove-orphans
    fi
    
    print_status "Cleanup completed!"
}

# Show container status
status() {
    print_status "Container status:"
    
    # Check all possible compose files
    for compose_file in "docker-compose.yml" "docker-compose.openai.yml" "docker-compose.bedrock.yml"; do
        if [ -f "$compose_file" ]; then
            echo -e "\n${BLUE}=== $compose_file ===${NC}"
            docker-compose -f "$compose_file" ps 2>/dev/null || echo "No services running"
        fi
    done
    
    echo ""
    print_status "Health status:"
    curl -s http://localhost:8080/healthz || print_error "Health check failed - 8fs not running?"
}

# Test the API with different providers
test_api() {
    local provider=${1:-"auto"}
    
    print_status "Testing 8fs API..."
    
    # Test health endpoint
    echo "Testing health endpoint..."
    curl -f http://localhost:8080/healthz || { print_error "Health check failed"; exit 1; }
    
    # Test metrics endpoint
    echo -e "\nTesting metrics endpoint..."
    curl -f http://localhost:8080/metrics > /dev/null || { print_error "Metrics endpoint failed"; exit 1; }
    
    # Test AI-specific endpoints if available
    if [ "$provider" != "none" ]; then
        echo -e "\nTesting AI capabilities..."
        curl -s -X POST http://localhost:8080/api/v1/vectors/search/text \
            -H "Content-Type: application/json" \
            -d '{"query": "test search", "limit": 1}' > /dev/null && \
            print_ai "AI search endpoint available" || \
            print_warning "AI search endpoint not ready yet"
    fi
    
    print_status "API tests completed!"
}

# Setup environment files
setup_env() {
    local provider=${1}
    
    case "$provider" in
        "openai")
            if [ -f ".env.openai.example" ] && [ ! -f ".env" ]; then
                cp .env.openai.example .env
                print_status "Created .env from OpenAI example"
                print_warning "Please edit .env and set your OPENAI_API_KEY"
                print_info "nano .env"
            else
                print_info ".env file already exists or example not found"
            fi
            ;;
        "bedrock")
            if [ -f ".env.bedrock.example" ] && [ ! -f ".env" ]; then
                cp .env.bedrock.example .env
                print_status "Created .env from Bedrock example"
                print_warning "Please edit .env and set your AWS credentials"
                print_info "nano .env"
            else
                print_info ".env file already exists or example not found"
            fi
            ;;
        *)
            print_error "Unknown provider for setup: $provider"
            print_info "Use: $0 setup-env {openai|bedrock}"
            ;;
    esac
}

# Configuration validation
validate_config() {
    local provider=${1:-"ollama"}
    
    print_info "=== Validating $provider Configuration ==="
    
    case "$provider" in
        "ollama"|"local")
            compose_file="docker-compose.yml"
            print_ai "✓ Ollama: No additional configuration required"
            if docker --version > /dev/null 2>&1; then
                print_ai "✓ Docker available"
            else
                print_error "✗ Docker not found"
                return 1
            fi
            ;;
        "openai")
            compose_file="docker-compose.openai.yml"
            if [ -f ".env" ]; then
                if grep -q "OPENAI_API_KEY=sk-" .env; then
                    print_ai "✓ OpenAI API key configured"
                elif grep -q "OPENAI_API_KEY=sk-your-openai-api-key-here" .env; then
                    print_error "✗ Please set your actual OpenAI API key in .env"
                    return 1
                else
                    print_warning "! OpenAI API key may not be properly configured"
                fi
            else
                print_error "✗ .env file not found. Run: $0 setup-env openai"
                return 1
            fi
            ;;
        "bedrock")
            compose_file="docker-compose.bedrock.yml"
            if [ -f ".env" ]; then
                has_access_key=false
                has_secret_key=false
                
                if grep -q "AWS_BEDROCK_ACCESS_KEY_ID=.\+" .env; then
                    has_access_key=true
                    print_ai "✓ AWS Access Key configured"
                else
                    print_error "✗ Please set your AWS Access Key in .env"
                fi
                
                if grep -q "AWS_BEDROCK_SECRET_ACCESS_KEY=.\+" .env; then
                    has_secret_key=true  
                    print_ai "✓ AWS Secret Key configured"
                else
                    print_error "✗ Please set your AWS Secret Key in .env"
                fi
                
                if [ "$has_access_key" = false ] || [ "$has_secret_key" = false ]; then
                    return 1
                fi
            else
                print_error "✗ .env file not found. Run: $0 setup-env bedrock"
                return 1
            fi
            ;;
        *)
            print_error "Unknown provider: $provider"
            return 1
            ;;
    esac
    
    # Test compose file syntax
    if [ -f "$compose_file" ]; then
        print_ai "✓ Compose file found: $compose_file"
        if docker-compose -f "$compose_file" config > /dev/null 2>&1; then
            print_ai "✓ Compose file syntax valid"
        else
            print_error "✗ Compose file has syntax errors"
            return 1
        fi
    else
        print_error "✗ Compose file not found: $compose_file"
        return 1
    fi
    
    print_status "Configuration validation completed successfully!"
    return 0
}

# Show configuration info
show_config() {
    local provider=${1:-"all"}
    
    print_info "=== 8fs Configuration Info ==="
    
    if [ "$provider" = "all" ] || [ "$provider" = "ollama" ]; then
        echo -e "\n${PURPLE}🤖 Ollama (Local AI):${NC}"
        echo "  • Compose: docker-compose.yml"
        echo "  • Model: all-minilm:latest (384 dimensions)"
        echo "  • Memory: ~2GB+ required for Ollama"
        echo "  • Pros: Private, no API costs, offline capable"
        echo "  • Cons: Higher memory usage, slower on CPU"
    fi
    
    if [ "$provider" = "all" ] || [ "$provider" = "openai" ]; then
        echo -e "\n${PURPLE}🧠 OpenAI:${NC}"
        echo "  • Compose: docker-compose.openai.yml"
        echo "  • Model: text-embedding-3-small (1536 dimensions)"
        echo "  • Setup: cp .env.openai.example .env"
        echo "  • Pros: Fast, high-quality embeddings, low memory"
        echo "  • Cons: API costs, requires internet, data privacy"
    fi
    
    if [ "$provider" = "all" ] || [ "$provider" = "bedrock" ]; then
        echo -e "\n${PURPLE}☁️ AWS Bedrock:${NC}"
        echo "  • Compose: docker-compose.bedrock.yml"
        echo "  • Model: amazon.titan-embed-text-v1"
        echo "  • Setup: cp .env.bedrock.example .env"
        echo "  • Pros: Enterprise-grade, AWS integration, compliance"
        echo "  • Cons: AWS costs, requires AWS account, complexity"
    fi
    
    echo -e "\n${BLUE}Commands:${NC}"
    echo "  $0 run [provider]        # Start with provider (ollama|openai|bedrock)"
    echo "  $0 setup-env [provider]  # Setup environment file"
    echo "  $0 ollama-pull [model]   # Pull specific Ollama model"
}

# Show usage
usage() {
    echo "8fs Docker Management Script with AI Provider Support"
    echo ""
    echo "Usage: $0 {command} [options]"
    echo ""
    echo "Basic Commands:"
    echo "  build                    - Build the Docker image"
    echo "  run [provider]           - Start 8fs with AI provider (ollama|openai|bedrock)"
    echo "  run-monitoring [provider]- Start with Prometheus monitoring"
    echo "  stop [provider|all]      - Stop containers"
    echo "  logs [service] [provider]- Show container logs"
    echo "  clean [provider|all]     - Remove containers and images"
    echo "  status                   - Show container status"
    echo "  test [provider]          - Test API endpoints"
    echo ""
    echo "AI Provider Management:"
    echo "  setup-env {openai|bedrock} - Setup environment file for provider"
    echo "  config [provider]         - Show configuration info"
    echo "  validate [provider]       - Validate provider configuration"
    echo ""
    echo "Ollama-specific Commands:"
    echo "  ollama-status            - Show Ollama container status"
    echo "  ollama-models            - List available models"
    echo "  ollama-pull [model]      - Pull a specific model"
    echo "  ollama-logs              - Show Ollama logs"
    echo ""
    echo "Examples:"
    echo "  $0 run                   # Start with Ollama (default)"
    echo "  $0 run openai            # Start with OpenAI embeddings"
    echo "  $0 run bedrock           # Start with AWS Bedrock"
    echo "  $0 setup-env openai      # Setup OpenAI environment"
    echo "  $0 ollama-pull llama2    # Pull llama2 model"
    echo ""
    echo "AI Providers:"
    echo "  ollama  - Local AI with Ollama (default, privacy-first)"
    echo "  openai  - OpenAI API embeddings (fast, requires API key)"
    echo "  bedrock - AWS Bedrock embeddings (enterprise, requires AWS)"
}

# Main script logic
case "${1:-help}" in
    build)
        build
        ;;
    run)
        run "${2:-ollama}"
        ;;
    run-monitoring)
        run_with_monitoring "${2:-ollama}"
        ;;
    stop)
        stop "${2:-all}"
        ;;
    logs)
        logs "${2:-8fs}" "${3:-ollama}"
        ;;
    clean)
        clean "${2:-all}"
        ;;
    status)
        status
        ;;
    test)
        test_api "${2:-auto}"
        ;;
    setup-env)
        setup_env "${2}"
        ;;
    config)
        show_config "${2:-all}"
        ;;
    validate)
        validate_config "${2:-ollama}"
        ;;
    ollama-status)
        ollama_status
        ;;
    ollama-models)
        ollama_models
        ;;
    ollama-pull)
        ollama_pull "${2}"
        ;;
    ollama-logs)
        ollama_logs
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
