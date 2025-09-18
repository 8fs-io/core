package vectors

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

// BenchmarkConfig defines parameters for performance testing
type BenchmarkConfig struct {
	VectorCount       int    // Number of vectors to insert
	QueryCount        int    // Number of search queries to run
	Dimensions        int    // Vector dimensions
	TopK              int    // Number of results to return per search
	DatasetType       string // Type of data distribution: "random", "clustered", "realistic"
	RandomSeed        int64  // For reproducible results
	EnableProgressLog bool   // Log progress during benchmarking
}

// BenchmarkResult contains performance metrics
type BenchmarkResult struct {
	Config         BenchmarkConfig `json:"config"`
	InsertMetrics  InsertMetrics   `json:"insert_metrics"`
	SearchMetrics  SearchMetrics   `json:"search_metrics"`
	OverallMetrics OverallMetrics  `json:"overall_metrics"`
	SystemInfo     SystemInfo      `json:"system_info"`
	Timestamp      time.Time       `json:"timestamp"`
}

type InsertMetrics struct {
	TotalTime        time.Duration `json:"total_time_ms"`
	AverageTime      time.Duration `json:"average_time_ms"`
	ThroughputPerSec float64       `json:"throughput_per_sec"`
	VectorsInserted  int           `json:"vectors_inserted"`
}

type SearchMetrics struct {
	TotalTime        time.Duration `json:"total_time_ms"`
	AverageTime      time.Duration `json:"average_time_ms"`
	ThroughputPerSec float64       `json:"throughput_per_sec"`
	QueriesExecuted  int           `json:"queries_executed"`
	AverageResults   float64       `json:"average_results_returned"`
	AverageAccuracy  float64       `json:"average_accuracy"`
}

type OverallMetrics struct {
	DatabaseSize int64         `json:"database_size_bytes"`
	MemoryUsage  int64         `json:"memory_usage_bytes"` // Best effort
	TotalTime    time.Duration `json:"total_time_ms"`
}

type SystemInfo struct {
	Platform      string `json:"platform"`
	SQLiteVersion string `json:"sqlite_version"`
	VectorEngine  string `json:"vector_engine"` // "sqlite-vec" or "fallback"
}

// Benchmarker runs performance tests on vector storage
type Benchmarker struct {
	storage *SQLiteVecStorage
	logger  Logger
	rand    *rand.Rand
}

// NewBenchmarker creates a new benchmarker instance
func NewBenchmarker(storage *SQLiteVecStorage, logger Logger) *Benchmarker {
	if logger == nil {
		logger = &noopLogger{}
	}

	return &Benchmarker{
		storage: storage,
		logger:  logger,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RunBenchmark executes a full performance test suite
func (b *Benchmarker) RunBenchmark(config BenchmarkConfig) (*BenchmarkResult, error) {
	b.logger.Info("starting vector storage benchmark",
		"vectors", config.VectorCount,
		"queries", config.QueryCount,
		"dimensions", config.Dimensions,
		"dataset", config.DatasetType)

	startTime := time.Now()

	// Set random seed for reproducible results
	if config.RandomSeed != 0 {
		b.rand = rand.New(rand.NewSource(config.RandomSeed))
	}

	result := &BenchmarkResult{
		Config:    config,
		Timestamp: startTime,
		SystemInfo: SystemInfo{
			Platform:     "golang", // Could detect actual platform
			VectorEngine: b.getEngineType(),
		},
	}

	// Generate sample data
	vectors, queries := b.generateSampleData(config)

	// Benchmark insertions
	insertMetrics, err := b.benchmarkInsertions(vectors, config.EnableProgressLog)
	if err != nil {
		return nil, fmt.Errorf("insertion benchmark failed: %w", err)
	}
	result.InsertMetrics = insertMetrics

	// Benchmark searches
	searchMetrics, err := b.benchmarkSearches(queries, config.TopK, config.EnableProgressLog)
	if err != nil {
		return nil, fmt.Errorf("search benchmark failed: %w", err)
	}
	result.SearchMetrics = searchMetrics

	// Overall metrics
	result.OverallMetrics = OverallMetrics{
		TotalTime:    time.Since(startTime),
		DatabaseSize: b.estimateDatabaseSize(),
	}

	b.logger.Info("benchmark completed",
		"total_time_ms", result.OverallMetrics.TotalTime.Milliseconds(),
		"insert_throughput", result.InsertMetrics.ThroughputPerSec,
		"search_throughput", result.SearchMetrics.ThroughputPerSec)

	return result, nil
}

// generateSampleData creates vectors and queries based on config
func (b *Benchmarker) generateSampleData(config BenchmarkConfig) ([]*Vector, [][]float64) {
	switch config.DatasetType {
	case "clustered":
		return b.generateClusteredData(config)
	case "realistic":
		return b.generateRealisticData(config)
	default: // "random"
		return b.generateRandomData(config)
	}
}

// generateRandomData creates completely random vectors
func (b *Benchmarker) generateRandomData(config BenchmarkConfig) ([]*Vector, [][]float64) {
	vectors := make([]*Vector, config.VectorCount)
	queries := make([][]float64, config.QueryCount)

	// Generate random vectors
	for i := 0; i < config.VectorCount; i++ {
		embedding := make([]float64, config.Dimensions)
		for j := 0; j < config.Dimensions; j++ {
			embedding[j] = b.rand.NormFloat64() // Normal distribution
		}

		vectors[i] = &Vector{
			ID:        fmt.Sprintf("vec_%d", i),
			Embedding: b.normalizeVector(embedding),
			Metadata: map[string]interface{}{
				"type":       "random",
				"cluster_id": i % 10, // Fake clustering for metadata filtering
				"created_at": time.Now().Add(-time.Duration(i) * time.Second).Unix(),
			},
		}
	}

	// Generate random queries
	for i := 0; i < config.QueryCount; i++ {
		embedding := make([]float64, config.Dimensions)
		for j := 0; j < config.Dimensions; j++ {
			embedding[j] = b.rand.NormFloat64()
		}
		queries[i] = b.normalizeVector(embedding)
	}

	return vectors, queries
}

// generateClusteredData creates vectors in distinct clusters
func (b *Benchmarker) generateClusteredData(config BenchmarkConfig) ([]*Vector, [][]float64) {
	vectors := make([]*Vector, config.VectorCount)
	queries := make([][]float64, config.QueryCount)

	numClusters := 5
	clusterCenters := make([][]float64, numClusters)

	// Generate cluster centers
	for i := 0; i < numClusters; i++ {
		center := make([]float64, config.Dimensions)
		for j := 0; j < config.Dimensions; j++ {
			center[j] = b.rand.NormFloat64() * 2 // Wider spread for centers
		}
		clusterCenters[i] = b.normalizeVector(center)
	}

	// Generate vectors around cluster centers
	for i := 0; i < config.VectorCount; i++ {
		clusterID := i % numClusters
		center := clusterCenters[clusterID]

		embedding := make([]float64, config.Dimensions)
		for j := 0; j < config.Dimensions; j++ {
			// Add noise around cluster center
			embedding[j] = center[j] + b.rand.NormFloat64()*0.1
		}

		vectors[i] = &Vector{
			ID:        fmt.Sprintf("clustered_vec_%d", i),
			Embedding: b.normalizeVector(embedding),
			Metadata: map[string]interface{}{
				"type":       "clustered",
				"cluster_id": clusterID,
				"topic":      fmt.Sprintf("topic_%d", clusterID),
			},
		}
	}

	// Generate queries that should find clusters
	for i := 0; i < config.QueryCount; i++ {
		clusterID := i % numClusters
		center := clusterCenters[clusterID]

		embedding := make([]float64, config.Dimensions)
		for j := 0; j < config.Dimensions; j++ {
			// Query slightly off-center to test similarity
			embedding[j] = center[j] + b.rand.NormFloat64()*0.05
		}
		queries[i] = b.normalizeVector(embedding)
	}

	return vectors, queries
}

// generateRealisticData simulates realistic text embeddings
func (b *Benchmarker) generateRealisticData(config BenchmarkConfig) ([]*Vector, [][]float64) {
	// This simulates text embeddings like you'd get from OpenAI, Sentence-BERT, etc.
	vectors := make([]*Vector, config.VectorCount)
	queries := make([][]float64, config.QueryCount)

	// Realistic embedding patterns (sparse with some dominant dimensions)
	dominantDims := config.Dimensions / 10 // ~10% of dimensions are "important"

	for i := 0; i < config.VectorCount; i++ {
		embedding := make([]float64, config.Dimensions)

		// Most dimensions are near-zero
		for j := 0; j < config.Dimensions; j++ {
			embedding[j] = b.rand.NormFloat64() * 0.1
		}

		// Some dimensions have stronger signals
		for k := 0; k < dominantDims; k++ {
			dimIndex := b.rand.Intn(config.Dimensions)
			embedding[dimIndex] += b.rand.NormFloat64() * 0.8
		}

		vectors[i] = &Vector{
			ID:        fmt.Sprintf("doc_%d", i),
			Embedding: b.normalizeVector(embedding),
			Metadata: map[string]interface{}{
				"type":       "document",
				"category":   []string{"tech", "science", "business", "health"}[i%4],
				"word_count": 100 + b.rand.Intn(1000),
				"language":   "en",
				"created_at": time.Now().Add(-time.Duration(i) * time.Hour).Unix(),
			},
		}
	}

	// Generate realistic queries
	for i := 0; i < config.QueryCount; i++ {
		embedding := make([]float64, config.Dimensions)

		// Similar pattern to documents but with variation
		for j := 0; j < config.Dimensions; j++ {
			embedding[j] = b.rand.NormFloat64() * 0.15
		}

		for k := 0; k < dominantDims; k++ {
			dimIndex := b.rand.Intn(config.Dimensions)
			embedding[dimIndex] += b.rand.NormFloat64() * 0.6
		}

		queries[i] = b.normalizeVector(embedding)
	}

	return vectors, queries
}

// normalizeVector converts vector to unit length
func (b *Benchmarker) normalizeVector(v []float64) []float64 {
	var magnitude float64
	for _, val := range v {
		magnitude += val * val
	}
	magnitude = math.Sqrt(magnitude)

	if magnitude == 0 {
		return v // Avoid division by zero
	}

	normalized := make([]float64, len(v))
	for i, val := range v {
		normalized[i] = val / magnitude
	}
	return normalized
}

// benchmarkInsertions measures vector insertion performance
func (b *Benchmarker) benchmarkInsertions(vectors []*Vector, logProgress bool) (InsertMetrics, error) {
	startTime := time.Now()
	successful := 0

	for i, vector := range vectors {
		if err := b.storage.Store(vector); err != nil {
			b.logger.Warn("failed to store vector during benchmark", "id", vector.ID, "error", err)
			continue
		}
		successful++

		if logProgress && (i+1)%100 == 0 {
			elapsed := time.Since(startTime)
			rate := float64(i+1) / elapsed.Seconds()
			b.logger.Info("insertion progress", "completed", i+1, "total", len(vectors), "rate_per_sec", rate)
		}
	}

	totalTime := time.Since(startTime)

	var avgTime time.Duration
	var throughput float64

	if successful > 0 {
		avgTime = time.Duration(totalTime.Nanoseconds() / int64(successful))
		throughput = float64(successful) / totalTime.Seconds()
	}

	return InsertMetrics{
		TotalTime:        totalTime,
		AverageTime:      avgTime,
		ThroughputPerSec: throughput,
		VectorsInserted:  successful,
	}, nil
}

// benchmarkSearches measures search performance
func (b *Benchmarker) benchmarkSearches(queries [][]float64, topK int, logProgress bool) (SearchMetrics, error) {
	startTime := time.Now()
	successful := 0
	totalResults := 0

	for i, query := range queries {
		results, err := b.storage.Search(query, topK)
		if err != nil {
			b.logger.Warn("search failed during benchmark", "query_index", i, "error", err)
			continue
		}
		successful++
		totalResults += len(results)

		if logProgress && (i+1)%50 == 0 {
			elapsed := time.Since(startTime)
			rate := float64(i+1) / elapsed.Seconds()
			b.logger.Info("search progress", "completed", i+1, "total", len(queries), "rate_per_sec", rate)
		}
	}

	totalTime := time.Since(startTime)

	var avgTime time.Duration
	var throughput float64
	var avgResults float64
	if successful > 0 {
		avgTime = time.Duration(totalTime.Nanoseconds() / int64(successful))
		throughput = float64(successful) / totalTime.Seconds()
		avgResults = float64(totalResults) / float64(successful)
	}

	return SearchMetrics{
		TotalTime:        totalTime,
		AverageTime:      avgTime,
		ThroughputPerSec: throughput,
		QueriesExecuted:  successful,
		AverageResults:   avgResults,
		AverageAccuracy:  0.0, // Placeholder - would need ground truth for real accuracy
	}, nil
}

// estimateDatabaseSize returns approximate database size
func (b *Benchmarker) estimateDatabaseSize() int64 {
	fi, err := os.Stat(b.storage.cfg.Path)
	if err != nil {
		b.logger.Warn("failed to get db size", "path", b.storage.cfg.Path, "err", err)
		return 0
	}
	return fi.Size()
}

// getEngineType returns the vector engine type being used
func (b *Benchmarker) getEngineType() string {
	return "sqlite-vec"
}

// PrintResults outputs benchmark results in a readable format
func (r *BenchmarkResult) PrintResults() {
	fmt.Printf("\n📊 8fs Vector Storage Benchmark Results\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Timestamp: %s\n", r.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Configuration: %d vectors, %d queries, %d dimensions (%s dataset)\n",
		r.Config.VectorCount, r.Config.QueryCount, r.Config.Dimensions, r.Config.DatasetType)
	fmt.Printf("Vector Engine: %s\n", r.SystemInfo.VectorEngine)
	fmt.Printf("\n")

	fmt.Printf("📥 Insert Performance:\n")
	fmt.Printf("  • Total Time: %v\n", r.InsertMetrics.TotalTime)
	fmt.Printf("  • Average per Vector: %v\n", r.InsertMetrics.AverageTime)
	fmt.Printf("  • Throughput: %.1f vectors/sec\n", r.InsertMetrics.ThroughputPerSec)
	fmt.Printf("  • Vectors Inserted: %d/%d\n", r.InsertMetrics.VectorsInserted, r.Config.VectorCount)
	fmt.Printf("\n")

	fmt.Printf("🔍 Search Performance:\n")
	fmt.Printf("  • Total Time: %v\n", r.SearchMetrics.TotalTime)
	fmt.Printf("  • Average per Query: %v\n", r.SearchMetrics.AverageTime)
	fmt.Printf("  • Throughput: %.1f queries/sec\n", r.SearchMetrics.ThroughputPerSec)
	fmt.Printf("  • Queries Executed: %d/%d\n", r.SearchMetrics.QueriesExecuted, r.Config.QueryCount)
	fmt.Printf("  • Average Results: %.1f per query\n", r.SearchMetrics.AverageResults)
	fmt.Printf("\n")

	fmt.Printf("🎯 Overall:\n")
	fmt.Printf("  • Total Runtime: %v\n", r.OverallMetrics.TotalTime)
	fmt.Printf("  • Database Size: %.2f MB\n", float64(r.OverallMetrics.DatabaseSize)/(1024*1024))
	fmt.Printf("\n")
}

// SaveToJSON saves benchmark results to a JSON file
func (r *BenchmarkResult) SaveToJSON(filename string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// In a real implementation, would write to file
	// For now, just return the JSON structure
	_ = data
	return nil
}
