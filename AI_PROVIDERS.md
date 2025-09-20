# 8fs AI Provider Integration Guide

8fs supports multiple AI embedding providers for maximum flexibility in deployment scenarios. This document covers setup and usage for each provider.

## 🤖 Supported AI Providers

### 1. Ollama (Default) - Privacy-First Local AI
- **Best for**: Privacy-sensitive deployments, offline environments, development
- **Model**: all-minilm:latest (384 dimensions)
- **Memory**: ~2GB+ required
- **Setup**: Zero configuration, model auto-pulled

### 2. OpenAI - High-Performance Cloud AI
- **Best for**: Production deployments, high throughput, quality embeddings
- **Model**: text-embedding-3-small (1536 dimensions) or text-embedding-ada-002
- **Memory**: Minimal (~256MB)
- **Setup**: Requires OpenAI API key

### 3. AWS Bedrock - Enterprise-Grade AI
- **Best for**: Enterprise AWS environments, compliance requirements
- **Model**: amazon.titan-embed-text-v1, cohere.embed-english-v3
- **Memory**: Minimal (~256MB)
- **Setup**: Requires AWS credentials and Bedrock access

## 🚀 Quick Start

### Using Ollama (Default)
```bash
# Start with local Ollama
./docker.sh run ollama
# or simply
./docker.sh run
```

### Using OpenAI
```bash
# Setup OpenAI configuration
./docker.sh setup-env openai
# Edit .env file with your API key
nano .env
# Start with OpenAI
./docker.sh run openai
```

### Using AWS Bedrock
```bash
# Setup Bedrock configuration  
./docker.sh setup-env bedrock
# Edit .env file with AWS credentials
nano .env
# Start with Bedrock
./docker.sh run bedrock
```

## ⚙️ Configuration Details

### Environment Variables

#### Common AI Settings
- `AI_ENABLED`: Enable/disable AI features (default: `true`)
- `AI_PROVIDER`: AI provider (`ollama`, `openai`, `bedrock`)
- `AI_TIMEOUT`: Request timeout (default: `30s`)
- `AI_CHUNK_SIZE`: Text chunk size for processing (default: `500`)
- `AI_MAX_RETRIES`: Max retry attempts (default: `3`)

#### OpenAI Configuration
- `OPENAI_API_KEY`: Your OpenAI API key (required)
- `OPENAI_ORG_ID`: Organization ID (optional)
- `OPENAI_EMBED_MODEL`: Embedding model (default: `text-embedding-3-small`)

#### AWS Bedrock Configuration
- `AWS_BEDROCK_REGION`: AWS region (default: `us-east-1`)
- `AWS_BEDROCK_ACCESS_KEY_ID`: AWS access key (required)
- `AWS_BEDROCK_SECRET_ACCESS_KEY`: AWS secret key (required)
- `AWS_BEDROCK_EMBED_MODEL`: Bedrock model (default: `amazon.titan-embed-text-v1`)

### Docker Compose Files

Each provider has its own optimized Docker Compose configuration:

- `docker-compose.yml` - Ollama with local AI service
- `docker-compose.openai.yml` - OpenAI cloud embeddings
- `docker-compose.bedrock.yml` - AWS Bedrock embeddings

## 🛠️ Management Commands

### Basic Operations
```bash
# Build the image
./docker.sh build

# Start with specific provider
./docker.sh run ollama|openai|bedrock

# Check status
./docker.sh status

# View logs
./docker.sh logs 8fs ollama

# Stop services
./docker.sh stop all
```

### Ollama-Specific Commands
```bash
# Check Ollama status
./docker.sh ollama-status

# List available models
./docker.sh ollama-models

# Pull a specific model
./docker.sh ollama-pull llama2

# View Ollama logs
./docker.sh ollama-logs
```

### Provider Information
```bash
# Show all provider info
./docker.sh config

# Show specific provider
./docker.sh config openai
```

## 🔧 Advanced Configuration

### Custom Models

#### OpenAI Models
Available models (update in .env):
```bash
OPENAI_EMBED_MODEL=text-embedding-3-small  # 1536 dim, cost-effective
OPENAI_EMBED_MODEL=text-embedding-3-large  # 3072 dim, highest quality
OPENAI_EMBED_MODEL=text-embedding-ada-002  # 1536 dim, legacy
```

#### AWS Bedrock Models
Available models (update in .env):
```bash
AWS_BEDROCK_EMBED_MODEL=amazon.titan-embed-text-v1    # 1536 dim
AWS_BEDROCK_EMBED_MODEL=amazon.titan-embed-text-v2    # 1024 dim
AWS_BEDROCK_EMBED_MODEL=cohere.embed-english-v3       # 1024 dim
AWS_BEDROCK_EMBED_MODEL=cohere.embed-multilingual-v3  # 1024 dim
```

#### Custom Ollama Models
```bash
# Pull and use different models
./docker.sh ollama-pull nomic-embed-text
# Update docker-compose.yml to use the new model
```

### Resource Limits

Each provider has optimized resource limits:

**Ollama**: 512MB memory limit (for 8fs) + ~2GB for Ollama service
**OpenAI**: 256MB memory limit (lightweight)  
**Bedrock**: 256MB memory limit (lightweight)

### Production Considerations

#### OpenAI Production
- Set proper API rate limits
- Monitor token usage and costs
- Implement fallback strategies
- Consider data privacy implications

#### Bedrock Production  
- Configure proper IAM roles
- Set up VPC endpoints for private access
- Monitor AWS costs and quotas
- Implement proper error handling

#### Ollama Production
- Allocate sufficient memory (4GB+ recommended)
- Consider GPU acceleration for better performance
- Set up model management and updates
- Plan for horizontal scaling

## 🔍 Testing

### API Testing
```bash
# Test basic API
./docker.sh test

# Test with specific provider
./docker.sh test openai

# Manual embedding test
curl -X POST http://localhost:8080/api/v1/vectors/search/text \
  -H "Content-Type: application/json" \
  -d '{"query": "artificial intelligence", "limit": 5}'
```

### Performance Comparison
```bash
# Upload test document
curl -X PUT "http://localhost:8080/test-bucket/ai-doc.txt" \
  -H "Content-Type: text/plain" \
  -d "Artificial intelligence is transforming how we build software..."

# Search performance test
time curl -X POST http://localhost:8080/api/v1/vectors/search/text \
  -H "Content-Type: application/json" \
  -d '{"query": "machine learning software", "limit": 10}'
```

## 🚨 Troubleshooting

### Common Issues

#### OpenAI Provider
```bash
# Invalid API key error
# Solution: Check OPENAI_API_KEY in .env file

# Rate limit exceeded
# Solution: Implement exponential backoff or upgrade plan

# Model not found
# Solution: Check OPENAI_EMBED_MODEL value
```

#### Bedrock Provider
```bash
# Access denied error
# Solution: Check AWS credentials and IAM permissions

# Region not supported
# Solution: Update AWS_BEDROCK_REGION to supported region

# Model not available
# Solution: Check model availability in your region
```

#### Ollama Provider
```bash
# Out of memory
# Solution: Increase Docker memory limits or system RAM

# Model pull failed
# Solution: Check internet connection and retry
./docker.sh ollama-pull all-minilm:latest

# Slow performance
# Solution: Consider GPU acceleration or upgrade hardware
```

### Logs and Debugging
```bash
# Check 8fs logs
./docker.sh logs 8fs

# Check Ollama logs  
./docker.sh ollama-logs

# Check all container status
./docker.sh status

# Test health endpoints
curl http://localhost:8080/healthz
curl http://localhost:11434/api/tags  # Ollama only
```

## 📊 Performance Characteristics

| Provider | Latency | Memory | Cost | Privacy | Offline |
|----------|---------|---------|------|---------|---------|
| Ollama   | Medium  | High    | Free | High    | Yes     |
| OpenAI   | Low     | Low     | Pay  | Low     | No      |
| Bedrock  | Low     | Low     | Pay  | Medium  | No      |

## 🔗 Integration Examples

See the main README.md for complete integration examples and the client_examples/ directory for language-specific implementations.

---

For more details, see the main [README.md](./README.md) and visit our [documentation](./docs/).