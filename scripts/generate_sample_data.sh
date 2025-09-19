#!/bin/bash
# Simple data generator using the benchmark tool
set -e

# Configuration
DB_PATH="${1:-./data/sample_data.db}"
VECTOR_COUNT="${2:-1000}"
DIMENSIONS="${3:-384}"
DATASET_TYPE="${4:-realistic}"

echo "🎯 Generating sample data using benchmark tool..."
echo "   Database: $DB_PATH"
echo "   Vectors: $VECTOR_COUNT"
echo "   Dimensions: $DIMENSIONS" 
echo "   Dataset: $DATASET_TYPE"
echo ""

# Check if we're in the right directory (has go.mod)
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Please run this script from the project root directory."
    echo "   Current directory: $(pwd)"
    exit 1
fi

# Ensure data directory exists
mkdir -p "$(dirname "$DB_PATH")"

# Clean existing database
rm -f "$DB_PATH"

# Build benchmark tool if needed
if [ ! -f "./benchmark" ]; then
    echo "Building benchmark tool..."
    if ! go build -o benchmark ./cmd/benchmark/; then
        echo "❌ Error: Failed to build benchmark tool"
        exit 1
    fi
fi

# Generate data by running a benchmark and keeping the database
echo "Generating vectors..."
./benchmark \
    -db "$DB_PATH" \
    -vectors "$VECTOR_COUNT" \
    -queries 1 \
    -dims "$DIMENSIONS" \
    -dataset "$DATASET_TYPE" \
    -topk 1 \
    -verbose=false \
    -cleanup=false \
    -output /tmp/generate_results.json

# Verify the data was created
if [ -f "$DB_PATH" ]; then
    SIZE=$(du -h "$DB_PATH" | cut -f1)
    echo "✅ Sample data generated successfully!"
    echo "   Database size: $SIZE"
    echo "   Location: $DB_PATH"
    
    # Quick database info
    echo ""
    echo "📊 Database Info:"
    sqlite3 "$DB_PATH" "SELECT COUNT(*) as vector_count FROM vectors;" 2>/dev/null || \
    sqlite3 "$DB_PATH" "SELECT COUNT(*) as vector_count FROM fallback_vectors;" 2>/dev/null || \
    echo "   Unable to count vectors (database may use different schema)"
    
else
    echo "❌ Failed to generate sample data"
    exit 1
fi

echo ""
echo "💡 Use this database for testing:"
echo "   ./cmd/benchmark/benchmark -db $DB_PATH -queries 50 -dims $DIMENSIONS"
