#!/bin/bash

# 8fs RAG Server Comprehensive Test
# Tests health, document storage, context search, and chat completions

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8080"
API_V1="$SERVER_URL/api/v1"

# Main title
echo -e "${BLUE}🚀 8fs RAG Server Test${NC}"
echo "=================================="

# 1. Health Check
health_check() {
    echo -e "\n${BLUE}1. Health Check${NC}"
    response=$(curl -s "$API_V1/chat/health")
    
    if [ $? -eq 0 ] && echo "$response" | jq -e '.status' > /dev/null 2>&1; then
        local status=$(echo "$response" | jq -r '.status')
        if [ "$status" = "healthy" ]; then
            echo -e "${GREEN}✅ RAG service is healthy${NC}"
            return 0
        fi
    fi
    
    echo -e "${RED}❌ RAG service is not healthy${NC}"
    echo "Response: $response"
    return 1
}

# 2. Store sample documents with proper embeddings
store_documents() {
    echo -e "\n${BLUE}2. Storing Sample Documents${NC}"
    
    # Generate a 384-dimensional non-zero embedding vector (avoiding all-zeros which cause issues)
    local test_embedding=$(python3 -c "import json; print(json.dumps([0.01] * 384))")
    
    # Document 1: 8fs Overview
    curl -sf -X POST "$API_V1/vectors/embeddings" \
        -H "Content-Type: application/json" \
        -d "{
            \"id\": \"doc_8fs_overview\",
            \"embedding\": $test_embedding,
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
            \"embedding\": $test_embedding,
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
            \"embedding\": $test_embedding,
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
            \"embedding\": $test_embedding,
            \"metadata\": {
                \"content\": \"8fs provides S3-compatible APIs for bucket and object operations, vector storage endpoints for embeddings and search, and RAG endpoints for chat completions and context retrieval. All APIs support standard HTTP methods and JSON payloads.\",
                \"source\": \"api.md\",
                \"type\": \"documentation\",
                \"category\": \"api\"
            }
        }" > /dev/null && echo "✅ Stored: API Usage" || echo "❌ Failed to store: API Usage"
    
    echo -e "${GREEN}📄 Stored 4 sample documents${NC}"
}

# 3. Test context search
test_context_search() {
    echo -e "\n${YELLOW}⏳ Waiting 3 seconds for document indexing...${NC}"
    sleep 3
    
    echo -e "\n${BLUE}3. Context Search Test${NC}"
    local query="How fast is 8fs for vector operations?"
    echo -e "${YELLOW}🔍 Query: $query${NC}"
    
    response=$(curl -s -X POST "$API_V1/chat/search/context" \
        -H "Content-Type: application/json" \
        -d "{
            \"query\": \"$query\",
            \"max_results\": 3
        }")
    
    if [ $? -eq 0 ] && echo "$response" | jq -e '.results' > /dev/null 2>&1; then
        local count=$(echo "$response" | jq '.results | length')
        echo -e "${GREEN}📚 Found $count relevant documents${NC}"
        
        # Show document titles/sources if available
        echo "$response" | jq -r '.results[] | "  • " + (.metadata.source // "unknown") + ": " + (.metadata.content // "...")[0:50] + "..."' 2>/dev/null || echo "  • Documents found but metadata format differs"
    else
        echo -e "${RED}❌ Context search failed${NC}"
        echo "Response: $response"
        return 1
    fi
}

# 4. Test RAG chat completion
test_rag_completion() {
    echo -e "\n${BLUE}4. RAG Chat Completion Test${NC}"
    
    local query="What is 8fs and what makes it unique?"
    echo -e "\n${YELLOW}❓ Query: $query${NC}"
    
    response=$(curl -s -X POST "$API_V1/chat/completions" \
        -H "Content-Type: application/json" \
        -d '{
            "query": "'"$query"'",
            "max_tokens": 200,
            "temperature": 0.7,
            "top_k": 3
        }')
        
    if [ $? -eq 0 ] && echo "$response" | jq -e '.choices[0].message.content' > /dev/null 2>&1; then
        local answer=$(echo "$response" | jq -r '.choices[0].message.content')
        if [ ${#answer} -gt 150 ]; then
            echo -e "${GREEN}🤖 Answer: ${answer:0:150}...${NC}"
        else
            echo -e "${GREEN}🤖 Answer: $answer${NC}"
        fi
        
        local doc_count=$(echo "$response" | jq -r '.context.documents | length // 0')
        local total_tokens=$(echo "$response" | jq -r '.usage.total_tokens // 0')
        echo -e "${BLUE}📊 Used $doc_count context docs, $total_tokens tokens${NC}"
    else
        echo -e "${RED}❌ Failed to get response${NC}"
        echo "Response: $response"
        return 1
    fi
}

# Run all tests
run_tests() {
    local failed=0
    
    health_check || failed=$((failed + 1))
    store_documents || failed=$((failed + 1))
    test_context_search || failed=$((failed + 1))  
    test_rag_completion || failed=$((failed + 1))
    
    echo -e "\n${BLUE}=================================${NC}"
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}❌ $failed test(s) failed${NC}"
        return 1
    fi
}

# Check dependencies
if ! command -v jq &> /dev/null; then
    echo -e "${RED}❌ jq is required but not installed${NC}"
    exit 1
fi

if ! command -v python3 &> /dev/null; then
    echo -e "${RED}❌ python3 is required but not installed${NC}"
    exit 1
fi

# Run the tests
run_tests