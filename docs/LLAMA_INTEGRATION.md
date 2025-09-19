# Llama Integration with 8fs

This document demonstrates how to integrate Large Language Models (Llama) with 8fs for document ingestion, semantic search, and Retrieval-Augmented Generation (RAG).

## Overview

The Llama integration provides:

- **Document Processing**: Automatic text chunking and embedding generation
- **Semantic Search**: Vector similarity search using Llama-generated embeddings  
- **RAG Pipeline**: Retrieval-augmented generation for question answering
- **Flexible API**: Support for local Ollama or remote Llama API endpoints

## Architecture

```text
Document → Text Chunking → Llama Embeddings → 8fs Vector Storage
                                                      ↓
Query → Llama Embedding → Vector Search → Context → RAG Response
```

## Prerequisites

### 1. Start 8fs Server

```bash
# Build and start 8fs
make build
./bin/8fs
```

### 2. Install Ollama (Optional)

For real Llama embeddings and text generation:

```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Start Ollama server
ollama serve

# Pull a model (in another terminal)
ollama pull llama2
# Or for embeddings-optimized models:
ollama pull all-minilm
```

## Quick Start

### 1. Build the Demo Tool

```bash
make llama-demo-build
# or
go build -o bin/llama-demo ./cmd/llama-demo
```

### 2. Ingest a Document

```bash
# Create a sample document
cat > example.txt << EOF
Artificial Intelligence (AI) is transforming industries through machine learning, 
deep learning, and natural language processing. Modern AI systems can understand
text, generate images, and solve complex problems across various domains.
EOF

# Ingest the document
./bin/llama-demo -cmd ingest -input example.txt -verbose
```

### 3. Search for Similar Content

```bash
# Search for AI-related content
./bin/llama-demo -cmd search -query "machine learning applications" -topk 3
```

### 4. Perform RAG (Retrieval-Augmented Generation)

```bash
# Ask a question using RAG
./bin/llama-demo -cmd rag -query "What is artificial intelligence?"
```

## Command Reference

### Ingest Command

```bash
./bin/llama-demo -cmd ingest [options]
```

**Options:**

- `-input`: Path to text file to ingest
- `-chunk`: Chunk size in words (default: 500)
- `-verbose`: Enable verbose output

**Example:**

```bash
./bin/llama-demo -cmd ingest -input document.txt -chunk 300 -verbose
```

### Search Command

```bash
./bin/llama-demo -cmd search [options]
```

**Options:**

- `-query`: Search query text
- `-topk`: Number of results to return (default: 5)
- `-verbose`: Enable verbose output

**Example:**

```bash
./bin/llama-demo -cmd search -query "neural networks" -topk 5
```

### RAG Command

```bash
./bin/llama-demo -cmd rag [options]
```

**Options:**

- `-query`: Question to answer using RAG
- `-topk`: Number of context documents (default: 5)
- `-verbose`: Enable verbose output

**Example:**

```bash
./bin/llama-demo -cmd rag -query "How does deep learning work?" -topk 3
```

## Configuration

### Server URLs

```bash
# Use custom 8fs server
./bin/llama-demo -server http://remote-8fs:8080 -cmd search -query "test"

# Use custom Ollama endpoint  
./bin/llama-demo -llama http://remote-ollama:11434 -cmd ingest -input doc.txt
```

### Models

```bash
# Use different Llama model
./bin/llama-demo -model llama2:13b -cmd rag -query "Explain quantum computing"

# Use embedding-optimized model
./bin/llama-demo -model all-minilm -cmd ingest -input document.txt
```

## API Integration

The demo tool uses 8fs's REST API endpoints:

### Store Vector Embedding

```bash
curl -X POST http://localhost:8080/api/v1/vectors/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc_chunk_1",
    "embedding": [0.1, 0.2, 0.3, ...],
    "metadata": {
      "text": "Document content...",
      "chunk_id": 1,
      "created_at": "2023-01-01T00:00:00Z"
    }
  }'
```

### Search Similar Vectors

```bash
curl -X POST http://localhost:8080/api/v1/vectors/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": [0.1, 0.2, 0.3, ...],
    "top_k": 5
  }'
```

## Programming Integration

### Go Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "fmt"
)

func storeDocument(serverURL, docID, text string, embedding []float64) error {
    reqBody := map[string]interface{}{
        "id":        docID,
        "embedding": embedding,
        "metadata": map[string]interface{}{
            "text": text,
            "created_at": time.Now(),
        },
    }
    
    bodyBytes, _ := json.Marshal(reqBody)
    resp, err := http.Post(
        serverURL+"/api/v1/vectors/embeddings",
        "application/json", 
        bytes.NewBuffer(bodyBytes),
    )
    
    return err
}
```

### Python Example

```python
import requests
import json

def store_document(server_url, doc_id, text, embedding):
    data = {
        "id": doc_id,
        "embedding": embedding,
        "metadata": {
            "text": text,
            "created_at": "2023-01-01T00:00:00Z"
        }
    }
    
    response = requests.post(
        f"{server_url}/api/v1/vectors/embeddings",
        json=data
    )
    
    return response.status_code == 201
```

## Performance Considerations

### Chunking Strategy

- **Small chunks (100-300 words)**: Better precision, more storage
- **Large chunks (500-1000 words)**: Better context, faster ingestion
- **Overlapping chunks**: Improved continuity, increased storage

### Embedding Models

- **all-minilm**: Fast, good for general text
- **llama2**: Larger, better understanding, slower
- **sentence-transformers**: Optimized for similarity

### Batch Processing

```bash
# Process multiple documents
for file in docs/*.txt; do
    ./bin/llama-demo -cmd ingest -input "$file" -verbose
done
```

## Troubleshooting

### Common Issues

1. **Server Connection Error**

   ```text
   Error: failed to store vector: failed to call API
   ```

   - Ensure 8fs server is running: `./bin/8fs`
   - Check server URL: `-server http://localhost:8080`

2. **Ollama Connection Error**

   ```text
   Error: failed to call Ollama API
   ```

   - Start Ollama: `ollama serve`
   - Pull model: `ollama pull llama2`
   - Check Ollama URL: `-llama http://localhost:11434`

3. **Empty Search Results**
   - Verify documents are ingested: check server logs
   - Try different search queries
   - Check embedding dimensions match

### Debug Mode

```bash
# Enable verbose logging
./bin/llama-demo -cmd search -query "test" -verbose

# Check 8fs server logs
tail -f server.log
```

## Advanced Usage

### Custom Embedding Pipeline

Modify `cmd/llama-demo/main.go` to:

- Use different embedding models
- Add preprocessing steps
- Implement custom chunking logic
- Add metadata enrichment

### Production Deployment

1. **Scale Ollama**: Use multiple Ollama instances
2. **Database**: Use persistent SQLite-vec storage
3. **Caching**: Cache embeddings for repeated queries
4. **Monitoring**: Add metrics and health checks

## Use Cases

### Document Q&A System

```bash
# 1. Ingest documentation
./bin/llama-demo -cmd ingest -input docs/api-reference.txt

# 2. Answer questions
./bin/llama-demo -cmd rag -query "How do I authenticate with the API?"
```

### Knowledge Base Search

```bash
# 1. Build knowledge base
for doc in knowledge-base/*.txt; do
    ./bin/llama-demo -cmd ingest -input "$doc"
done

# 2. Semantic search
./bin/llama-demo -cmd search -query "database optimization" -topk 10
```

### Content Recommendation

```bash
# 1. Index content
./bin/llama-demo -cmd ingest -input articles/

# 2. Find similar content  
./bin/llama-demo -cmd search -query "machine learning tutorial" -topk 5
```

## Next Steps

- **Scale**: Deploy with Docker and load balancers
- **Optimize**: Tune embedding models and chunk sizes
- **Monitor**: Add metrics and performance tracking
- **Extend**: Build web interface and API clients
- **Secure**: Add authentication and access controls

For more examples and advanced usage, see the [8fs documentation](../README.md).
