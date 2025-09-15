package vectors

import (
	"fmt"
	"math"
)

// Vector represents an embedding vector with metadata
type Vector struct {
	ID        string                 `json:"id"`
	Embedding []float64              `json:"embedding"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// VectorMath provides pure Go implementations of vector operations
type VectorMath struct{}

// NewVectorMath creates a new VectorMath instance
func NewVectorMath() *VectorMath {
	return &VectorMath{}
}

// CosineSimilarity calculates cosine similarity between two vectors
// Returns a value between -1 and 1, where 1 means identical direction
func (vm *VectorMath) CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0.0, fmt.Errorf("vector dimensions don't match: %d vs %d", len(a), len(b))
	}
	
	if len(a) == 0 {
		return 0.0, fmt.Errorf("cannot compute similarity of empty vectors")
	}
	
	dotProduct := vm.DotProduct(a, b)
	normA := vm.L2Norm(a)
	normB := vm.L2Norm(b)
	
	// Handle zero vectors
	if normA == 0 || normB == 0 {
		return 0.0, nil
	}
	
	similarity := dotProduct / (normA * normB)
	
	// Clamp to [-1, 1] to handle floating point precision issues
	if similarity > 1.0 {
		similarity = 1.0
	} else if similarity < -1.0 {
		similarity = -1.0
	}
	
	return similarity, nil
}

// DotProduct calculates dot product of two vectors
func (vm *VectorMath) DotProduct(a, b []float64) float64 {
	var sum float64
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}
	return sum
}

// L2Norm calculates L2 (Euclidean) norm of a vector
func (vm *VectorMath) L2Norm(v []float64) float64 {
	var sum float64
	for _, val := range v {
		sum += val * val
	}
	return math.Sqrt(sum)
}

// ValidateDimensions checks if vector dimensions are within acceptable range
func (vm *VectorMath) ValidateDimensions(embedding []float64) error {
	dim := len(embedding)
	if dim < 384 || dim > 1536 {
		return fmt.Errorf("embedding dimension %d out of range [384, 1536]", dim)
	}
	return nil
}

// ValidateVector performs comprehensive vector validation
func (vm *VectorMath) ValidateVector(v *Vector) error {
	if v.ID == "" {
		return fmt.Errorf("vector ID cannot be empty")
	}
	
	if len(v.Embedding) == 0 {
		return fmt.Errorf("embedding cannot be empty")
	}
	
	if err := vm.ValidateDimensions(v.Embedding); err != nil {
		return err
	}
	
	// Check for NaN or infinite values
	for i, val := range v.Embedding {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return fmt.Errorf("invalid value at dimension %d: %f", i, val)
		}
	}
	
	return nil
}

// SearchResult represents a vector search match
type SearchResult struct {
	Vector *Vector `json:"vector"`
	Score  float64 `json:"score"`
}

// SearchResults is a slice of search results that can be sorted
type SearchResults []SearchResult

func (sr SearchResults) Len() int           { return len(sr) }
func (sr SearchResults) Swap(i, j int)      { sr[i], sr[j] = sr[j], sr[i] }
func (sr SearchResults) Less(i, j int) bool { return sr[i].Score > sr[j].Score } // Higher score first
