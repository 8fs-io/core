# 8fs RAG (Retrieval-Augmented Generation) Server

8fs now includes comprehensive RAG capabilities, turning it into a complete AI-native storage and generation platform. This combines S3-compatible object storage, vector search, and text generation into a single binary.

## ✅ RAG Features

### 🤖 OpenAI-Compatible Chat API
- **`POST /api/v1/chat/completions`** - OpenAI-compatible chat completions with RAG
- **Automatic Context Retrieval**: Query → Embedding → Vector Search → Context Formation → Generation
- **Multi-Provider Support**: Ollama (Llama), OpenAI (GPT), AWS Bedrock (Claude)
- **Streaming Support**: Real-time response streaming (coming soon)

### 🔍 Context Management
- **`POST /api/v1/chat/search/context`** - Search and retrieve relevant context documents
- **`POST /api/v1/chat/generate/context`** - Generate text with pre-provided context
- **Smart Relevance Filtering**: Configurable minimum relevance scores
- **Metadata Enrichment**: Source tracking, chunking information, and custom metadata

### ⚡ Performance Optimized
- **Vector Search Backend**: SQLite-vec with 1,700+ vectors/sec performance
- **Concurrent Processing**: Async embedding generation and retrieval
- **Caching Ready**: Built for Redis/memory caching integration
- **Token Optimization**: Efficient context window management

## 🚀 Quick Start

### 1. Start 8fs with RAG enabled

```bash
# Build with RAG support
./build.sh

# Run with default config (enables vectors + RAG)
./bin/8fs
```

### 2. Store documents with embeddings

```bash
# Store a document - embeddings are auto-generated
curl -X POST http://localhost:8080/api/v1/vectors/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc_1",
    "embedding": [],
    "metadata": {
      "content": "8fs is an S3-compatible storage server with native vector search and RAG capabilities.",
      "source": "overview.md",
      "chunk": 1
    }
  }'
```

### 3. Perform RAG-based chat completion

```bash
# OpenAI-compatible RAG chat
curl -X POST http://localhost:8080/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is 8fs and what makes it unique?",
    "max_tokens": 500,
    "temperature": 0.7,
    "top_k": 5
  }'
```

### 4. Search for context only

```bash
# Retrieve relevant context without generation
curl -X POST http://localhost:8080/api/v1/chat/search/context \
  -H "Content-Type: application/json" \
  -d '{
    "query": "performance characteristics",
    "top_k": 10
  }'
```

## 📊 API Reference

### Chat Completions (OpenAI Compatible)

**POST** `/api/v1/chat/completions`

```json
{
  "query": "Your question here",
  "max_tokens": 4000,
  "temperature": 0.7,
  "top_k": 5,
  "metadata": {},
  "stream": false
}
```

**Response:**
```json
{
  "id": "rag_1234567890",
  "object": "chat.completion",
  "created": 1701234567,
  "model": "llama3.2:latest",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Based on the provided context..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 150,
    "completion_tokens": 200,
    "total_tokens": 350
  },
  "context": {
    "documents": [...],
    "total_found": 5,
    "query_time": "15ms"
  },
  "process_time": "1.2s"
}
```

### Context Search

**POST** `/api/v1/chat/search/context`

```json
{
  "query": "search term",
  "top_k": 10,
  "metadata": {}
}
```

**Response:**
```json
{
  "documents": [
    {
      "id": "doc_1",
      "content": "Document content...",
      "metadata": {...},
      "score": 0.95,
      "source": "file.pdf",
      "chunk": 1
    }
  ],
  "total_found": 10,
  "query_time": "5ms"
}
```

### Direct Generation with Context

**POST** `/api/v1/chat/generate/context`

```json
{
  "prompt": "User question",
  "context": "Provided context text...",
  "max_tokens": 500,
  "temperature": 0.7,
  "system_msg": "Custom system prompt"
}
```

## ⚙️ Configuration

### Environment Variables

```bash
# AI Provider Settings
AI_BASE_URL=http://localhost:11434     # Ollama endpoint
AI_EMBEDDING_MODEL=all-minilm:latest   # Embedding model
AI_CHAT_MODEL=llama3.2:latest         # Chat/generation model
AI_TIMEOUT=30s

# RAG Settings
RAG_DEFAULT_TOP_K=5                    # Default context documents to retrieve
RAG_DEFAULT_MAX_TOKENS=4000           # Default max tokens for generation
RAG_DEFAULT_TEMPERATURE=0.7           # Default generation temperature
RAG_CONTEXT_WINDOW_SIZE=8000          # Max context window size
RAG_MIN_RELEVANCE_SCORE=0.1           # Minimum relevance score for documents

# Vector Storage
VECTOR_ENABLED=true
VECTOR_DB_PATH=./vectors.db
VECTOR_DIMENSION=384
```

### Config File (config.yaml)

```yaml
ai:
  base_url: "http://localhost:11434"
  embedding_model: "all-minilm:latest"
  chat_model: "llama3.2:latest"
  timeout: "30s"

vector:
  enabled: true
  db_path: "./vectors.db"
  dimension: 384

rag:
  default_top_k: 5
  default_max_tokens: 4000
  default_temperature: 0.7
  context_window_size: 8000
  min_relevance_score: 0.1
  system_prompt: "You are a helpful AI assistant..."
```

## 🔧 AI Provider Setup

### Ollama (Recommended for Local Development)

```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull required models
ollama pull all-minilm:latest      # For embeddings (384-dim)
ollama pull llama3.2:latest        # For text generation
```

### OpenAI (Production)

```bash
# Set API key
export OPENAI_API_KEY=your-api-key

# Configure in environment
AI_PROVIDER=openai
AI_EMBEDDING_MODEL=text-embedding-ada-002
AI_CHAT_MODEL=gpt-4
```

### AWS Bedrock (Enterprise)

```bash
# Set AWS credentials
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-west-2

# Configure in environment  
AI_PROVIDER=bedrock
AI_EMBEDDING_MODEL=amazon.titan-embed-text-v1
AI_CHAT_MODEL=anthropic.claude-v2
```

## 📈 Performance & Monitoring

### Health Checks

```bash
# RAG service health
curl http://localhost:8080/api/v1/chat/health

# Vector storage health
curl http://localhost:8080/api/v1/vectors/health
```

### Metrics

- **Context Retrieval Time**: Vector search performance
- **Generation Time**: AI model response time  
- **Total RAG Time**: End-to-end request processing
- **Token Usage**: Prompt and completion token consumption
- **Cache Hit Rates**: Context and embedding cache performance

### Expected Performance

- **Context Retrieval**: 5-50ms (depending on corpus size)
- **Text Generation**: 1-5s (depending on model and tokens)
- **Total RAG Time**: 1-10s end-to-end
- **Throughput**: 5-50 concurrent RAG requests (depending on AI provider)

## 🧪 Testing

### Python Test Client

```bash
# Run comprehensive RAG tests
python3 test_rag_client.py
```

### curl Examples

```bash
# Store sample document
./scripts/test_rag.sh store

# Test RAG completion
./scripts/test_rag.sh chat

# Search context only  
./scripts/test_rag.sh search
```

## 🛠️ Development

### Adding New AI Providers

1. Extend `internal/domain/ai/service.go` with provider-specific implementation
2. Add provider configuration to `internal/config/config.go`  
3. Update container initialization in `internal/container/container.go`
4. Add provider-specific request/response handling

### Custom RAG Pipelines

1. Implement `rag.Service` interface with custom logic
2. Register in container with custom configuration
3. Add custom endpoints in `internal/transport/http/handlers/rag.go`

### Vector Storage Backends

1. Implement vector storage interface for new backends (Pinecone, Weaviate, etc.)
2. Update container to support multiple vector storage options
3. Add provider-specific configuration

## 🚧 Roadmap

- [ ] **Streaming Responses**: Server-sent events for real-time RAG
- [ ] **Conversation Memory**: Multi-turn conversation context management
- [ ] **Hybrid Search**: Semantic + keyword search combination
- [ ] **RAG Evaluation**: Quality metrics and A/B testing framework
- [ ] **Web UI Dashboard**: Visual RAG management and testing interface
- [ ] **Advanced Chunking**: PDF, DOCX, and semantic boundary detection
- [ ] **Multi-Modal RAG**: Image and document understanding capabilities

## 📚 Examples

See `/client_examples/` directory for:
- Python RAG client (`rag_client.py`)
- Node.js integration (`rag_client.js`)
- Curl scripts (`test_rag.sh`)
- Jupyter notebooks (`rag_examples.ipynb`)

## 🤝 Contributing

RAG development follows the same contribution guidelines as 8fs core. See `CONTRIBUTING.md` for details on:
- Code style and testing requirements
- RAG-specific testing procedures
- Performance benchmarking standards
- Documentation requirements