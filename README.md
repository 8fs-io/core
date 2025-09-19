# 8fs – S3-Compatible Storage Server with Native Vector Search

## 🌟 Vision

**8fs is the first S3-compatible storage server with built-in vector storage for AI developers.**

Not a MinIO clone—8fs unifies object storage and vector embeddings in one lightweight binary for local AI labs, from laptops to Raspberry Pi clusters.

> *S3 + vector storage in one <50MB binary—perfect for indie AI workflows.*

See [`VISION.md`](./VISION.md) for the full vision, market, and technical roadmap.

> **Go module path:** `github.com/8fs-io/core`

## 🆕 Recent Changes

- **🤖 RAG (Retrieval-Augmented Generation) Server:**
  - **NEW**: Complete RAG implementation with `/api/v1/chat/completions` (OpenAI compatible)
  - **Auto Context Retrieval**: Query → Vector Search → Context Formation → AI Generation
  - **Multi-Provider Support**: Ollama (Llama), OpenAI (GPT), AWS Bedrock (Claude)
  - **Production Ready**: Token usage tracking, relevance filtering, context management
- **🚀 Async Document Indexing:**
  - **NEW**: Automatic background processing for all uploaded text documents
  - Zero-configuration AI-native storage - upload any text and it becomes instantly searchable
  - Background worker system with retry logic and job status tracking
  - **Performance**: Non-blocking uploads with async embedding generation via Ollama
- **🔍 Text-Based Semantic Search:**
  - **NEW**: `POST /api/v1/vectors/search/text` - Search with plain text queries
  - No manual embedding generation required - just upload and search
  - Semantic similarity matching using all-minilm:latest model (384 dimensions)
  - **User Experience**: From upload to searchable in seconds, automatically
- **📊 Enhanced Testing & Performance:**
  - **NEW**: Comprehensive S3 compatibility test suite (PR #34)
  - **NEW**: Complete benchmarking tools and performance documentation
  - **NEW**: Data generation utilities for testing and development
  - **Validated**: Production-scale performance (1,700+ vec/sec insert)
- **Pure SQLite-vec Vector Storage:**
  - Pure sqlite-vec implementation with CGO support for maximum performance
  - Advanced dimension handling (3-1536) with intelligent validation
  - **Performance**: 1,700+ vec/sec insert, 8.9 search/sec at production scale
- **Comprehensive Performance Suite:**
  - Published benchmark results: [PERFORMANCE.md](./PERFORMANCE.md)
  - Multiple dataset generators: random, clustered, realistic patterns
  - Automated benchmarking tools via `make benchmark-*` targets
  - Production-validated with up to 5,000 vectors at 384 dimensions

## 🗺 Roadmap / Next Steps

- **🤖 AI Integration & Async Processing**
  - [x] Async document indexing with background workers
  - [x] Text-based semantic search API (`/api/v1/vectors/search/text`)
  - [x] Automatic embedding generation via Ollama integration
  - [x] Job status tracking and monitoring APIs
  - [x] Zero-configuration AI-native storage experience
  - [ ] Multi-model support (different embedding models per bucket)
  - [ ] Chunking strategies for large documents
  - [ ] Batch processing optimization for high-volume uploads
- **Vector storage**
  - [x] Pure SQLite-vec implementation with CGO support
  - [x] Production-scale performance validation (1,700+ vec/sec)
  - [x] Comprehensive benchmarking suite with published metrics
  - [x] Advanced error handling and dimension validation
  - [ ] SIMD acceleration for vector math operations  
  - [ ] Parallel search execution for multi-query workloads
  - [ ] Advanced indexing strategies for >10K vector datasets
- **S3 compatibility**
  - [x] Complete S3 API implementation with async AI processing
  - [x] Bucket-level indexing control (enabled by default)
  - [x] **Comprehensive S3 compatibility test suite** - **NEW!**
  - [x] AWS CLI integration testing and documentation
- **General**
  - [x] Docker and container deployment ready
  - [x] Comprehensive monitoring and health checks
  - [ ] Multi-tenant support: Tenant isolation and management
  - [ ] Web UI: Dashboard for storage and vector management
  - [ ] Enhanced backends: Additional storage drivers beyond filesystem
  - [ ] Advanced features: Versioning, lifecycle policies, metadata search

---

**8fs** is a high-performance, S3-compatible storage server built with Go, featuring clean architecture and production-ready deployment options.  
Perfect for developers who want a simple, self-hosted storage solution.

## 🤖 AI-Native Storage

**NEW!** 8fs now provides zero-configuration AI capabilities:

- **📄 Text-to-Vector**: Upload any text document → automatically indexed for semantic search
- **🔍 Plain Text Search**: Search with natural language - no embeddings required
- **⚡ Async Processing**: Non-blocking indexing with background workers
- **🧠 Ollama Integration**: Local AI model (all-minilm) for privacy-first embeddings

```bash
# Upload a text file - it becomes instantly searchable
curl -X PUT "http://localhost:8080/my-bucket/docs.txt" \
     -H "Content-Type: text/plain" \
     -d "Machine learning is transforming software development"

# Search with plain text - finds semantically similar content
curl -X POST "http://localhost:8080/api/v1/vectors/search/text" \
     -H "Content-Type: application/json" \
     -d '{"query": "AI software engineering", "limit": 5}'
```

---

## 🚀 Quick Start

### Prerequisites
- Go 1.20+ 
- CGO enabled (for sqlite-vec support)
- Docker (optional, for containerized deployment)

### Build and Run

#### Option 1: Using the Build Script (Recommended)
```bash
# Simple build
./build.sh

# Cross-platform build for multiple architectures
./build.sh --cross-platform

# Build without running tests
./build.sh --no-tests

# Build without cleaning previous builds
./build.sh --no-clean
```

#### Option 2: Using Makefile
```bash
# Show all available commands
make help

# Build the binary
make build

# Build and run
make run

# Run in development mode
make dev

# Build for multiple platforms
make cross-platform

# Build Docker image
make docker
```

#### Option 3: Direct Go Build
```bash
# Build with CGO enabled for sqlite-vec support
CGO_ENABLED=1 go build -o bin/8fs ./cmd/server
```

### Running the Server

#### Default Configuration
```bash
./bin/8fs
```

#### With Custom Credentials
```bash
DEFAULT_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE \
DEFAULT_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
./bin/8fs
```

#### Using Docker
```bash
# Build image
docker build -t 8fs:latest .

# Run container
docker run -p 8080:8080 \
  -e DEFAULT_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE \
  -e DEFAULT_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
  8fs:latest
```

#### Using Docker Compose
```bash
# Start AI-enabled services (includes Ollama for text search)
docker-compose up -d

# Start with monitoring (Prometheus + Grafana)
docker-compose --profile monitoring up -d

# Your AI-native 8fs server is now running!
# S3-compatible API: http://localhost:8080
# Vector API: http://localhost:8080/api/v1/vectors
# Text Search API: http://localhost:8080/api/v1/vectors/search/text
# Health: http://localhost:8080/healthz
# Ollama AI Service: http://localhost:11434 (internal)
```

### Vector Storage Quick Start

Once your server is running, you can start using vector storage:

#### 🚀 NEW: Upload & Search with Text (Zero Setup!)

```bash
# 1. Upload any text document - it becomes automatically searchable
curl -X PUT "http://localhost:8080/my-bucket/document.txt" \
  -H "Content-Type: text/plain" \
  --data-binary "This is my document about machine learning and AI."

# 2. Search with plain text - no embeddings needed!
curl -X POST "http://localhost:8080/api/v1/vectors/search/text" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "artificial intelligence",
    "top_k": 5
  }'
```

#### Advanced: Direct Vector Operations

#### Store a Vector Embedding
```bash
curl -X POST http://localhost:8080/api/v1/vectors/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc1",
    "embedding": [0.1, 0.2, 0.3, 0.4, 0.5],
    "metadata": {"title": "Sample Document", "category": "example"}
  }'
```

#### Search Similar Vectors
```bash
curl -X POST http://localhost:8080/api/v1/vectors/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": [0.1, 0.2, 0.3, 0.4, 0.5],
    "top_k": 5
  }'
```

#### Generate Sample Data for Testing
```bash
# Generate 100 sample vectors for testing
./bin/generate-data --count 100 --dims 384 --type realistic

# Run performance benchmarks
make benchmark-quick
```

---

## ✅ Features

### 🤖 RAG (Retrieval-Augmented Generation) - **NEW!**
- ✅ **OpenAI-Compatible Chat API**: `POST /api/v1/chat/completions` - full RAG pipeline
- ✅ **Context Retrieval**: `POST /api/v1/chat/search/context` - semantic document search
- ✅ **Multi-Provider Support**: Ollama (Llama), OpenAI (GPT), AWS Bedrock (Claude)
- ✅ **Smart Context Management**: Relevance filtering, source tracking, token optimization
- ✅ **Production Ready**: Token usage tracking, performance monitoring, health checks

### 🤖 AI-Native Storage
- ✅ **Async Document Indexing**: Auto-processing of uploaded text documents
- ✅ **Text-Based Search**: `POST /api/v1/vectors/search/text` - search with plain text
- ✅ **Zero Configuration**: Upload text → instantly searchable, no setup required
- ✅ **Background Workers**: Non-blocking processing with retry logic
- ✅ **Job Monitoring**: Status tracking and health check APIs
- ✅ **Ollama Integration**: Automatic embedding generation via all-minilm model

### S3 API Compatibility
- ✅ Bucket operations (create, delete, list)
- ✅ Object operations (upload, download, delete) 
- ✅ Object metadata and listing
- ✅ AWS Signature v4 authentication
- ✅ **Bucket-level indexing control** (enabled by default)

### Vector Storage API
- ✅ **Text search** (`POST /api/v1/vectors/search/text`) - **NEW!**
- ✅ Embedding storage (`POST /api/v1/vectors/embeddings`)
- ✅ Cosine similarity search (`POST /api/v1/vectors/search`)
- ✅ Vector retrieval (`GET /api/v1/vectors/embeddings/:id`)
- ✅ Vector deletion (`DELETE /api/v1/vectors/embeddings/:id`)
- ✅ Vector listing (`GET /api/v1/vectors/embeddings`)
- ✅ Dimension validation (3-1,536 dimensions)
- ✅ Metadata filtering and search
- ✅ **Comprehensive benchmarking suite** - **NEW!**
- ✅ **Performance metrics & monitoring** - **NEW!**

### Production Ready
- 📊 **Metrics**: Prometheus integration
- 📝 **Logging**: Structured logs with audit trails
- ❤️ **Health Checks**: `/healthz` endpoint
- 🐳 **Docker**: Ready-to-deploy containers
- 🚀 **Performance**: Optimized builds

### API Endpoints

#### S3-Compatible Endpoints
- `GET /healthz` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /` - List buckets
- `PUT /:bucket` - Create bucket
- `DELETE /:bucket` - Delete bucket
- `GET /:bucket` - List objects
- `PUT /:bucket/:key` - Store object
- `GET /:bucket/:key` - Retrieve object
- `HEAD /:bucket/:key` - Get object metadata
- `DELETE /:bucket/:key` - Delete object

#### Vector Storage Endpoints
- `POST /api/v1/vectors/embeddings` - Store vector embedding
- `POST /api/v1/vectors/search` - Search similar vectors
- `GET /api/v1/vectors/embeddings/:id` - Retrieve specific vector
- `GET /api/v1/vectors/embeddings` - List all vectors
- `DELETE /api/v1/vectors/embeddings/:id` - Delete vector

#### RAG (Retrieval-Augmented Generation) Endpoints - **NEW!**
- `POST /api/v1/chat/completions` - OpenAI-compatible RAG chat completions
- `POST /api/v1/chat/search/context` - Search and retrieve relevant context
- `POST /api/v1/chat/generate/context` - Generate text with provided context
- `GET /api/v1/chat/health` - RAG service health check

---

## 🏗️ Architecture

Clean architecture with clear separation of concerns:

```
8fs/
├── cmd/server/                 # Application entry point
├── internal/
│   ├── config/                 # Configuration management
│   ├── container/              # Dependency injection
│   ├── domain/storage/         # Business logic
│   ├── infrastructure/storage/ # Data persistence
│   └── transport/http/         # HTTP transport layer
├── pkg/
│   ├── errors/                 # Error utilities
│   └── logger/                 # Structured logging
├── build.sh                    # Build automation script
├── Makefile                    # Make-based build targets
├── Dockerfile                  # Container configuration
└── docker-compose.yml          # Multi-service deployment
```

## Configuration

Environment variables:

### Basic Configuration
- `DEFAULT_ACCESS_KEY`: AWS access key (default: `AKIAIOSFODNN7EXAMPLE`)
- `DEFAULT_SECRET_KEY`: AWS secret key (default: `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`)
- `PORT`: Server port (default: `8080`)
- `STORAGE_PATH`: Storage directory (default: `./storage`)
- `GIN_MODE`: Gin mode (`debug`, `release`) (default: `debug`)

### 🤖 AI Integration Configuration (NEW!)
- `AI_ENABLED`: Enable/disable AI features (default: `true`)
- `AI_PROVIDER`: AI provider (`ollama`, `openai`, `bedrock`) (default: `ollama`)
- `OLLAMA_BASE_URL`: Ollama server endpoint (default: `http://ollama:11434`)
- `OLLAMA_MODEL`: Embedding model to use (default: `all-minilm:latest`)

#### OpenAI Provider (NEW!)
- `OPENAI_API_KEY`: Your OpenAI API key (required for openai provider)
- `OPENAI_ORG_ID`: Organization ID (optional)
- `OPENAI_EMBED_MODEL`: Embedding model (default: `text-embedding-3-small`)

#### AWS Bedrock Provider (NEW!)
- `AWS_BEDROCK_REGION`: AWS region (default: `us-east-1`)
- `AWS_BEDROCK_ACCESS_KEY_ID`: AWS access key (required for bedrock provider)
- `AWS_BEDROCK_SECRET_ACCESS_KEY`: AWS secret key (required for bedrock provider)
- `AWS_BEDROCK_EMBED_MODEL`: Bedrock model (default: `amazon.titan-embed-text-v1`)

### 🐳 Docker AI Provider Setup

8fs now supports multiple AI embedding providers with dedicated Docker configurations:

```bash
# Local Ollama (default) - Privacy-first
./docker.sh run ollama

# OpenAI - High-performance cloud embeddings
./docker.sh setup-env openai  # Setup API key
./docker.sh run openai

# AWS Bedrock - Enterprise-grade embeddings
./docker.sh setup-env bedrock  # Setup AWS credentials
./docker.sh run bedrock

# Show provider information
./docker.sh config
```

> **Provider Comparison**: See [AI_PROVIDERS.md](./AI_PROVIDERS.md) for detailed provider comparison, setup guides, and performance characteristics.

> **Note**: The text search feature requires Ollama with the `all-minilm:latest` model. When using Docker Compose, this is automatically configured and the model is pulled on startup.

## Development

### Running Tests
```bash
make test              # Run all tests
go test ./...          # Or direct go test
make coverage          # Generate coverage report
```

### Development Mode
```bash
make dev              # Run with auto-reload
# Or manually:
go run ./cmd/server
```

### Monitoring Stack

When running with `docker-compose --profile monitoring up -d`:
- **Prometheus**: http://localhost:9090 (metrics collection)
- **Grafana**: http://localhost:3000 (dashboards)
  - Default login: admin/admin

## Performance

### System Performance
- **Binary Size**: ~10MB (optimized)
- **Memory Usage**: ~15MB baseline  
- **Cold Start**: ~50ms
- **Request Latency**: <1ms (95th percentile)

### Vector Storage Performance
- **Insert Performance**: 1,700+ vectors/second (production scale)
- **Search Performance**: 1.8-8.9 queries/second (depending on dataset size)
- **Dimension Support**: 3-1,536 dimensions with validation
- **Storage Engine**: Pure SQLite-vec with CGO support for maximum performance

#### Benchmark Results
| Dataset Size | Dimensions | Insert/sec | Search/sec | Total Time |
|--------------|------------|------------|------------|------------|
| 100 vectors  | 3D         | 2,355.6    | 2,307.8    | 64ms       |
| 1,000 vectors| 384D       | 2,144.0    | 8.9        | 11.7s      |
| 5,000 vectors| 384D       | 1,736.6    | 1.8        | 1m55s      |

**Quick Performance Test:**
```bash
make benchmark-quick     # Fast validation (100 vectors)
make benchmark-realistic # Production simulation (1,000 vectors)  
make benchmark-large     # Scale testing (5,000 vectors)
```

For detailed performance analysis, see [PERFORMANCE.md](./PERFORMANCE.md).

---


## � Roadmap / Next Steps

- **Vector storage**
  - [ ] Add enhanced test coverage for sqlite-vec integration
  - [ ] Document vector config, sqlite-vec setup, and logging in detail
  - [ ] Add Prometheus metrics for vector queries (counters, histograms)
- **S3 compatibility**
  - [ ] Port/restore S3 compatibility tests using new router and DI
- **General**
  - [ ] Multi-tenant support: Tenant isolation and management
  - [ ] Web UI: Dashboard for storage management
  - [ ] Enhanced backends: Additional storage drivers
  - [ ] Advanced features: Versioning, lifecycle policies


## AI/LLM Integration

8fs includes powerful AI integration capabilities for document processing and retrieval-augmented generation:

- **Llama Integration**: Complete RAG pipeline with text chunking, embedding generation, and semantic search
- **Document Processing**: Automatic text chunking and vector embedding storage
- **Semantic Search**: Vector similarity search for content discovery
- **Ready-to-Use Tools**: Command-line demo tools for quick integration

See [`docs/LLAMA_INTEGRATION.md`](./docs/LLAMA_INTEGRATION.md) for comprehensive documentation and examples.

### Quick Start with Llama Demo

```bash
# Build the Llama integration demo
make llama-demo-build

# Ingest a document
./bin/llama-demo -cmd ingest -input document.txt

# Search for similar content
./bin/llama-demo -cmd search -query "machine learning" -topk 5

# Perform RAG (Retrieval-Augmented Generation)
./bin/llama-demo -cmd rag -query "What is artificial intelligence?"
```


## Client Usage Examples

Your 8fs server is compatible with all standard S3 clients! Check out these example files:

- **`CLIENT_EXAMPLES.md`** - Comprehensive examples for Python, Node.js, cURL, AWS CLI, and more
- **`test_python_client.py`** - Ready-to-run Python boto3 example
- **`test_nodejs_client.js`** - Node.js AWS SDK v3 example  
- **`test_clients.sh`** - Automated test runner for all clients

### AWS CLI Examples

You can use AWS CLI with 8fs for common S3 operations. First, configure a profile for 8fs:

For detailed limitations and performance characteristics, refer to [AWS_CLI_INTEGRATION.md](./docs/AWS_CLI_INTEGRATION.md).

```bash
aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE --profile 8fs
aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --profile 8fs
aws configure set region us-east-1 --profile 8fs
```

Then use AWS CLI commands with the endpoint URL parameter pointing to your 8fs server:

```bash
# List buckets
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls

# Create a bucket
aws --profile 8fs --endpoint-url http://localhost:8080 s3 mb s3://my-bucket

# Upload a file
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp README.md s3://my-bucket/

# List objects in a bucket
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls s3://my-bucket

# Download a file
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp s3://my-bucket/README.md README-copy.md

# Delete a file
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rm s3://my-bucket/README.md

# Delete a bucket (must be empty first)
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rb s3://my-bucket
```

For more detailed AWS CLI integration instructions, see [AWS_CLI_INTEGRATION.md](./docs/AWS_CLI_INTEGRATION.md).

### Quick Python Example

```python
import boto3
from botocore.config import Config

# Configure for 8fs
s3 = boto3.client(
    's3',
    endpoint_url='http://localhost:8080',
    aws_access_key_id='AKIAIOSFODNN7EXAMPLE',
    aws_secret_access_key='wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
    config=Config(s3={'addressing_style': 'path'})
)

# Use like regular S3!
s3.create_bucket(Bucket='my-bucket')
s3.put_object(Bucket='my-bucket', Key='hello.txt', Body=b'Hello World!')
response = s3.get_object(Bucket='my-bucket', Key='hello.txt')
print(response['Body'].read().decode())
```

### Quick Node.js Example

```javascript
import { S3Client, CreateBucketCommand } from "@aws-sdk/client-s3";

const s3 = new S3Client({
    endpoint: "http://localhost:8080",
    forcePathStyle: true,
    credentials: {
        accessKeyId: "AKIAIOSFODNN7EXAMPLE",
        secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    }
});

await s3.send(new CreateBucketCommand({ Bucket: "my-bucket" }));
```

### Test All Clients

```bash
# Install dependencies (optional)
pip install boto3                    # Python
npm install @aws-sdk/client-s3       # Node.js

# Start server
./bin/8fs &

# Run comprehensive tests
./test_clients.sh
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run tests: `make test`
4. Build: `make build`
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
