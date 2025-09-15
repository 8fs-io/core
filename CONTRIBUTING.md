# Contributing to 8fs

Thank you for your interest in contributing to 8fs! We're building the first S3-compatible storage server with built-in vector storage for AI developers.

## 🎯 Project Mission

8fs aims to be the "SQLite of AI storage" - simple, lightweight, and perfect for local AI development. We're focused on:
- Unified storage for models (S3) and embeddings (vectors) in one binary
- Raspberry Pi and edge deployment optimization  
- Zero-config setup with no external dependencies
- Developer-friendly APIs for AI workflows

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or later
- SQLite3 (for development)
- Docker (optional, for testing)

### Local Development Setup

1. **Clone and build**:
```bash
git clone https://github.com/8fs-io/core.git
cd core
go mod download
go build -o 8fs ./cmd/server
```

2. **Run locally**:
```bash
./8fs
# Server starts on http://localhost:8080
```

3. **Test S3 API**:
```bash
# Check health
curl http://localhost:8080/healthz

# List buckets  
curl http://localhost:8080/

# Create bucket
curl -X PUT http://localhost:8080/test-bucket
```

4. **Test Vector API** (v0.1+):
```bash
# Store embedding
curl -X POST http://localhost:8080/vectors/embeddings \
  -H "Content-Type: application/json" \
  -d '{"id": "doc1", "embedding": [0.1, 0.2, 0.3], "metadata": {"type": "test"}}'

# Search similar
curl -X POST http://localhost:8080/vectors/search \
  -H "Content-Type: application/json" \
  -d '{"query": [0.1, 0.2, 0.3], "top_k": 5}'
```

## 🔧 How to Contribute

### Reporting Issues
- **Bug Reports**: Use the "Bug Report" template with reproduction steps
- **Feature Requests**: Use the "Feature Request" template with use case details
- **Questions**: Use GitHub Discussions for general questions

### Code Contributions

#### 1. Find an Issue
- Look for issues labeled `good first issue` for newcomers
- Check `help wanted` for items needing attention
- Propose new features in GitHub Discussions first

#### 2. Development Workflow
```bash
# 1. Fork the repository on GitHub
# 2. Clone your fork
git clone https://github.com/YOUR-USERNAME/core.git
cd core

# 3. Create a feature branch
git checkout -b feat/your-feature-name

# 4. Make your changes
# 5. Test your changes
go test ./...

# 6. Commit with conventional commits
git commit -m "feat: add vector search caching"

# 7. Push and create PR
git push origin feat/your-feature-name
```

#### 3. Pull Request Guidelines
- **Title**: Use conventional commits format (`feat:`, `fix:`, `docs:`, etc.)
- **Description**: Explain what and why, not just how
- **Tests**: Add tests for new functionality
- **Documentation**: Update README/docs if needed
- **Small PRs**: Keep changes focused and reviewable

### Code Style

#### Go Code Standards
- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for public APIs
- Keep functions small and focused
- Handle errors explicitly

#### Project Structure
```
8fs/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── container/      # Dependency injection
│   ├── domain/         # Business logic and models
│   ├── infrastructure/ # External dependencies (storage, etc.)
│   └── transport/      # HTTP handlers and routing
├── pkg/                # Public packages
└── docs/               # Documentation
```

#### Commit Message Format
We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

**Examples**:
- `feat: add vector similarity search`
- `fix: handle empty embedding arrays`
- `docs: update vector API examples`
- `test: add integration tests for S3 operations`

## 🎯 Areas Needing Help

### Good First Issues
- [ ] Add WASM build configuration
- [ ] Improve vector search performance benchmarks
- [ ] Add Docker health check for vector endpoints
- [ ] Write vector API integration tests
- [ ] Add Prometheus metrics for vector operations

### Priority Features (v0.1)
- [ ] Vector storage implementation (`/vectors/embeddings`, `/vectors/search`)
- [ ] SQLite-based vector indexing
- [ ] Python client examples
- [ ] Raspberry Pi performance optimization

### Future Features (v0.2+)
- [ ] WASM browser build
- [ ] Web UI for storage management
- [ ] P2P synchronization for edge deployments
- [ ] Advanced vector search (HNSW indexing)

## 🧪 Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/domain/storage/

# Run integration tests
go test -tags=integration ./...
```

### Test Organization
- **Unit tests**: `*_test.go` files alongside source code
- **Integration tests**: `integration_test.go` with build tags
- **Benchmarks**: `*_bench_test.go` for performance testing

### Adding Tests
- Cover happy path and error cases
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test public APIs thoroughly

## 📚 Documentation

### Code Documentation
- Document all public functions and types
- Use Go doc conventions
- Include usage examples in doc comments
- Keep comments up to date with code changes

### User Documentation  
- Update README.md for new features
- Add API examples that work
- Include common use cases and troubleshooting
- Keep installation instructions current

## 🏗️ Release Process

### Version Numbering
We use [Semantic Versioning](https://semver.org/):
- **v0.1.x**: MVP with basic S3 + vector storage
- **v0.2.x**: WASM build and web UI  
- **v0.3.x**: P2P synchronization and clustering

### Release Workflow
1. Feature development on feature branches
2. PR review and testing
3. Merge to `main` branch
4. Tag release with version number
5. Automated Docker image build
6. Release notes and changelog update

## 💬 Community

### Communication
- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: Questions, ideas, general discussion
- **Discord**: Real-time chat and community support
- **X/Twitter**: [@8fs_storage](https://twitter.com/8fs_storage) for updates

### Code of Conduct
Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md). We're committed to providing a welcoming and inclusive environment for all contributors.

## 🤔 Questions?

- Check existing GitHub Issues and Discussions
- Read the [README.md](README.md) for basic usage
- Ask questions in GitHub Discussions
- Join our Discord community

## 🙏 Recognition

Contributors will be:
- Listed in our README.md contributors section
- Mentioned in release notes for significant contributions
- Invited to join our Discord community
- Credited in any talks or blog posts about 8fs

Thank you for helping make AI storage simple and accessible! 🚀

---

**Happy Contributing!** Every contribution, no matter how small, helps make 8fs better for the AI developer community.
