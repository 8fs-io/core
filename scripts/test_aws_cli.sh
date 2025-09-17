#!/usr/bin/env bash
# AWS CLI Integration Test Script for 8fs (POSIX-compliant)
# Tests common AWS CLI S3 operations against a running 8fs server
set -euo pipefail

echo "=== 8fs AWS CLI Integration Test ==="

# Check if 8fs server is running
echo "Checking if 8fs server is accessible..."
if ! curl -fsS http://localhost:8080/healthz | grep -q "healthy"; then
  echo "Error: 8fs server is not running or not accessible at http://localhost:8080" >&2
  echo "Please start the server before running this test" >&2
  exit 1
fi

# Set up AWS CLI profile for 8fs (idempotent)
echo "Setting up AWS CLI profile for 8fs..."
aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE --profile 8fs
aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --profile 8fs
aws configure set region us-east-1 --profile 8fs

# Test variables
TIMESTAMP=$(date +%s)
TEST_BUCKET="test-bucket-${TIMESTAMP}"
TEST_FILE="test-file-${TIMESTAMP}.txt"
TEST_FOLDER="test-folder-${TIMESTAMP}"
LOCAL_FOLDER="./${TEST_FOLDER}"
LOCAL_FILE="./${TEST_FILE}"

# Cleanup local artifacts on exit
cleanup() {
  echo "Cleaning up local test files..."
  rm -f "${LOCAL_FILE}" || true
  rm -rf "${LOCAL_FOLDER}" || true
}
trap cleanup EXIT

# Create test file
echo "Creating test file..."
echo "This is a test file for 8fs AWS CLI integration testing." > "${LOCAL_FILE}"

# Create test folder with some files
echo "Creating test folder with files..."
mkdir -p "${LOCAL_FOLDER}"
echo "File 1 content" > "${LOCAL_FOLDER}/file1.txt"
echo "File 2 content" > "${LOCAL_FOLDER}/file2.txt"
echo "File 3 content" > "${LOCAL_FOLDER}/file3.txt"

# Test 1: Create bucket
echo "Test 1: Creating bucket..."
aws --profile 8fs --endpoint-url http://localhost:8080 s3 mb "s3://${TEST_BUCKET}"
echo "✓ Bucket created successfully"

# Test 2: Copy file to bucket
echo "Test 2: Copying file to bucket..."
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp "${LOCAL_FILE}" "s3://${TEST_BUCKET}/"
echo "✓ File copied successfully"

# Test 3: List objects in bucket
echo "Test 3: Listing objects in bucket..."
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls "s3://${TEST_BUCKET}/"
echo "✓ Objects listed successfully"

# Test 4: Sync local folder to bucket
echo "Test 4: Syncing local folder to bucket..."
aws --profile 8fs --endpoint-url http://localhost:8080 s3 sync "${LOCAL_FOLDER}" "s3://${TEST_BUCKET}/${TEST_FOLDER}/"
echo "✓ Folder synced successfully"

# Verify sync
echo "Verifying sync operation..."
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls "s3://${TEST_BUCKET}/${TEST_FOLDER}/" --recursive

# Test 5: Delete object from bucket
echo "Test 5: Deleting object from bucket..."
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rm "s3://${TEST_BUCKET}/${TEST_FILE}"
echo "✓ Object deleted successfully"

# Test 6: Delete folder prefix
echo "Test 6: Deleting folder from bucket..."
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rm "s3://${TEST_BUCKET}/${TEST_FOLDER}/" --recursive
echo "✓ Folder deleted successfully"

# Test 7: Delete bucket
echo "Test 7: Deleting bucket..."
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rb "s3://${TEST_BUCKET}"
echo "✓ Bucket deleted successfully"

echo "=== All AWS CLI integration tests passed! ==="
