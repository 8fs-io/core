@echo off
REM AWS CLI Integration Test Script for 8fs (Windows .bat)
REM Mirrors the POSIX script logic for Windows users

setlocal enabledelayedexpansion

echo === 8fs AWS CLI Integration Test ===

echo Checking if 8fs server is accessible...
curl -s http://localhost:8080/healthz | findstr "healthy" >nul
if errorlevel 1 (
  echo Error: 8fs server is not running or not accessible at http://localhost:8080
  echo Please start the server before running this test
  exit /b 1
)

echo Setting up AWS CLI profile for 8fs...
aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE --profile 8fs
aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --profile 8fs
aws configure set region us-east-1 --profile 8fs

for /f %%a in ('powershell -NoProfile -Command "[int][double]::Parse((Get-Date -UFormat %s))"') do set TIMESTAMP=%%a
set TEST_BUCKET=test-bucket-!TIMESTAMP!
set TEST_FILE=test-file-!TIMESTAMP!.txt
set TEST_FOLDER=test-folder-!TIMESTAMP!
set LOCAL_FOLDER=.\!TEST_FOLDER!
set LOCAL_FILE=.\!TEST_FILE!

echo Creating test file...
echo This is a test file for 8fs AWS CLI integration testing. > "!LOCAL_FILE!"

echo Creating test folder with files...
if not exist "!LOCAL_FOLDER!" mkdir "!LOCAL_FOLDER!"
echo File 1 content > "!LOCAL_FOLDER!\file1.txt"
echo File 2 content > "!LOCAL_FOLDER!\file2.txt"
echo File 3 content > "!LOCAL_FOLDER!\file3.txt"

echo Test 1: Creating bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 mb s3://!TEST_BUCKET!
if errorlevel 1 exit /b 1
echo ✓ Bucket created successfully

echo Test 2: Copying file to bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp "!LOCAL_FILE!" s3://!TEST_BUCKET!/
if errorlevel 1 exit /b 1
echo ✓ File copied successfully

echo Test 3: Listing objects in bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls s3://!TEST_BUCKET!/
if errorlevel 1 exit /b 1
echo ✓ Objects listed successfully

echo Test 4: Syncing local folder to bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 sync "!LOCAL_FOLDER!" s3://!TEST_BUCKET!//!TEST_FOLDER!/
if errorlevel 1 exit /b 1
echo ✓ Folder synced successfully

echo Verifying sync operation...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls s3://!TEST_BUCKET!//!TEST_FOLDER!/ --recursive
if errorlevel 1 exit /b 1

echo Test 5: Deleting object from bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rm s3://!TEST_BUCKET!//!TEST_FILE!
if errorlevel 1 exit /b 1
echo ✓ Object deleted successfully

echo Test 6: Deleting folder from bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rm s3://!TEST_BUCKET!//!TEST_FOLDER!/ --recursive
if errorlevel 1 exit /b 1
echo ✓ Folder deleted successfully

echo Test 7: Deleting bucket...
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rb s3://!TEST_BUCKET!
if errorlevel 1 exit /b 1
echo ✓ Bucket deleted successfully

echo Cleaning up local test files...
if exist "!LOCAL_FILE!" del "!LOCAL_FILE!"
if exist "!LOCAL_FOLDER!" rmdir /s /q "!LOCAL_FOLDER!"

echo === All AWS CLI integration tests passed! ===
endlocal