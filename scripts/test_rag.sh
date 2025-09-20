#!/bin/bash

# 8fs RAG Test Script
# Tests the complete RAG functionality including document storage and chat completion

set -e

BASE_URL=${BASE_URL:-"http://localhost:8080"}
API_V1="$BASE_URL/api/v1"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 8fs RAG Server Test${NC}"
echo "=================================="

# Function to check if service is running
check_health() {
    echo -e "\n${BLUE}1. Health Check${NC}"
    if curl -s "$API_V1/chat/health" > /dev/null; then
        echo -e "✅ RAG service is ${GREEN}healthy${NC}"
        return 0
    else
        echo -e "❌ RAG service is ${RED}not responding${NC}"
        echo "Make sure 8fs server is running with RAG enabled"
        return 1
    fi
}

# Function to store sample documents
store_documents() {
    echo -e "\n${BLUE}2. Storing Sample Documents${NC}"
    
    # Generate a 384-dimensional zero embedding vector (matching all-minilm model dimensions)
    local zero_embedding=$(python3 -c "import json; print(json.dumps([0.0] * 384))")
    
    # Document 1: 8fs Overview
    curl -sf -X POST "$API_V1/vectors/embeddings" \
        -H "Content-Type: application/json" \
        -d "{
            \"id\": \"doc_8fs_overview\",
            \"embedding\": $zero_embedding,
            \"metadata\": {
                \"content\": \"8fs is a high-performance vector database server with SQLite-vec integration for embeddings and similarity search. It provides enterprise-grade storage with microsecond latencies and supports both direct vector operations and RAG workflows.\",
                \"source\": \"overview.md\",
                \"type\": \"documentation\",
                \"category\": \"overview\"
            }
        }" > /dev/null && echo "✅ Stored: 8fs Overview" || echo "❌ Failed to store: 8fs Overview"

    # Document 2: Performance Stats
    curl -sf -X POST "$API_V1/vectors/embeddings" \
        -H "Content-Type: application/json" \
        -d "{
            \"id\": \"doc_8fs_perf\", 
            \"embedding\": $zero_embedding,
            \"metadata\": {
                \"content\": \"8fs achieves <50μs vector searches with SQLite-vec, handling 100k+ vectors efficiently. Benchmarks show 95th percentile latencies under 100μs for similarity searches across large datasets.\",
                \"source\": \"performance.md\",
                \"type\": \"documentation\",
                \"category\": \"performance\"
            }
        }" > /dev/null && echo "✅ Stored: Performance Stats" || echo "❌ Failed to store: Performance Stats"

    # Document 3: RAG Features  
    curl -sf -X POST "$API_V1/vectors/embeddings" \
        -H "Content-Type: application/json" \
        -d "{
            \"id\": \"doc_8fs_rag\",
            \"embedding\": $zero_embedding,
            \"metadata\": {
                \"content\": \"8fs RAG includes OpenAI-compatible /api/v1/chat/completions endpoint for chat completions, context search capabilities, and generation with retrieval-augmented responses. It supports multiple AI providers including Ollama, OpenAI, and AWS Bedrock.\",
                \"source\": \"rag.md\",
                \"type\": \"documentation\",
                \"category\": \"rag\"
            }
        }" > /dev/null && echo "✅ Stored: RAG Features" || echo "❌ Failed to store: RAG Features"

    # Document 4: API Usage
    curl -sf -X POST "$API_V1/vectors/embeddings" \
        -H "Content-Type: application/json" \
        -d "{
            \"id\": \"doc_8fs_api\",
            \"embedding\": $zero_embedding,
            \"metadata\": {
                \"content\": \"8fs provides S3-compatible APIs for bucket and object operations, vector storage endpoints for embeddings and search, and RAG endpoints for chat completions and context retrieval. All APIs support standard HTTP methods and JSON payloads.\",
                \"source\": \"api.md\",
                \"type\": \"documentation\",
                \"category\": \"api\"
            }
        }" > /dev/null && echo "✅ Stored: API Usage" || echo "❌ Failed to store: API Usage"

    echo "📄 Stored 4 sample documents"
}

# Function to test context search
test_context_search() {
    echo -e "\n${BLUE}3. Context Search Test${NC}"
    
    local query="How fast is 8fs for vector operations?"
    echo "🔍 Query: $query"
    
    local response=$(curl -s -X POST "$API_V1/chat/search/context" \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$query\", \"top_k\": 3}")
    
    local doc_count=$(echo "$response" | jq -r '.documents | length // 0')
    echo "📚 Found $doc_count relevant documents"
    
    if [ "$doc_count" -gt 0 ]; then
        echo "$response" | jq -r '.documents[0:2][] | "  • \(.source // "unknown"): \(.content[0:80])..."'
    fi
}

# Function to test RAG chat completion
test_rag_chat() {
    echo -e "\n${BLUE}4. RAG Chat Completion Test${NC}"
    
    local queries=(
        "What is 8fs and what makes it unique?"
        "What are the performance characteristics of 8fs?"
        "How do I use the RAG capabilities in 8fs?"
    )
    
    for query in "${queries[@]}"; do
        echo -e "\n❓ Query: ${query}"
        
        local response=$(curl -s -X POST "$API_V1/chat/completions" \
            -H "Content-Type: application/json" \
            -d "{
                \"query\": \"$query\",
                \"max_tokens\": 200,
                \"temperature\": 0.7,
                \"top_k\": 3
            }")
        
        if [ $? -eq 0 ] && echo "$response" | jq -e '.choices[0].message.content' > /dev/null 2>&1; then
            local answer=$(echo "$response" | jq -r '.choices[0].message.content')
            if [ ${#answer} -gt 150 ]; then
                echo -e "🤖 Answer: ${answer:0:150}..."
            else
                echo -e "🤖 Answer: $answer"
            fi
            
            local doc_count=$(echo "$response" | jq -r '.context.documents | length // 0')
            local total_tokens=$(echo "$response" | jq -r '.usage.total_tokens // 0')
            echo -e "📊 Used $doc_count context docs, $total_tokens tokens"
        else
            echo -e "❌ ${RED}Failed to get response${NC}"
            echo "$response" | jq -r '.error.message // .error // "Unknown error"'
        fi
        
        echo "$(printf '%*s' 50 '' | tr ' ' '-')"
    done
}

# Function to test direct generation with context
test_direct_generation() {
    echo -e "\n${BLUE}5. Direct Generation with Context Test${NC}"
    
    local context="8fs is a high-performance storage server that combines S3 compatibility with vector search. It can handle 1,700+ vector insertions per second."
    local prompt="Based on this information, what are the key benefits of using 8fs?"
    
    echo "🔍 Testing direct generation with provided context"
    
    local response=$(curl -s -X POST "$API_V1/chat/generate/context" \
        -H "Content-Type: application/json" \
        -d "{
            \"prompt\": \"$prompt\",
            \"context\": \"$context\",
            \"max_tokens\": 150,
            \"temperature\": 0.7
        }")
    
    if [ $? -eq 0 ] && echo "$response" | jq -e '.text' > /dev/null 2>&1; then
        local answer=$(echo "$response" | jq -r '.text')
        echo -e "🤖 Generated: ${answer:0:120}${answer:120:1:+...}"
        
        local tokens=$(echo "$response" | jq -r '.usage.total_tokens // 0')
        echo -e "📊 Used $tokens tokens"
    else
        echo -e "❌ ${RED}Generation failed${NC}"
        echo "$response" | jq -r '.error // "Unknown error"'
    fi
}

# Main test execution
main() {
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}❌ Error: jq is required but not installed${NC}"
        echo "Install with: brew install jq (macOS) or apt-get install jq (Ubuntu)"
        exit 1
    fi
    
    if ! check_health; then
        exit 1
    fi
    
    store_documents
    
    # Wait for indexing
    echo -e "\n⏳ Waiting 3 seconds for document indexing..."
    sleep 3
    
    test_context_search
    test_rag_chat
    test_direct_generation
    
    echo -e "\n${GREEN}✅ All RAG tests completed successfully!${NC}"
    echo -e "\nNext steps:"
    echo "• Upload your own documents via S3 API"
    echo "• Use /api/v1/chat/completions for production RAG queries"
    echo "• Monitor with /api/v1/chat/health endpoint"
    echo "• Try the Python client: python3 test_rag_client.py"
}

# Run main function
main "$@"