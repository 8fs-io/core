# 8fs – S3-Compatible Storage Server with Native Vector Search

## 🌟 Vision

**8fs is the first S3-compatible storage server with built-in vector storage for AI developers.**

Not a MinIO clone—8fs unifies object storage and vector embeddings in one lightweight binary for local AI labs, from laptops to Raspberry Pi clusters.

> *S3 + vector storage in one <50MB binary—perfect for indie AI workflows.*

See [`VISION.md`](./VISION.md) for the full vision, market, and technical roadmap.

> **Go module path:** `github.com/8fs-io/core`

## 🆕 Recent Changes

- **Pure SQLite-vec Vector Storage:**
  - **BREAKING**: Removed fallback implementation for simplified, production-focused architecture
  - Pure sqlite-vec implementation with CGO support for maximum performance
  - Advanced dimension handling (3-1536) with intelligent validation
  - **Performance**: 1,700+ vec/sec insert, 8.9 search/sec at production scale
- **Comprehensive Performance Suite:**
  - Published benchmark results: [PERFORMANCE.md](./PERFORMANCE.md)
  - Multiple dataset generators: random, clustered, realistic patterns
  - Automated benchmarking tools via `make benchmark-*` targets
  - Production-validated with up to 5,000 vectors at 384 dimensions
- **Developer Experience:**
  - Enhanced error handling with proper HTTP status codes for dimension mismatches
  - Clean package structure without naming conflicts
  - Comprehensive test coverage including integration and performance tests

## 🗺 Roadmap / Next Steps

- **Vector storage**
  - [x] Pure SQLite-vec implementation with CGO support
  - [x] Production-scale performance validation (1,700+ vec/sec)
  - [x] Comprehensive benchmarking suite with published metrics
  - [x] Advanced error handling and dimension validation
  - [ ] SIMD acceleration for vector math operations  
  - [ ] Parallel search execution for multi-query workloads
  - [ ] Advanced indexing strategies for >10K vector datasets
- **S3 compatibility**
  - [ ] Restore and enhance S3 compatibility test suite
  - [ ] AWS CLI integration testing and documentation
- **General**
  - [ ] Multi-tenant support: Tenant isolation and management
  - [ ] Web UI: Dashboard for storage and vector management
  - [ ] Enhanced backends: Additional storage drivers beyond filesystem
  - [ ] Advanced features: Versioning, lifecycle policies, metadata search

---

**8fs** is a high-performance, S3-compatible storage server built with Go, featuring clean architecture and production-ready deployment options.  
Perfect for developers who want a simple, self-hosted storage solution.

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
# Start basic services
docker-compose up -d

# Start with monitoring (Prometheus + Grafana)
docker-compose --profile monitoring up -d
```

### Vector Storage Quick Start

Once your server is running, you can start using vector storage:

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

### S3 API Compatibility
- ✅ Bucket operations (create, delete, list)
- ✅ Object operations (upload, download, delete)
- ✅ Object metadata and listing
- ✅ AWS Signature v4 authentication

### Vector Storage API
- ✅ Embedding storage (`POST /api/v1/vectors/embeddings`)
- ✅ Cosine similarity search (`POST /api/v1/vectors/search`)
- ✅ Vector retrieval (`GET /api/v1/vectors/embeddings/:id`)
- ✅ Vector deletion (`DELETE /api/v1/vectors/embeddings/:id`)
- ✅ Vector listing (`GET /api/v1/vectors/embeddings`)
- ✅ Dimension validation (3-1,536 dimensions)
- ✅ Metadata filtering and search

### Production Ready
- 📊 **Metrics**: Prometheus integration
- 📝 **Logging**: Structured logs with audit trails
- ❤️ **Health Checks**: `/healthz` endpoint
- � **Docker**: Ready-to-deploy containers
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
- `DEFAULT_ACCESS_KEY`: AWS access key (default: `AKIAIOSFODNN7EXAMPLE`)
- `DEFAULT_SECRET_KEY`: AWS secret key (default: `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`)
- `PORT`: Server port (default: `8080`)
- `STORAGE_PATH`: Storage directory (default: `./storage`)
- `GIN_MODE`: Gin mode (`debug`, `release`) (default: `debug`)

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
