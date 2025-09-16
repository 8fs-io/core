# SQLite-Vec Implementation

This document describes the SQLite-vec integration for 8fs vector storage.

## Overview

The SQLite-vec implementation provides high-performance vector similarity search using the sqlite-vec extension. It includes graceful fallback to pure Go linear search when the extension is not available.

## Architecture

### SQLiteVecStorage

The main storage backend that provides:

- **Vector Storage**: Store embeddings with metadata in SQLite database
- **Similarity Search**: Optimized vector search using sqlite-vec or fallback to linear search  
- **Metadata Support**: Full JSON metadata storage and retrieval
- **Graceful Degradation**: Automatic fallback when sqlite-vec extension is not available

### Key Components

1. **Storage Layer** (`sqlitevec.go`)
   - Database connection management
   - Schema initialization
   - Vector serialization/deserialization
   - Search operations

2. **Testing Suite** (`sqlitevec_test.go`) 
   - Basic storage and retrieval tests
   - Search functionality validation
   - Performance benchmarks

## Usage

### Basic Usage

```go
// Create storage instance
storage, err := NewSQLiteVecStorage("vectors.db")
if err != nil {
    log.Fatal(err)
}
defer storage.Close()

// Store a vector
vector := &Vector{
    ID:        "doc_123",
    Embedding: []float64{0.1, 0.2, 0.3, ...}, // 384-1536 dimensions
    Metadata:  map[string]interface{}{
        "document": "example.txt",
        "chunk": 1,
    },
}

err = storage.Store(vector)
if err != nil {
    log.Fatal(err)
}

// Search for similar vectors
query := []float64{0.1, 0.2, 0.3, ...}
results, err := storage.Search(query, 10) // Top 10 results
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("ID: %s, Score: %.4f\n", result.Vector.ID, result.Score)
}
```

## Implementation Details

### Schema

The implementation supports two schemas:

1. **SQLite-vec Virtual Table** (when extension is available):
```sql
CREATE VIRTUAL TABLE embeddings USING vec0(
    id TEXT PRIMARY KEY,
    embedding FLOAT[384],
    metadata TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
```

2. **Fallback Table** (when extension is not available):
```sql
CREATE TABLE embeddings (
    id TEXT PRIMARY KEY,
    embedding BLOB,
    metadata TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
```

### Serialization

- **Embeddings**: JSON serialization for compatibility
- **Metadata**: JSON serialization for structured data support
- **Future**: Binary format optimization for production sqlite-vec

### Search Modes

1. **Optimized Search** (with sqlite-vec):
   - Uses vector index for fast similarity search
   - Sub-millisecond search times for large datasets
   - Supports various distance metrics

2. **Linear Search** (fallback):
   - Pure Go implementation using existing VectorMath
   - Cosine similarity calculation
   - In-memory result sorting

## Performance

### Expected Performance (with sqlite-vec extension)

| Dataset Size | Search Time | Memory Usage |
|-------------|-------------|--------------|
| 1K vectors  | ~0.1ms      | Low          |
| 10K vectors | ~0.5ms      | Moderate     |
| 100K vectors| ~2ms        | Higher       |

### Fallback Performance (without extension)

| Dataset Size | Search Time | Memory Usage |
|-------------|-------------|--------------|
| 1K vectors  | ~2ms        | Low          |
| 10K vectors | ~20ms       | Moderate     |
| 100K vectors| ~200ms      | Higher       |

## Installation & Setup

### SQLite-vec Extension

To use the optimized sqlite-vec features, you'll need to install the sqlite-vec extension:

```bash
# Option 1: Build from source
git clone https://github.com/asg017/sqlite-vec.git
cd sqlite-vec
make install

# Option 2: Use pre-built binaries (when available)
# Check releases at https://github.com/asg017/sqlite-vec/releases
```

### Go Dependencies

The implementation requires:
- `github.com/mattn/go-sqlite3` - SQLite driver with CGO support
- Standard library packages (json, fmt, database/sql)

## Testing

Run the test suite:

```bash
# Basic functionality tests
go test ./internal/domain/vectors -run TestSQLiteVecStorage -v

# Performance benchmarks
go test ./internal/domain/vectors -bench BenchmarkSQLiteVec -v
```

## Future Enhancements

1. **Binary Serialization**: Optimize for sqlite-vec binary format
2. **Index Configuration**: Support different vector indexes (HNSW, IVF, etc.)
3. **Batch Operations**: Bulk insert/update operations  
4. **Metadata Filtering**: Combined vector + metadata queries
5. **Connection Pooling**: Multi-connection support for high throughput

## Error Handling

The implementation includes comprehensive error handling:

- Database connection failures
- Extension loading issues (graceful fallback)
- Serialization/deserialization errors
- Invalid vector dimensions
- Search operation failures

## Compatibility

- **Go Version**: 1.21+
- **SQLite Version**: 3.38+ (for extension support)
- **CGO**: Required for sqlite3 driver
- **Platforms**: Linux, macOS, Windows (with CGO)
