package vectors

import (
	"math/rand"
	"path/filepath"
	"testing"
)

func TestSQLiteVecStorage(t *testing.T) {
	// Create temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_vectors.db")

	// Initialize storage
	storage, err := NewSQLiteVecStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteVecStorage: %v", err)
	}
	defer storage.Close()

	// Test data
	testVectors := []*Vector{
		{
			ID:        "test1",
			Embedding: []float64{0.1, 0.2, 0.3, 0.4},
			Metadata:  map[string]interface{}{"description": "Test vector 1"},
		},
		{
			ID:        "test2", 
			Embedding: []float64{0.5, 0.6, 0.7, 0.8},
			Metadata:  map[string]interface{}{"description": "Test vector 2"},
		},
	}

	// Test storing vectors
	for _, vector := range testVectors {
		err := storage.Store(vector)
		if err != nil {
			t.Errorf("Failed to store vector %s: %v", vector.ID, err)
		}
	}

	// Test search
	query := []float64{0.4, 0.5, 0.6, 0.7}
	results, err := storage.Search(query, 2)
	if err != nil {
		t.Fatalf("Failed to search vectors: %v", err)
	}

	// Verify results
	if len(results) == 0 {
		t.Error("Expected search results, got none")
	}

	t.Logf("Search completed successfully with %d results", len(results))
	for i, result := range results {
		t.Logf("Result %d: ID=%s, Score=%.4f", i+1, result.Vector.ID, result.Score)
	}
}

// Helper function for generating test data
func generateRandomVector(dimensions int) []float64 {
	vec := make([]float64, dimensions)
	for i := 0; i < dimensions; i++ {
		vec[i] = rand.Float64()*2.0 - 1.0 // Range [-1, 1]
	}
	return vec
}
