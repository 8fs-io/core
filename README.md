# 8fs – S3-Compatible Storage Server with Native Vector Search

## 🌟 Vision

**8fs is the first S3-compatible storage server with built-in vector storage for AI developers.**

Not a MinIO clone—8fs unifies object storage and vector embeddings in a lightweight binary for local AI labs, from laptops to Raspberry Pi clusters.

> *S3 + vector storage in one <50MB binary—perfect for indie AI workflows.*

See [`VISION.md`](./VISION.md) for the full vision, market, and technical roadmap.

> **Go module path:** `github.com/8fs-io/core`

## 🆕 Recent Changes

- **Canonical Go module path:** Now using `github.com/8fs-io/core` for all imports and go.mod.
- **Vector storage subsystem:**
  - Pluggable vector storage with [sqlite-vec](https://github.com/asg017/sqlite-vec) extension for fast vector search.
  - Automatic fallback to pure Go linear search if extension is unavailable.
  - Fixed embedding dimension (384) for all vectors.
  - Dependency injection and config-driven enable/disable.
  - Structured logging and audit support.
  - Graceful shutdown and lifecycle management.
- **Codebase refactor:**
  - All root Go files now use a library package (no more package main in root).
  - Integration tests and S3 compatibility tests modernized or archived.
  - All internal imports updated to canonical org/repo path.


**8fs** is a high-performance, S3-compatible storage server built with Go, featuring clean architecture and production-ready deployment options.  
Perfect for developers who want a simple, self-hosted storage solution.

---

## 🚀 Quick Start

### Prerequisites
- Go 1.20+ 
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
go build -o bin/8fs ./cmd/server
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

---

## ✅ Features

### S3 API Compatibility
- ✅ Bucket operations (create, delete, list)
- ✅ Object operations (upload, download, delete)
- ✅ Object metadata and listing
- ✅ AWS Signature v4 authentication

### Production Ready
- 📊 **Metrics**: Prometheus integration
- 📝 **Logging**: Structured logs with audit trails
- ❤️ **Health Checks**: `/healthz` endpoint
- � **Docker**: Ready-to-deploy containers
- 🚀 **Performance**: Optimized builds

### API Endpoints
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

- **Binary Size**: ~10MB (optimized)
- **Memory Usage**: ~15MB baseline  
- **Cold Start**: ~50ms
- **Request Latency**: <1ms (95th percentile)

---


## � Roadmap / Next Steps

- **Vector storage**
  - [ ] Add explicit test coverage for fallback (extension-disabled) path
  - [ ] Optimize linear search sort (replace bubble sort with sort.Slice)
  - [ ] Document vector config, fallback, and logging in detail
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
