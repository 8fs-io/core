package main

import (
	"fmt"

	"github.com/8fs-io/core/internal/domain/vectors"
)

func main() {
	v := &vectors.Vector{
		ID:        "test",
		Embedding: []float64{1.0, 2.0, 3.0},
		Metadata:  map[string]interface{}{"test": true},
	}
	fmt.Printf("Vector created: %v\n", v.ID)
}
