package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/8fs-io/core/internal/domain/vectors"
	"github.com/8fs-io/core/pkg/logger"
)

func main() {
	var (
		dbPath      = flag.String("db", "./data/sample.db", "Database file path")
		count       = flag.Int("count", 100, "Number of vectors to generate")
		dimensions  = flag.Int("dims", 384, "Vector dimensions")
		datasetType = flag.String("type", "realistic", "Dataset type: random, clustered, realistic")
		seed        = flag.Int64("seed", 0, "Random seed")
		verbose     = flag.Bool("verbose", true, "Enable verbose output")
	)
	flag.Parse()

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rand.Seed(*seed)

	fmt.Printf("🎯 Generating %d sample vectors (%dD, %s dataset)\n", *count, *dimensions, *datasetType)

	// Initialize logger
	cfg := logger.Config{
		Level:  "INFO",
		Format: "text",
		Output: "stdout",
	}
	logger, err := logger.New(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage
	storageCfg := vectors.SQLiteVecConfig{
		Path:      *dbPath,
		Dimension: *dimensions,
	}
	storage, err := vectors.NewSQLiteVecStorage(storageCfg, logger)
	if err != nil {
		fmt.Printf("Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	// Generate and store vectors
	generator := NewDataGenerator(*seed)
	vectorData := generator.GenerateVectors(*count, *dimensions, *datasetType)

	successful := 0
	for i, vector := range vectorData {
		if err := storage.Store(vector); err != nil {
			if *verbose {
				fmt.Printf("Failed to store vector %s: %v\n", vector.ID, err)
			}
			continue
		}
		successful++

		if *verbose && (i+1)%50 == 0 {
			fmt.Printf("Stored %d/%d vectors\n", i+1, len(vectorData))
		}
	}

	fmt.Printf("✅ Successfully generated %d/%d vectors in %s\n", successful, *count, *dbPath)
}

type DataGenerator struct {
	rand *rand.Rand
}

func NewDataGenerator(seed int64) *DataGenerator {
	return &DataGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (g *DataGenerator) GenerateVectors(count, dimensions int, datasetType string) []*vectors.Vector {
	switch datasetType {
	case "clustered":
		return g.generateClusteredVectors(count, dimensions)
	case "random":
		return g.generateRandomVectors(count, dimensions)
	default: // "realistic"
		return g.generateRealisticVectors(count, dimensions)
	}
}

func (g *DataGenerator) generateRandomVectors(count, dimensions int) []*vectors.Vector {
	result := make([]*vectors.Vector, count)

	for i := 0; i < count; i++ {
		embedding := make([]float64, dimensions)
		for j := 0; j < dimensions; j++ {
			embedding[j] = g.rand.NormFloat64()
		}

		result[i] = &vectors.Vector{
			ID:        fmt.Sprintf("random_%d", i),
			Embedding: g.normalizeVector(embedding),
			Metadata: map[string]interface{}{
				"type":       "random",
				"index":      i,
				"created_at": time.Now().Add(-time.Duration(i) * time.Minute).Unix(),
			},
		}
	}

	return result
}

func (g *DataGenerator) generateClusteredVectors(count, dimensions int) []*vectors.Vector {
	result := make([]*vectors.Vector, count)
	numClusters := 5

	// Generate cluster centers
	clusterCenters := make([][]float64, numClusters)
	for i := 0; i < numClusters; i++ {
		center := make([]float64, dimensions)
		for j := 0; j < dimensions; j++ {
			center[j] = g.rand.NormFloat64() * 2
		}
		clusterCenters[i] = g.normalizeVector(center)
	}

	categories := []string{"technology", "science", "business", "health", "entertainment"}

	for i := 0; i < count; i++ {
		clusterID := i % numClusters
		center := clusterCenters[clusterID]

		embedding := make([]float64, dimensions)
		for j := 0; j < dimensions; j++ {
			embedding[j] = center[j] + g.rand.NormFloat64()*0.1
		}

		result[i] = &vectors.Vector{
			ID:        fmt.Sprintf("cluster_%d_%d", clusterID, i),
			Embedding: g.normalizeVector(embedding),
			Metadata: map[string]interface{}{
				"type":       "clustered",
				"cluster_id": clusterID,
				"category":   categories[clusterID],
				"index":      i,
				"created_at": time.Now().Add(-time.Duration(i) * time.Minute).Unix(),
			},
		}
	}

	return result
}

func (g *DataGenerator) generateRealisticVectors(count, dimensions int) []*vectors.Vector {
	result := make([]*vectors.Vector, count)

	// Simulate text embedding patterns
	dominantDims := dimensions / 10

	documentTypes := []string{"article", "blog", "paper", "report", "review"}
	topics := []string{"ai", "blockchain", "cloud", "data", "security", "mobile", "web", "iot"}

	for i := 0; i < count; i++ {
		embedding := make([]float64, dimensions)

		// Most dimensions are small
		for j := 0; j < dimensions; j++ {
			embedding[j] = g.rand.NormFloat64() * 0.1
		}

		// Some dimensions have stronger signals
		for k := 0; k < dominantDims; k++ {
			dimIndex := g.rand.Intn(dimensions)
			embedding[dimIndex] += g.rand.NormFloat64() * 0.8
		}

		docType := documentTypes[i%len(documentTypes)]
		topic := topics[i%len(topics)]

		result[i] = &vectors.Vector{
			ID:        fmt.Sprintf("doc_%s_%s_%d", docType, topic, i),
			Embedding: g.normalizeVector(embedding),
			Metadata: map[string]interface{}{
				"type":       "document",
				"doc_type":   docType,
				"topic":      topic,
				"word_count": 100 + g.rand.Intn(2000),
				"language":   "en",
				"author":     fmt.Sprintf("author_%d", g.rand.Intn(20)),
				"published":  time.Now().Add(-time.Duration(g.rand.Intn(365*24)) * time.Hour).Unix(),
				"relevance":  g.rand.Float64(),
			},
		}
	}

	return result
}

func (g *DataGenerator) normalizeVector(v []float64) []float64 {
	var magnitude float64
	for _, val := range v {
		magnitude += val * val
	}
	magnitude = math.Sqrt(magnitude)

	if magnitude == 0 {
		return v
	}

	normalized := make([]float64, len(v))
	for i, val := range v {
		normalized[i] = val / magnitude
	}
	return normalized
}
