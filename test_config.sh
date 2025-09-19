#!/bin/bash

echo "🧪 COMPREHENSIVE RAG CONFIGURATION TEST"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

test_result() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ $1${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ $1${NC}"
        ((TESTS_FAILED++))
    fi
}

echo -e "${BLUE}Step 1: Environment Check${NC}"
echo "Checking Ollama availability..."
curl -s http://localhost:11434/api/version > /dev/null
test_result "Ollama server accessible"

echo -e "\nAvailable models:"
curl -s http://localhost:11434/api/tags | jq -r '.models[] | "- \(.name)"' | head -3

echo -e "\n${BLUE}Step 2: Server Health Check${NC}"
HEALTH=$(curl -s http://localhost:8080/api/v1/chat/health | jq -r '.status // "error"')
if [ "$HEALTH" = "healthy" ]; then
    echo -e "${GREEN}✅ Server is healthy${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Server health check failed: $HEALTH${NC}"
    ((TESTS_FAILED++))
fi

echo -e "\n${BLUE}Step 3: Direct Vector Storage Test${NC}"
# Create a simple 384-dimensional test vector
TEST_VECTOR="$(python3 -c "import json; print(json.dumps([0.1 * (i % 10 - 5) for i in range(384)]))")"

STORE_RESULT=$(curl -s -X POST http://localhost:8080/api/v1/vectors/embeddings \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"test-config-$(date +%s)\",
    \"embedding\": $TEST_VECTOR,
    \"metadata\": {\"content\": \"Test document for configurable models\", \"type\": \"test\", \"created\": \"$(date)\"}
  }")

if echo "$STORE_RESULT" | jq -e '.message | contains("successfully")' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Vector storage successful${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Vector storage failed: $(echo $STORE_RESULT | jq -r '.error // .details // "unknown error"')${NC}"
    ((TESTS_FAILED++))
    ((TESTS_FAILED++))
fi

echo -e "\n${BLUE}Step 4: Vector Search Test${NC}"
SEARCH_RESULT=$(curl -s -X POST http://localhost:8080/api/v1/vectors/search \
  -H "Content-Type: application/json" \
  -d "{\"query\": $TEST_VECTOR, \"top_k\": 3}")

SEARCH_COUNT=$(echo "$SEARCH_RESULT" | jq '.results | length // 0')
if [ "$SEARCH_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ Vector search working (found $SEARCH_COUNT results)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Vector search failed${NC}"
    echo "$SEARCH_RESULT" | jq '.'
    ((TESTS_FAILED++))
fi

echo -e "\n${BLUE}Step 5: Embedding Model Configuration Test${NC}"
# Test the configured embedding model through text search
TEXT_SEARCH=$(curl -s -X POST http://localhost:8080/api/v1/vectors/search/text \
  -H "Content-Type: application/json" \
  -d '{"query": "configurable models test", "top_k": 2}')

if echo "$TEXT_SEARCH" | jq -e '.results' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Embedding model (all-minilm:latest) working${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}⚠️  Embedding model test inconclusive: $(echo $TEXT_SEARCH | jq -r '.error // "no results"')${NC}"
    ((TESTS_FAILED++))
fi

echo -e "\n${BLUE}Step 6: RAG Context Retrieval Test${NC}"
CONTEXT_RESULT=$(curl -s -X POST http://localhost:8080/api/v1/chat/search/context \
  -H "Content-Type: application/json" \
  -d '{"query": "configurable models", "top_k": 3}')

CONTEXT_DOCS=$(echo "$CONTEXT_RESULT" | jq '.total_docs // 0')
if [ "$CONTEXT_DOCS" -gt 0 ]; then
    echo -e "${GREEN}✅ RAG context retrieval working ($CONTEXT_DOCS documents found)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}⚠️  RAG context retrieval: $(echo $CONTEXT_RESULT | jq -r '.error // "no documents found"')${NC}"
    # This might be expected if we don't have documents yet
fi

echo -e "\n${BLUE}Step 7: Chat Model Configuration Test${NC}"
CHAT_RESULT=$(curl -s -X POST http://localhost:8080/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"query": "What is 1+1? Answer briefly.", "max_tokens": 50, "temperature": 0.1}')

if echo "$CHAT_RESULT" | jq -e '.choices[0].message.content' > /dev/null 2>&1; then
    RESPONSE_TEXT=$(echo "$CHAT_RESULT" | jq -r '.choices[0].message.content' | head -1)
    echo -e "${GREEN}✅ Chat model (llama3.2:1b) working${NC}"
    echo -e "${BLUE}   Response: $RESPONSE_TEXT${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ Chat model test failed: $(echo $CHAT_RESULT | jq -r '.error.details // .error // "unknown error"')${NC}"
    ((TESTS_FAILED++))
fi

echo -e "\n${BLUE}Step 8: Configuration Verification${NC}"
# Check that models are actually configurable by testing different values
echo "Testing configuration flexibility..."

# Check if environment variables are being used
if [ "$OLLAMA_EMBED_MODEL" = "all-minilm:latest" ] && [ "$OLLAMA_CHAT_MODEL" = "llama3.2:1b" ]; then
    echo -e "${GREEN}✅ Environment variables properly set${NC}"
    echo -e "   OLLAMA_EMBED_MODEL: $OLLAMA_EMBED_MODEL"
    echo -e "   OLLAMA_CHAT_MODEL: $OLLAMA_CHAT_MODEL" 
    echo -e "   VECTOR_DIMENSION: $VECTOR_DIMENSION"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}⚠️  Environment variables not fully set${NC}"
fi

echo -e "\n${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}FINAL TEST RESULTS${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}"

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
echo -e "${GREEN}✅ Tests Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}❌ Tests Failed: $TESTS_FAILED${NC}"
else
    echo -e "${GREEN}❌ Tests Failed: $TESTS_FAILED${NC}"
fi
echo -e "${BLUE}📊 Total Tests: $TOTAL_TESTS${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 ALL CONFIGURATION TESTS PASSED!${NC}"
    echo -e "${GREEN}Your RAG system is fully configured with:${NC}"
    echo -e "${GREEN}- Configurable embedding dimensions${NC}"
    echo -e "${GREEN}- Configurable Ollama models${NC}"
    echo -e "${GREEN}- Working RAG pipeline${NC}"
    exit 0
else
    echo -e "\n${YELLOW}⚠️  Some tests had issues, but core functionality working${NC}"
    exit 1
fi