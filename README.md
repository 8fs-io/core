# 8fs – S3-Compatible Storage Server

**8fs** is a high-performance, S3-compatible storage server built with Go, featuring clean architecture, comprehensive monitoring, and production-ready deployment options.  
It provides a **drop-in S3 alternative** for developers, students, and startups who want a free and self-hosted storage solution.

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

## ✅ Current Features

### S3 API Compatibility
- ✅ Bucket operations (GET, PUT, DELETE)
- ✅ Object operations (GET, PUT, HEAD, DELETE)
- ✅ Object listing with pagination
- ✅ AWS Signature v4 authentication
- ✅ Standard S3 HTTP status codes and headers

### Observability
- 📊 **Prometheus Metrics**: HTTP requests, storage usage, authentication events
- 📝 **Structured Logging**: Request/response logging with correlation IDs
- 🔍 **Audit Trail**: Complete event tracking for compliance
- ❤️ **Health Checks**: `/healthz` endpoint for monitoring

### Production Features
- 🔒 **Security**: AWS-compatible authentication and authorization
- 🚀 **Performance**: Optimized binary builds with stripped symbols
- 🐳 **Deployment**: Docker support with multi-stage builds
- 📦 **Portability**: Cross-platform builds (Linux, macOS, Windows)
- 🔄 **Graceful Shutdown**: Proper signal handling and cleanup

### API Endpoints
- `GET /healthz` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /` - List buckets
- `PUT /:bucket` - Create bucket
- `DELETE /:bucket` - Delete bucket
- `GET /:bucket` - List objects in bucket
- `PUT /:bucket/:key` - Store object
- `GET /:bucket/:key` - Retrieve object
- `HEAD /:bucket/:key` - Get object metadata
- `DELETE /:bucket/:key` - Delete object

---

## 🏗️ Architecture

This project follows clean architecture principles with clear separation of concerns:

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
├── legacy/                     # Archived legacy code
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
# Run all tests
make test
# or
go test ./...

# Run integration tests specifically  
go test -v -run TestIntegration

# Generate coverage report
make coverage
```

### Development Mode
```bash
# Auto-reload on changes (requires air)
make dev

# Or manually with default config
DEFAULT_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE \
DEFAULT_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
go run ./cmd/server
```

### Monitoring Stack

When running with `docker-compose --profile monitoring up -d`:
- **Prometheus**: http://localhost:9090 (metrics collection)
- **Grafana**: http://localhost:3000 (dashboards)
  - Default login: admin/admin

## Migration from Legacy

The legacy monolithic code has been archived in the `legacy/` directory. The new architecture provides:
- Better testability and maintainability
- Improved error handling and logging
- Enhanced monitoring and observability
- Production-ready deployment options

## Performance

Benchmarks on MacBook Pro M1:
- Cold start: ~50ms
- Binary size: ~10MB (optimized)
- Memory usage: ~15MB baseline
- Request latency: <1ms (95th percentile)

---

## 🚀 Future Roadmap

- **Multi-tenant support**: Tenant isolation and management
- **Web UI**: Dashboard for storage management
- **Advanced features**: Versioning, lifecycle policies
- **Additional backends**: Garage, local filesystem optimizations
- **Enhanced security**: OAuth/OIDC integration

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
