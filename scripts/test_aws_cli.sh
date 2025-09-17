@echo off
REM AWS CLI Integration Test Script for 8fs
REM This script tests common AWS CLI S3 operations against a running 8fs server

echo === 8fs AWS CLI Integration Test ===

REM Check if 8fs server is running
echo Checking if 8fs server is accessible...
curl -s http://localhost:8080/healthz | findstr "healthy" >nul
if errorlevel 1 (
    echo Error: 8fs server is not running or not accessible at http://localhost:8080
    echo Please start the server with '.\bin\8fs.exe' before running this test
    exit /b 1
)

REM Set up AWS CLI profile for 8fs (if not already configured)
echo Setting up AWS CLI profile for 8fs...
aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE --profile 8fs
aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --profile 8fs
aws configure set region us-east-1 --profile 8fs

REM Test variables
set TEST_BUCKET=test-bucket-%RANDOM%%TIME:~6,2%
set TEST_FILE=test-file-%RANDOM%%TIME:~6,2%.txt
set TEST_FOLDER=test-folder-%RANDOM%%TIME:~6,2%
set LOCAL_FOLDER=.\%TEST_FOLDER%
set LOCAL_FILE=.\%TEST_FILE%

REM Create test file
echo Creating test file...
echo This is a test file for 8fs AWS CLI integration testing. > "%LOCAL_FILE%"

REM Create test folder with some files
echo Creating test folder with files...
if not exist "%LOCAL_FOLDER%" mkdir "%LOCAL_FOLDER%"
echo File 1 content > "%LOCAL_FOLDER%\file1.txt"
echo File 2 content > "%LOCAL_FOLDER%\file2.txt"
echo File 3 content > "%LOCAL_FOLDER%\file3.txt"

REM Test 1: Create bucket
echo Test 1: Creating bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 mb s3://%TEST_BUCKET%
if errorlevel 1 exit /b 1
echo ✓ Bucket created successfully

REM Test 2: Copy file to bucket
echo Test 2: Copying file to bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp "%LOCAL_FILE%" s3://%TEST_BUCKET%/
if errorlevel 1 exit /b 1
echo ✓ File copied successfully

REM Test 3: List objects in bucket
echo Test 3: Listing objects in bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls s3://%TEST_BUCKET%/
if errorlevel 1 exit /b 1
echo ✓ Objects listed successfully

REM Test 4: Sync local folder to bucket
echo Test 4: Syncing local folder to bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 sync "%LOCAL_FOLDER%" s3://%TEST_BUCKET%/%TEST_FOLDER%/
if errorlevel 1 exit /b 1
echo ✓ Folder synced successfully

REM List objects after sync to verify
echo Verifying sync operation...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls s3://%TEST_BUCKET%/%TEST_FOLDER%/ --recursive
if errorlevel 1 exit /b 1

REM Test 5: Delete object from bucket
echo Test 5: Deleting object from bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rm s3://%TEST_BUCKET%/%TEST_FILE%
if errorlevel 1 exit /b 1
echo ✓ Object deleted successfully

REM Test 6: Delete all objects with a specific prefix (folder)
echo Test 6: Deleting folder from bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rm s3://%TEST_BUCKET%/%TEST_FOLDER%/ --recursive
if errorlevel 1 exit /b 1
echo ✓ Folder deleted successfully

REM Test 7: Delete bucket
echo Test 7: Deleting bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rb s3://%TEST_BUCKET%
if errorlevel 1 exit /b 1
echo ✓ Bucket deleted successfully

REM Cleanup local test files
echo Cleaning up local test files...
if exist "%LOCAL_FILE%" del "%LOCAL_FILE%"
if exist "%LOCAL_FOLDER%" rmdir /s /q "%LOCAL_FOLDER%"

echo === All AWS CLI integration tests passed! ===
