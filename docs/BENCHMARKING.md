# 8fs Performance Benchmarking

This directory contains performance testing tools for the 8fs vector storage system.

## Quick Start

### Run Comprehensive Benchmarks
```bash
# Run comparative benchmarks across multiple configurations
make benchmark-compare

# Run single benchmark with realistic data
make benchmark-realistic

# Run quick development test
make benchmark-quick
```

### Generate Sample Data
```bash
# Generate 1000 realistic sample vectors (384D)
make generate-sample-data

# Generate clustered test data
make generate-clustered-data

# Generate random test data for quick testing
make generate-random-data
```

## Benchmark Configurations

### Small Dataset (Development)
- **Vectors**: 100-1000
- **Dimensions**: 3, 384, 768
- **Dataset Types**: Random, Clustered, Realistic
- **Purpose**: Quick validation, development testing

### Medium Dataset (Integration)
- **Vectors**: 1000-5000
- **Dimensions**: 384 (standard)
- **Dataset Types**: Clustered, Realistic
- **Purpose**: Integration testing, performance validation

### Large Dataset (Performance)
- **Vectors**: 5000+
- **Dimensions**: 384, 768, 1536
- **Dataset Types**: Realistic
- **Purpose**: Production performance evaluation

## Dataset Types

### Random
- Completely random vectors with normal distribution
- **Use case**: Stress testing, worst-case scenarios
- **Characteristics**: No structure, uniform similarity distribution

### Clustered  
- Vectors grouped into 5 distinct clusters
- **Use case**: Testing similarity search effectiveness
- **Characteristics**: Clear clustering, good/bad match scenarios

### Realistic
- Simulates text embedding patterns (OpenAI, BERT-style)
- **Use case**: Production-like performance testing
- **Characteristics**: Sparse vectors, dominant dimensions, realistic metadata

## Performance Metrics

### Insert Performance
- **Throughput**: Vectors inserted per second
- **Latency**: Average time per vector insertion
- **Success Rate**: Percentage of successful insertions

### Search Performance  
- **Throughput**: Queries processed per second
- **Latency**: Average time per query
- **Accuracy**: Relevance of returned results
- **Results**: Average number of results per query

### System Metrics
- **Database Size**: Storage requirements
- **Memory Usage**: RAM consumption during operations
- **Engine Mode**: sqlite-vec native implementation
- **Extension Loading**: sqlite-vec extension initialization status

## Benchmark Results Format

Results are saved in JSON format with the following structure:

```json
{
  "config": {
    "vector_count": 1000,
    "query_count": 100,
    "dimensions": 384,
    "dataset_type": "realistic"
  },
  "insert_metrics": {
    "total_time_ms": 2500,
    "average_time_ms": 2.5,
    "throughput_per_sec": 400.0,
    "vectors_inserted": 1000
  },
  "search_metrics": {
    "total_time_ms": 500,
    "average_time_ms": 5.0,
    "throughput_per_sec": 200.0,
    "queries_executed": 100,
    "average_results": 8.2
  },
  "overall_metrics": {
    "total_time_ms": 3000,
    "database_size_bytes": 1048576,
    "sqlite_vec_mode": true,
    "extension_loaded": true
  }
}
```

## Environment Requirements

- **Go**: 1.23+
- **SQLite**: 3.38+ with sqlite-vec extension (required)
- **Memory**: Minimum 512MB available RAM
- **Storage**: 100MB+ free space for test databases

## Command Line Usage

### Benchmark Command
```bash
./benchmark \
  -db ./data/benchmark.db \
  -output ./results.json \
  -vectors 1000 \
  -queries 100 \
  -dims 384 \
  -topk 10 \
  -dataset realistic \
  -verbose \
  -compare
```

### Parameters
- `-db`: Database file path
- `-output`: Results output file
- `-vectors`: Number of vectors to insert
- `-queries`: Number of search queries
- `-dims`: Vector dimensions
- `-topk`: Results per search
- `-dataset`: Dataset type (random/clustered/realistic)  
- `-seed`: Random seed for reproducible results
- `-verbose`: Enable progress logging
- `-cleanup`: Clean database after benchmark
- `-compare`: Run comparative benchmarks

## Performance Expectations

### SQLite-vec Mode (Recommended)
- **Insert**: 200-1000 vectors/sec
- **Search**: 50-500 queries/sec
- **Dimensions**: Up to 1536D efficiently

*Note: Performance varies significantly based on hardware, dataset characteristics, and system load.*

## Troubleshooting

### SQLite-vec Extension Not Loading
```
ERROR Extension loading failed
```
- **Cause**: sqlite-vec extension not installed or incompatible
- **Solution**: Install sqlite-vec extension with CGO support
- **Requirements**: Go 1.23+, CGO enabled, compatible system libraries

### Low Performance
```
Throughput below expected ranges
```
- **Check**: Available system memory
- **Check**: Database file location (SSD vs HDD)  
- **Check**: Concurrent system load
- **Solution**: Use `-verbose` flag for detailed progress

### Memory Issues
```
Out of memory or excessive memory usage
```
- **Solution**: Reduce vector count or dimensions
- **Solution**: Run benchmarks in smaller batches
- **Solution**: Ensure sufficient available RAM

## Integration with 8fs

These benchmarks test the actual storage backend used by 8fs:

1. **Storage Layer**: `internal/domain/vectors/sqlitevec.go`
2. **HTTP Handlers**: `internal/transport/http/handlers/vectors.go`  
3. **Vector Math**: `internal/domain/vectors/math.go`

Benchmark results directly reflect the performance users can expect from the 8fs API endpoints.
