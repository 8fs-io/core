#!/bin/bash

# 8fs Client Test Runner
# This script demonstrates how to test various S3 client libraries with 8fs

set -e

echo "🚀 8fs Client Testing Suite"
echo "============================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if 8fs is running
check_server() {
    echo -e "${BLUE}🔍 Checking if 8fs server is running...${NC}"
    if curl -s http://localhost:8080/healthz > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 8fs server is running${NC}"
    else
        echo -e "${RED}❌ 8fs server is not running on port 8080${NC}"
        echo -e "${YELLOW}💡 Start the server with: ./bin/8fs${NC}"
        exit 1
    fi
}

# Test basic curl operations
test_curl() {
    echo -e "\n${BLUE}🌐 Testing basic cURL operations...${NC}"
    
    # Health check
    echo -n "  Health check: "
    if curl -s http://localhost:8080/healthz | grep -q "status.*ok"; then
        echo -e "${GREEN}✅ Pass${NC}"
    else
        echo -e "${RED}❌ Fail${NC}"
        return 1
    fi
    
    # Metrics
    echo -n "  Metrics endpoint: "
    if curl -s http://localhost:8080/metrics | grep -q "go_info"; then
        echo -e "${GREEN}✅ Pass${NC}"
    else
        echo -e "${RED}❌ Fail${NC}"
        return 1
    fi
    
    # Basic S3 operations (simplified, no proper signatures)
    echo -n "  List buckets: "
    if curl -s -w "%{http_code}" http://localhost:8080/ | grep -q "200"; then
        echo -e "${GREEN}✅ Pass${NC}"
    else
        echo -e "${RED}❌ Fail${NC}"
        return 1
    fi
}

# Test Python client
test_python() {
    echo -e "\n${BLUE}🐍 Testing Python boto3 client...${NC}"
    
    # Check if boto3 is available
    if ! python3 -c "import boto3" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  boto3 not found. Install with: pip install boto3${NC}"
        echo -e "${YELLOW}   Skipping Python test...${NC}"
        return 0
    fi
    
    if python3 test_python_client.py; then
        echo -e "${GREEN}✅ Python client test passed${NC}"
    else
        echo -e "${RED}❌ Python client test failed${NC}"
        return 1
    fi
}

# Test Node.js client
test_nodejs() {
    echo -e "\n${BLUE}🟢 Testing Node.js AWS SDK v3 client...${NC}"
    
    # Check if Node.js is available
    if ! command -v node >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Node.js not found. Skipping Node.js test...${NC}"
        return 0
    fi
    
    # Check if AWS SDK is available
    if ! node -e "import('@aws-sdk/client-s3')" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  @aws-sdk/client-s3 not found. Install with: npm install @aws-sdk/client-s3${NC}"
        echo -e "${YELLOW}   Skipping Node.js test...${NC}"
        return 0
    fi
    
    if node test_nodejs_client.js; then
        echo -e "${GREEN}✅ Node.js client test passed${NC}"
    else
        echo -e "${RED}❌ Node.js client test failed${NC}"
        return 1
    fi
}

# Test AWS CLI if available
test_aws_cli() {
    echo -e "\n${BLUE}⚡ Testing AWS CLI...${NC}"
    
    if ! command -v aws >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  AWS CLI not found. Skipping AWS CLI test...${NC}"
        return 0
    fi
    
    # Configure temporary profile
    aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE --profile 8fs-test
    aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --profile 8fs-test
    aws configure set region us-east-1 --profile 8fs-test
    
    echo -n "  List buckets: "
    if aws --profile 8fs-test --endpoint-url http://localhost:8080 s3 ls >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Pass${NC}"
    else
        echo -e "${RED}❌ Fail${NC}"
        return 1
    fi
    
    echo -n "  Create bucket: "
    if aws --profile 8fs-test --endpoint-url http://localhost:8080 s3 mb s3://cli-test-bucket >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Pass${NC}"
        
        echo -n "  Delete bucket: "
        if aws --profile 8fs-test --endpoint-url http://localhost:8080 s3 rb s3://cli-test-bucket >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Pass${NC}"
        else
            echo -e "${RED}❌ Fail${NC}"
        fi
    else
        echo -e "${RED}❌ Fail${NC}"
        return 1
    fi
    
    # Clean up profile
    aws configure --profile 8fs-test remove aws_access_key_id 2>/dev/null || true
    aws configure --profile 8fs-test remove aws_secret_access_key 2>/dev/null || true
    aws configure --profile 8fs-test remove region 2>/dev/null || true
}

# Test MinIO client if available
test_minio_cli() {
    echo -e "\n${BLUE}🗂️  Testing MinIO Client (mc)...${NC}"
    
    if ! command -v mc >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  MinIO client (mc) not found. Skipping mc test...${NC}"
        return 0
    fi
    
    # Configure alias
    mc alias set 8fs-test http://localhost:8080 AKIAIOSFODNN7EXAMPLE wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY >/dev/null 2>&1
    
    echo -n "  List buckets: "
    if mc ls 8fs-test >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Pass${NC}"
    else
        echo -e "${RED}❌ Fail${NC}"
        return 1
    fi
    
    echo -n "  Create/delete bucket: "
    if mc mb 8fs-test/mc-test-bucket >/dev/null 2>&1 && mc rb 8fs-test/mc-test-bucket >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Pass${NC}"
    else
        echo -e "${RED}❌ Fail${NC}"
        return 1
    fi
    
    # Clean up alias
    mc alias rm 8fs-test >/dev/null 2>&1 || true
}

# Performance test
test_performance() {
    echo -e "\n${BLUE}⚡ Running basic performance test...${NC}"
    
    echo -n "  10 concurrent health checks: "
    start_time=$(date +%s%N)
    for i in {1..10}; do
        curl -s http://localhost:8080/healthz >/dev/null &
    done
    wait
    end_time=$(date +%s%N)
    duration=$((($end_time - $start_time) / 1000000)) # Convert to milliseconds
    
    echo -e "${GREEN}✅ Completed in ${duration}ms${NC}"
    
    if [ $duration -lt 1000 ]; then
        echo -e "     ${GREEN}🚀 Excellent performance!${NC}"
    elif [ $duration -lt 3000 ]; then
        echo -e "     ${YELLOW}⚡ Good performance${NC}"
    else
        echo -e "     ${RED}⚠️  Performance could be improved${NC}"
    fi
}

# Show server info
show_server_info() {
    echo -e "\n${BLUE}📊 Server Information${NC}"
    echo "=================================="
    
    echo -n "Server Status: "
    if curl -s http://localhost:8080/healthz | jq -r '.status' 2>/dev/null; then
        echo -e "${GREEN}✅ Healthy${NC}"
    else
        echo -e "${RED}❌ Unhealthy${NC}"
    fi
    
    echo "Health Check Response:"
    curl -s http://localhost:8080/healthz | jq . 2>/dev/null || curl -s http://localhost:8080/healthz
    
    echo -e "\nKey Metrics:"
    curl -s http://localhost:8080/metrics | grep -E "(http_requests_total|buckets_total|storage_bytes_total)" | head -3
}

# Main test runner
main() {
    echo "Starting comprehensive 8fs client compatibility tests..."
    echo
    
    # Check prerequisites
    check_server
    
    # Run tests
    test_curl
    test_python
    test_nodejs
    test_aws_cli
    test_minio_cli
    test_performance
    
    # Show server info
    show_server_info
    
    echo
    echo -e "${GREEN}🎉 All available tests completed!${NC}"
    echo
    echo "Your 8fs server is compatible with:"
    echo "  ✅ Direct HTTP/cURL"
    echo "  ✅ Python boto3 (if installed)"
    echo "  ✅ Node.js AWS SDK v3 (if installed)"
    echo "  ✅ AWS CLI (if installed)"
    echo "  ✅ MinIO client (if installed)"
    echo
    echo -e "${BLUE}💡 Tips:${NC}"
    echo "  • Install missing clients to test full compatibility"
    echo "  • Check CLIENT_EXAMPLES.md for detailed usage examples"
    echo "  • Monitor server with: curl http://localhost:8080/metrics"
}

# Run the tests
main "$@"
