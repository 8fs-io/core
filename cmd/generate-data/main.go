package main

import (
	"flag"
	"fmt"
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
	generator := vectors.NewDataGenerator(*seed)
	config := vectors.GenerateVectorsConfig{
		Count:      *count,
		Dimensions: *dimensions,
		Type:       *datasetType,
	}
	vectorData := generator.GenerateVectors(config)

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
