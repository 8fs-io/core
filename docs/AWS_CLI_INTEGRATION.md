# AWS CLI Integration with 8fs

This document provides detailed instructions on how to configure and use AWS CLI with your 8fs S3-compatible storage server.

## Table of Contents

- [AWS CLI Configuration](#aws-cli-configuration)
- [Authentication Setup](#authentication-setup)
- [Common Operations](#common-operations)
- [Integration Examples](#integration-examples)
- [Troubleshooting Guide](#troubleshooting-guide)
- [Limitations and Differences](#limitations-and-differences)
- [Performance Characteristics](#performance-characteristics)

## AWS CLI Configuration

To use AWS CLI with 8fs, you need to configure a profile with the appropriate credentials and endpoint URL.

### Using aws configure command

```bash
# Configure a new profile for 8fs
aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE --profile 8fs
aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --profile 8fs
aws configure set region us-east-1 --profile 8fs
```

### Environment Variables

Alternatively, you can set environment variables for temporary configuration:

```bash
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_DEFAULT_REGION=us-east-1
export AWS_ENDPOINT_URL=http://localhost:8080
```

## Authentication Setup

8fs supports AWS Signature v4 authentication. The default credentials are:

- **Access Key ID**: `AKIAIOSFODNN7EXAMPLE`
- **Secret Access Key**: `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`

These can be customized when starting the 8fs server:

```bash
DEFAULT_ACCESS_KEY=your-access-key \
DEFAULT_SECRET_KEY=your-secret-key \
./bin/8fs
```

## Common Operations

### List Buckets

```bash
# Using profile
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls

# Using environment variables
aws s3 ls
```

### Create Bucket

```bash
# Using profile
aws --profile 8fs --endpoint-url http://localhost:8080 s3 mb s3://my-bucket

# Using environment variables
aws s3 mb s3://my-bucket
```

### Upload Files

```bash
# Copy a single file
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp myfile.txt s3://my-bucket/

# Copy a file with a specific key
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp myfile.txt s3://my-bucket/myfile.txt

# Upload all files in a directory
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp mydir/ s3://my-bucket/mydir/ --recursive
```

### List Objects

```bash
# List all objects in a bucket
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls s3://my-bucket

# List objects with a specific prefix
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls s3://my-bucket/mydir/

# List objects recursively
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls s3://my-bucket --recursive
```

### Download Files

```bash
# Download a single file
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp s3://my-bucket/myfile.txt myfile.txt

# Download all files in a bucket
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp s3://my-bucket/ ./local-dir/ --recursive
```

### Sync Directories

```bash
# Sync local directory to bucket
aws --profile 8fs --endpoint-url http://localhost:8080 s3 sync ./local-folder s3://my-bucket/folder/

# Sync bucket to local directory
aws --profile 8fs --endpoint-url http://localhost:8080 s3 sync s3://my-bucket/folder/ ./local-folder

# Sync with delete option (removes files that don't exist in source)
aws --profile 8fs --endpoint-url http://localhost:8080 s3 sync ./local-folder s3://my-bucket/folder/ --delete
```

### Delete Operations

```bash
# Delete a single object
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rm s3://my-bucket/myfile.txt

# Delete all objects in a bucket (bucket must be empty)
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rb s3://my-bucket

# Delete all objects with a specific prefix
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rm s3://my-bucket/mydir/ --recursive

# Delete bucket and all its contents
aws --profile 8fs --endpoint-url http://localhost:8080 s3 rb s3://my-bucket --force
```

## Integration Examples

### .aws/config Example

```ini
[profile 8fs]
region = us-east-1
output = json
endpoint_url = http://localhost:8080
```

### .aws/credentials Example

```ini
[8fs]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

### Using with LocalStack Patterns

If you're familiar with LocalStack, you can use similar patterns with 8fs:

```bash
# LocalStack pattern
aws --endpoint-url=http://localhost:4566 s3 ls

# Equivalent 8fs pattern
aws --endpoint-url=http://localhost:8080 s3 ls
```

### Using with MinIO Patterns

MinIO users can also adapt their patterns:

```bash
# MinIO client pattern
mc alias set 8fs http://localhost:8080 AKIAIOSFODNN7EXAMPLE wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
mc ls 8fs

# AWS CLI equivalent
aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE --profile 8fs
aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --profile 8fs
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls
```

## Troubleshooting Guide

### Common Issues and Solutions

1. **Connection Refused**
   ```
   Could not connect to the endpoint URL: http://localhost:8080/
   ```
   - Ensure 8fs server is running
   - Check if the port is correct (default is 8080)
   - Verify network connectivity

2. **Authentication Errors**
   ```
   An error occurred (InvalidAccessKeyId) when calling the ListBuckets operation: The AWS Access Key Id you provided does not exist in our records.
   ```
   - Verify your access key and secret key match the server configuration
   - Check that the credentials are correctly set in your profile or environment variables

3. **SignatureDoesNotMatch Error**
   ```
   An error occurred (SignatureDoesNotMatch) when calling the PutObject operation: The request signature we calculated does not match the signature you provided.
   ```
   - Ensure your system clock is synchronized
   - Verify the secret access key is correct

4. **Bucket Name Issues**
   ```
   An error occurred (IllegalLocationConstraintException) when calling the CreateBucket operation: The unspecified location constraint is incompatible for the region specific endpoint this request was sent to.
   ```
   - Use path-style addressing with `--endpoint-url`
   - Ensure the region is set to `us-east-1`

### Debugging with cURL

You can also test 8fs operations using cURL:

```bash
# List buckets
curl -X GET "http://localhost:8080/" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20250917/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"

# Create bucket
curl -X PUT "http://localhost:8080/my-bucket" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20250917/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"

# Upload file
curl -X PUT "http://localhost:8080/my-bucket/myfile.txt" \
  -H "Content-Type: text/plain" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20250917/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy" \
  -d "Hello World!"

# Download file
curl -X GET "http://localhost:8080/my-bucket/myfile.txt" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20250917/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"

# Delete file
curl -X DELETE "http://localhost:8080/my-bucket/myfile.txt" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20250917/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"
```

### Server Logs for Debugging

To get more detailed information about requests and responses, start the server in debug mode:

```bash
GIN_MODE=debug ./bin/8fs
```

## Limitations and Differences

### Supported S3 Operations

8fs currently supports a subset of S3 operations:

- Bucket operations: `CreateBucket`, `DeleteBucket`, `ListBuckets`
- Object operations: `PutObject`, `GetObject`, `DeleteObject`, `ListObjects`, `HeadObject`

### Unsupported Features

- **Multipart uploads**: Not currently supported
- **Bucket policies**: Not implemented
- **CORS configuration**: Not supported
- **Versioning**: Not implemented
- **Lifecycle policies**: Not supported
- **Static website hosting**: Not available
- **Pre-signed URLs**: Not implemented

### Configuration Differences

- 8fs uses path-style addressing by default (similar to MinIO)
- Region is fixed to `us-east-1` but can be configured differently
- The server doesn't enforce region constraints like AWS S3

## Performance Characteristics

### Request Latency

- **Bucket operations**: <1ms (95th percentile)
- **Object operations**: <1ms (95th percentile)
- **List operations**: <5ms (95th percentile)

### Throughput

- **Uploads**: Limited by network and disk I/O
- **Downloads**: Limited by network and disk I/O
- **Concurrent operations**: Depends on system resources

### Memory Usage

- **Baseline**: ~15MB
- **Per request**: Additional memory allocated for request processing
- **Caching**: No built-in caching (objects are read/written directly to disk)

### Binary Size

- **Optimized build**: ~10MB
- **Cross-platform builds**: Varies by architecture

## Testing with Different AWS CLI Versions

8fs should work with AWS CLI versions 1.x and 2.x. To check your version:

```bash
aws --version
```

If you encounter compatibility issues:
1. Try updating to the latest AWS CLI version
2. Check if the issue persists with the environment variable configuration method
3. Verify that you're using path-style addressing with `--endpoint-url`

## Conclusion

With this documentation, you should be able to successfully use AWS CLI with your 8fs server for common S3 operations. Remember to always use the `--endpoint-url` parameter to point to your 8fs server, and ensure your credentials match those configured on the server.
