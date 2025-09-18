package vectors

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// DataGenerator handles generation of test data for vectors
type DataGenerator struct {
	rand *rand.Rand
}

// NewDataGenerator creates a new data generator with the given seed
func NewDataGenerator(seed int64) *DataGenerator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &DataGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// GenerateVectorsConfig defines parameters for vector generation
type GenerateVectorsConfig struct {
	Count      int    // Number of vectors to generate
	Dimensions int    // Vector dimensions
	Type       string // Dataset type: "random", "clustered", "realistic"
}

// GenerateQueriesConfig defines parameters for query generation
type GenerateQueriesConfig struct {
	Count      int    // Number of queries to generate
	Dimensions int    // Vector dimensions
	Type       string // Dataset type: "random", "clustered", "realistic"
}

// GenerateVectors creates vectors based on the given configuration
func (g *DataGenerator) GenerateVectors(config GenerateVectorsConfig) []*Vector {
	switch config.Type {
	case "clustered":
		return g.generateClusteredVectors(config.Count, config.Dimensions)
	case "random":
		return g.generateRandomVectors(config.Count, config.Dimensions)
	default: // "realistic"
		return g.generateRealisticVectors(config.Count, config.Dimensions)
	}
}

// GenerateQueries creates query vectors based on the given configuration
func (g *DataGenerator) GenerateQueries(config GenerateQueriesConfig) [][]float64 {
	switch config.Type {
	case "clustered":
		return g.generateClusteredQueries(config.Count, config.Dimensions)
	case "random":
		return g.generateRandomQueries(config.Count, config.Dimensions)
	default: // "realistic"
		return g.generateRealisticQueries(config.Count, config.Dimensions)
	}
}

// GenerateVectorsAndQueries generates both vectors and queries for benchmarking
func (g *DataGenerator) GenerateVectorsAndQueries(vectorConfig GenerateVectorsConfig, queryConfig GenerateQueriesConfig) ([]*Vector, [][]float64) {
	vectors := g.GenerateVectors(vectorConfig)
	queries := g.GenerateQueries(queryConfig)
	return vectors, queries
}

// generateRandomVectors creates completely random vectors
func (g *DataGenerator) generateRandomVectors(count, dimensions int) []*Vector {
	result := make([]*Vector, count)

	for i := 0; i < count; i++ {
		embedding := make([]float64, dimensions)
		for j := 0; j < dimensions; j++ {
			embedding[j] = g.rand.NormFloat64()
		}

		result[i] = &Vector{
			ID:        fmt.Sprintf("random_%d", i),
			Embedding: g.normalizeVector(embedding),
			Metadata: map[string]interface{}{
				"type":       "random",
				"index":      i,
				"cluster_id": i % 10, // Fake clustering for metadata filtering
				"created_at": time.Now().Add(-time.Duration(i) * time.Minute).Unix(),
			},
		}
	}

	return result
}

// generateRandomQueries creates random query vectors
func (g *DataGenerator) generateRandomQueries(count, dimensions int) [][]float64 {
	queries := make([][]float64, count)

	for i := 0; i < count; i++ {
		embedding := make([]float64, dimensions)
		for j := 0; j < dimensions; j++ {
			embedding[j] = g.rand.NormFloat64()
		}
		queries[i] = g.normalizeVector(embedding)
	}

	return queries
}

// generateClusteredVectors creates vectors in distinct clusters
func (g *DataGenerator) generateClusteredVectors(count, dimensions int) []*Vector {
	result := make([]*Vector, count)
	numClusters := 5

	// Generate cluster centers
	clusterCenters := make([][]float64, numClusters)
	for i := 0; i < numClusters; i++ {
		center := make([]float64, dimensions)
		for j := 0; j < dimensions; j++ {
			center[j] = g.rand.NormFloat64() * 2 // Wider spread for centers
		}
		clusterCenters[i] = g.normalizeVector(center)
	}

	categories := []string{"technology", "science", "business", "health", "entertainment"}

	for i := 0; i < count; i++ {
		clusterID := i % numClusters
		center := clusterCenters[clusterID]

		embedding := make([]float64, dimensions)
		for j := 0; j < dimensions; j++ {
			// Add noise around cluster center
			embedding[j] = center[j] + g.rand.NormFloat64()*0.1
		}

		result[i] = &Vector{
			ID:        fmt.Sprintf("cluster_%d_%d", clusterID, i),
			Embedding: g.normalizeVector(embedding),
			Metadata: map[string]interface{}{
				"type":       "clustered",
				"cluster_id": clusterID,
				"category":   categories[clusterID],
				"topic":      fmt.Sprintf("topic_%d", clusterID),
				"index":      i,
				"created_at": time.Now().Add(-time.Duration(i) * time.Minute).Unix(),
			},
		}
	}

	return result
}

// generateClusteredQueries creates queries that should find clusters
func (g *DataGenerator) generateClusteredQueries(count, dimensions int) [][]float64 {
	queries := make([][]float64, count)
	numClusters := 5

	// Generate cluster centers (same as in vectors)
	clusterCenters := make([][]float64, numClusters)
	for i := 0; i < numClusters; i++ {
		center := make([]float64, dimensions)
		for j := 0; j < dimensions; j++ {
			center[j] = g.rand.NormFloat64() * 2
		}
		clusterCenters[i] = g.normalizeVector(center)
	}

	for i := 0; i < count; i++ {
		clusterID := i % numClusters
		center := clusterCenters[clusterID]

		embedding := make([]float64, dimensions)
		for j := 0; j < dimensions; j++ {
			// Query slightly off-center to test similarity
			embedding[j] = center[j] + g.rand.NormFloat64()*0.05
		}
		queries[i] = g.normalizeVector(embedding)
	}

	return queries
}

// generateRealisticVectors simulates realistic text embeddings
func (g *DataGenerator) generateRealisticVectors(count, dimensions int) []*Vector {
	result := make([]*Vector, count)

	// Realistic embedding patterns (sparse with some dominant dimensions)
	dominantDims := dimensions / 10 // ~10% of dimensions are "important"

	documentTypes := []string{"article", "blog", "paper", "report", "review"}
	topics := []string{"ai", "blockchain", "cloud", "data", "security", "mobile", "web", "iot"}

	for i := 0; i < count; i++ {
		embedding := make([]float64, dimensions)

		// Most dimensions are near-zero
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

		result[i] = &Vector{
			ID:        fmt.Sprintf("doc_%s_%s_%d", docType, topic, i),
			Embedding: g.normalizeVector(embedding),
			Metadata: map[string]interface{}{
				"type":       "document",
				"doc_type":   docType,
				"topic":      topic,
				"category":   []string{"tech", "science", "business", "health"}[i%4],
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

// generateRealisticQueries creates realistic query vectors
func (g *DataGenerator) generateRealisticQueries(count, dimensions int) [][]float64 {
	queries := make([][]float64, count)
	dominantDims := dimensions / 10

	for i := 0; i < count; i++ {
		embedding := make([]float64, dimensions)

		// Similar pattern to documents but with variation
		for j := 0; j < dimensions; j++ {
			embedding[j] = g.rand.NormFloat64() * 0.15
		}

		for k := 0; k < dominantDims; k++ {
			dimIndex := g.rand.Intn(dimensions)
			embedding[dimIndex] += g.rand.NormFloat64() * 0.6
		}

		queries[i] = g.normalizeVector(embedding)
	}

	return queries
}

// normalizeVector converts vector to unit length
func (g *DataGenerator) normalizeVector(v []float64) []float64 {
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