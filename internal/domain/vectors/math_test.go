package vectors

import (
	"fmt"
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	vm := NewVectorMath()
	
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
		wantErr  bool
	}{
		{
			name:     "identical vectors",
			a:        []float64{1.0, 2.0, 3.0},
			b:        []float64{1.0, 2.0, 3.0},
			expected: 1.0,
			wantErr:  false,
		},
		{
			name:     "opposite vectors",
			a:        []float64{1.0, 2.0, 3.0},
			b:        []float64{-1.0, -2.0, -3.0},
			expected: -1.0,
			wantErr:  false,
		},
		{
			name:     "orthogonal vectors",
			a:        []float64{1.0, 0.0},
			b:        []float64{0.0, 1.0},
			expected: 0.0,
			wantErr:  false,
		},
		{
			name:     "zero vector",
			a:        []float64{0.0, 0.0, 0.0},
			b:        []float64{1.0, 2.0, 3.0},
			expected: 0.0,
			wantErr:  false,
		},
		{
			name:     "different dimensions",
			a:        []float64{1.0, 2.0},
			b:        []float64{1.0, 2.0, 3.0},
			expected: 0.0,
			wantErr:  true,
		},
		{
			name:     "empty vectors",
			a:        []float64{},
			b:        []float64{},
			expected: 0.0,
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.CosineSimilarity(tt.a, tt.b)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestDotProduct(t *testing.T) {
	vm := NewVectorMath()
	
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{
			name:     "simple dot product",
			a:        []float64{1.0, 2.0, 3.0},
			b:        []float64{4.0, 5.0, 6.0},
			expected: 32.0, // 1*4 + 2*5 + 3*6 = 4 + 10 + 18 = 32
		},
		{
			name:     "zero dot product",
			a:        []float64{1.0, 0.0},
			b:        []float64{0.0, 1.0},
			expected: 0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.DotProduct(tt.a, tt.b)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestL2Norm(t *testing.T) {
	vm := NewVectorMath()
	
	tests := []struct {
		name     string
		v        []float64
		expected float64
	}{
		{
			name:     "unit vector",
			v:        []float64{1.0, 0.0, 0.0},
			expected: 1.0,
		},
		{
			name:     "3-4-5 triangle",
			v:        []float64{3.0, 4.0},
			expected: 5.0,
		},
		{
			name:     "zero vector",
			v:        []float64{0.0, 0.0, 0.0},
			expected: 0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.L2Norm(tt.v)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestValidateDimensions(t *testing.T) {
	vm := NewVectorMath()
	
	tests := []struct {
		name      string
		embedding []float64
		wantErr   bool
	}{
		{
			name:      "valid 384 dimensions",
			embedding: make([]float64, 384),
			wantErr:   false,
		},
		{
			name:      "valid 768 dimensions",
			embedding: make([]float64, 768),
			wantErr:   false,
		},
		{
			name:      "valid 1536 dimensions",
			embedding: make([]float64, 1536),
			wantErr:   false,
		},
		{
			name:      "too small - 383 dimensions",
			embedding: make([]float64, 383),
			wantErr:   true,
		},
		{
			name:      "too large - 1537 dimensions",
			embedding: make([]float64, 1537),
			wantErr:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateDimensions(tt.embedding)
			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateVector(t *testing.T) {
	vm := NewVectorMath()
	
	tests := []struct {
		name    string
		vector  *Vector
		wantErr bool
	}{
		{
			name: "valid vector",
			vector: &Vector{
				ID:        "test1",
				Embedding: make([]float64, 384),
				Metadata:  map[string]interface{}{"type": "test"},
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			vector: &Vector{
				ID:        "",
				Embedding: make([]float64, 384),
			},
			wantErr: true,
		},
		{
			name: "empty embedding",
			vector: &Vector{
				ID:        "test1",
				Embedding: []float64{},
			},
			wantErr: true,
		},
		{
			name: "invalid dimensions",
			vector: &Vector{
				ID:        "test1",
				Embedding: make([]float64, 100),
			},
			wantErr: true,
		},
		{
			name: "NaN value",
			vector: &Vector{
				ID:        "test1",
				Embedding: []float64{1.0, math.NaN(), 3.0},
			},
			wantErr: true,
		},
		{
			name: "infinite value",
			vector: &Vector{
				ID:        "test1",
				Embedding: []float64{1.0, math.Inf(1), 3.0},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateVector(tt.vector)
			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkCosineSimilarity(b *testing.B) {
	vm := NewVectorMath()
	
	// Test with common embedding dimensions
	dimensions := []int{384, 768, 1536}
	
	for _, dim := range dimensions {
		vec1 := make([]float64, dim)
		vec2 := make([]float64, dim)
		
		// Fill with random-ish values
		for i := 0; i < dim; i++ {
			vec1[i] = float64(i) * 0.001
			vec2[i] = float64(i) * 0.002
		}
		
		b.Run(fmt.Sprintf("dim-%d", dim), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				vm.CosineSimilarity(vec1, vec2)
			}
		})
	}
}

func BenchmarkL2Norm(b *testing.B) {
	vm := NewVectorMath()
	
	vec := make([]float64, 768)
	for i := 0; i < 768; i++ {
		vec[i] = float64(i) * 0.001
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.L2Norm(vec)
	}
}
