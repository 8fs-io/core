package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/8fs-io/core/internal/domain/vectors"
	"github.com/8fs-io/core/pkg/logger"
)

func main() {
	// Command line flags
	var (
		dbPath      = flag.String("db", "./data/benchmark.db", "Database file path")
		output      = flag.String("output", "./benchmark_results.json", "Output file for results")
		vectorCount = flag.Int("vectors", 1000, "Number of vectors to insert")
		queryCount  = flag.Int("queries", 100, "Number of search queries")
		dimensions  = flag.Int("dims", 384, "Vector dimensions")
		topK        = flag.Int("topk", 10, "Top K results per search")
		datasetType = flag.String("dataset", "realistic", "Dataset type: random, clustered, realistic")
		seed        = flag.Int64("seed", 0, "Random seed (0 for current time)")
		verbose     = flag.Bool("verbose", true, "Enable progress logging")
		cleanup     = flag.Bool("cleanup", true, "Clean up database after benchmark")
		comparative = flag.Bool("compare", false, "Run comparative benchmarks across multiple configurations")
	)
	flag.Parse()

	fmt.Printf("🚀 8fs Vector Storage Performance Benchmark\n")
	fmt.Printf("============================================\n\n")

	// Initialize logger
	cfg := logger.Config{
		Level:  "DEBUG",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	if *comparative {
		runComparativeBenchmarks(*dbPath, *output, logger, *cleanup)
		return
	}

	// Single benchmark configuration
	config := vectors.BenchmarkConfig{
		VectorCount:       *vectorCount,
		QueryCount:        *queryCount,
		Dimensions:        *dimensions,
		TopK:              *topK,
		DatasetType:       *datasetType,
		RandomSeed:        *seed,
		EnableProgressLog: *verbose,
	}

	if *seed == 0 {
		config.RandomSeed = time.Now().UnixNano()
	}

	result, err := runSingleBenchmark(*dbPath, config, logger, *cleanup)
	if err != nil {
		log.Fatalf("Benchmark failed: %v", err)
	}

	// Print results
	result.PrintResults()

	// Save results
	if err := saveResults(result, *output); err != nil {
		log.Printf("Warning: Failed to save results to %s: %v", *output, err)
	} else {
		fmt.Printf("📄 Results saved to: %s\n", *output)
	}
}

func runSingleBenchmark(dbPath string, config vectors.BenchmarkConfig, logger vectors.Logger, cleanup bool) (*vectors.BenchmarkResult, error) {
	// Ensure database directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Clean up existing database for fresh benchmark
	if cleanup {
		os.Remove(dbPath)
		defer os.Remove(dbPath)
	}

	// Initialize storage
	cfg := vectors.SQLiteVecConfig{
		Path:      dbPath,
		Dimension: config.Dimensions,
	}
	storage, err := vectors.NewSQLiteVecStorage(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storage.Close()

	// Create benchmarker
	benchmarker := vectors.NewBenchmarker(storage, logger)

	// Run benchmark
	return benchmarker.RunBenchmark(config)
}

func runComparativeBenchmarks(dbPath, output string, logger vectors.Logger, cleanup bool) {
	fmt.Printf("🔬 Running Comparative Benchmarks\n\n")

	configs := []vectors.BenchmarkConfig{
		// Small dataset tests
		{VectorCount: 100, QueryCount: 50, Dimensions: 3, TopK: 5, DatasetType: "random", RandomSeed: 12345, EnableProgressLog: false},
		{VectorCount: 100, QueryCount: 50, Dimensions: 384, TopK: 5, DatasetType: "random", RandomSeed: 12345, EnableProgressLog: false},
		{VectorCount: 100, QueryCount: 50, Dimensions: 768, TopK: 5, DatasetType: "random", RandomSeed: 12345, EnableProgressLog: false},

		// Medium dataset tests
		{VectorCount: 1000, QueryCount: 100, Dimensions: 384, TopK: 10, DatasetType: "clustered", RandomSeed: 12345, EnableProgressLog: true},
		{VectorCount: 1000, QueryCount: 100, Dimensions: 384, TopK: 10, DatasetType: "realistic", RandomSeed: 12345, EnableProgressLog: true},

		// Large dataset test
		{VectorCount: 5000, QueryCount: 200, Dimensions: 384, TopK: 10, DatasetType: "realistic", RandomSeed: 12345, EnableProgressLog: true},
	}

	results := make([]*vectors.BenchmarkResult, 0, len(configs))

	for i, config := range configs {
		fmt.Printf("📊 Running benchmark %d/%d: %s dataset, %d dimensions, %d vectors, %d queries\n",
			i+1, len(configs), config.DatasetType, config.Dimensions, config.VectorCount, config.QueryCount)

		// Use different database for each test
		testDBPath := fmt.Sprintf("%s.test_%d.db", dbPath, i)

		result, err := runSingleBenchmark(testDBPath, config, logger, cleanup)
		if err != nil {
			log.Printf("Benchmark %d failed: %v", i+1, err)
			continue
		}

		results = append(results, result)

		// Quick summary
		fmt.Printf("  ✅ Insert: %.1f vec/sec, Search: %.1f query/sec\n\n",
			result.InsertMetrics.ThroughputPerSec, result.SearchMetrics.ThroughputPerSec)
	}

	// Print comparative results
	printComparativeResults(results)

	// Save all results
	if err := saveComparativeResults(results, output); err != nil {
		log.Printf("Warning: Failed to save comparative results: %v", err)
	} else {
		fmt.Printf("📄 Comparative results saved to: %s\n", output)
	}
}

func printComparativeResults(results []*vectors.BenchmarkResult) {
	if len(results) == 0 {
		return
	}

	fmt.Printf("📈 Comparative Performance Summary\n")
	fmt.Printf("==================================\n\n")

	fmt.Printf("%-15s %-8s %-8s %-12s %-12s %-10s %-10s\n",
		"Dataset", "Dims", "Vectors", "Insert/sec", "Search/sec", "Fallback", "Total Time")
	fmt.Printf("%-15s %-8s %-8s %-12s %-12s %-10s %-10s\n",
		"-------", "----", "-------", "----------", "----------", "--------", "----------")

	for _, result := range results {
		fallback := "No" // Always "No" since we removed fallback mode

		fmt.Printf("%-15s %-8d %-8d %-12.1f %-12.1f %-10s %-10v\n",
			result.Config.DatasetType,
			result.Config.Dimensions,
			result.Config.VectorCount,
			result.InsertMetrics.ThroughputPerSec,
			result.SearchMetrics.ThroughputPerSec,
			fallback,
			result.OverallMetrics.TotalTime.Truncate(time.Millisecond))
	}

	// Find best performers
	var bestInsert, bestSearch *vectors.BenchmarkResult
	for _, result := range results {
		if bestInsert == nil || result.InsertMetrics.ThroughputPerSec > bestInsert.InsertMetrics.ThroughputPerSec {
			bestInsert = result
		}
		if bestSearch == nil || result.SearchMetrics.ThroughputPerSec > bestSearch.SearchMetrics.ThroughputPerSec {
			bestSearch = result
		}
	}

	fmt.Printf("\n🏆 Performance Leaders:\n")
	if bestInsert != nil {
		fmt.Printf("  • Best Insert: %.1f vec/sec (%s, %dd)\n",
			bestInsert.InsertMetrics.ThroughputPerSec, bestInsert.Config.DatasetType, bestInsert.Config.Dimensions)
	}
	if bestSearch != nil {
		fmt.Printf("  • Best Search: %.1f query/sec (%s, %dd)\n",
			bestSearch.SearchMetrics.ThroughputPerSec, bestSearch.Config.DatasetType, bestSearch.Config.Dimensions)
	}

	// Engine analysis - since we removed fallback, all are sqlite-vec
	sqliteVecCount := len(results)
	fallbackCount := 0

	fmt.Printf("\n🔧 Engine Distribution:\n")
	fmt.Printf("  • SQLite-vec: %d benchmarks\n", sqliteVecCount)
	fmt.Printf("  • Fallback: %d benchmarks\n", fallbackCount)
	fmt.Printf("\n")
}

func saveResults(result *vectors.BenchmarkResult, filename string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

func saveComparativeResults(results []*vectors.BenchmarkResult, filename string) error {
	// Create a comparative results structure
	comparative := struct {
		Timestamp time.Time                  `json:"timestamp"`
		Results   []*vectors.BenchmarkResult `json:"results"`
		Summary   map[string]interface{}     `json:"summary"`
	}{
		Timestamp: time.Now(),
		Results:   results,
		Summary: map[string]interface{}{
			"total_benchmarks": len(results),
			"engine_distribution": map[string]int{
				"sqlite-vec": countEngine(results, false),
				"fallback":   countEngine(results, true),
			},
		},
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	data, err := json.MarshalIndent(comparative, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal comparative results: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

func countEngine(results []*vectors.BenchmarkResult, fallback bool) int {
	// Since we removed fallback mode, return count based on parameter
	if fallback {
		return 0 // No fallback mode anymore
	}
	return len(results) // All results are sqlite-vec
}
