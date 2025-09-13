# 8fs Client Usage Examples

This document shows how to use your 8fs S3-compatible server with various client libraries and tools.

## Server Configuration

First, start your 8fs server:

```bash
# Default credentials
./bin/8fs

# Or with custom credentials
DEFAULT_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE \
DEFAULT_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
./bin/8fs
```

Server will be running on `http://localhost:8080`

---

## Python Examples

### Using boto3 (AWS SDK for Python)

```python
import boto3
from botocore.config import Config

# Configure boto3 client
s3_client = boto3.client(
    's3',
    endpoint_url='http://localhost:8080',
    aws_access_key_id='AKIAIOSFODNN7EXAMPLE',
    aws_secret_access_key='wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
    region_name='us-east-1',
    config=Config(
        signature_version='s3v4',
        s3={'addressing_style': 'path'}  # Use path-style addressing
    )
)

# Create a bucket
bucket_name = 'my-test-bucket'
s3_client.create_bucket(Bucket=bucket_name)
print(f"Created bucket: {bucket_name}")

# Upload a file
file_content = b"Hello from 8fs!"
s3_client.put_object(
    Bucket=bucket_name,
    Key='hello.txt',
    Body=file_content,
    ContentType='text/plain'
)
print("Uploaded hello.txt")

# Download a file
response = s3_client.get_object(Bucket=bucket_name, Key='hello.txt')
downloaded_content = response['Body'].read()
print(f"Downloaded content: {downloaded_content.decode()}")

# List objects in bucket
response = s3_client.list_objects_v2(Bucket=bucket_name)
if 'Contents' in response:
    for obj in response['Contents']:
        print(f"Object: {obj['Key']}, Size: {obj['Size']}")

# Delete object
s3_client.delete_object(Bucket=bucket_name, Key='hello.txt')
print("Deleted hello.txt")

# Delete bucket
s3_client.delete_bucket(Bucket=bucket_name)
print(f"Deleted bucket: {bucket_name}")
```

### Using requests library (Direct HTTP)

```python
import requests
import hashlib
import hmac
import base64
from datetime import datetime
import urllib.parse

class Simple8fsClient:
    def __init__(self, endpoint, access_key, secret_key):
        self.endpoint = endpoint.rstrip('/')
        self.access_key = access_key
        self.secret_key = secret_key
    
    def _sign_request(self, method, path, headers, body=b''):
        # Simple signature for demonstration
        # In production, use proper AWS Signature v4
        date = datetime.utcnow().strftime('%a, %d %b %Y %H:%M:%S GMT')
        headers['Date'] = date
        headers['Authorization'] = f'AWS {self.access_key}:signature'
        return headers
    
    def create_bucket(self, bucket_name):
        url = f"{self.endpoint}/{bucket_name}"
        headers = {'Content-Type': 'application/xml'}
        headers = self._sign_request('PUT', f'/{bucket_name}', headers)
        
        response = requests.put(url, headers=headers)
        return response.status_code == 200
    
    def put_object(self, bucket_name, key, content):
        url = f"{self.endpoint}/{bucket_name}/{key}"
        headers = {'Content-Type': 'application/octet-stream'}
        headers = self._sign_request('PUT', f'/{bucket_name}/{key}', headers, content)
        
        response = requests.put(url, data=content, headers=headers)
        return response.status_code == 200
    
    def get_object(self, bucket_name, key):
        url = f"{self.endpoint}/{bucket_name}/{key}"
        headers = {}
        headers = self._sign_request('GET', f'/{bucket_name}/{key}', headers)
        
        response = requests.get(url, headers=headers)
        if response.status_code == 200:
            return response.content
        return None

# Usage
client = Simple8fsClient(
    'http://localhost:8080',
    'AKIAIOSFODNN7EXAMPLE',
    'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'
)

# Create bucket and upload file
client.create_bucket('test-bucket')
client.put_object('test-bucket', 'test.txt', b'Hello World!')
content = client.get_object('test-bucket', 'test.txt')
print(f"Retrieved: {content.decode()}")
```

---

## Node.js Examples

### Using AWS SDK v3

```javascript
import { 
    S3Client, 
    CreateBucketCommand, 
    PutObjectCommand, 
    GetObjectCommand,
    ListObjectsV2Command,
    DeleteObjectCommand,
    DeleteBucketCommand 
} from "@aws-sdk/client-s3";

// Configure S3 client
const s3Client = new S3Client({
    endpoint: "http://localhost:8080",
    region: "us-east-1",
    credentials: {
        accessKeyId: "AKIAIOSFODNN7EXAMPLE",
        secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    },
    forcePathStyle: true, // Use path-style URLs
});

async function demo() {
    const bucketName = 'nodejs-test-bucket';
    
    try {
        // Create bucket
        await s3Client.send(new CreateBucketCommand({ 
            Bucket: bucketName 
        }));
        console.log(`Created bucket: ${bucketName}`);
        
        // Upload file
        const fileContent = "Hello from Node.js!";
        await s3Client.send(new PutObjectCommand({
            Bucket: bucketName,
            Key: 'hello.txt',
            Body: fileContent,
            ContentType: 'text/plain'
        }));
        console.log('Uploaded hello.txt');
        
        // Download file
        const response = await s3Client.send(new GetObjectCommand({
            Bucket: bucketName,
            Key: 'hello.txt'
        }));
        const downloadedContent = await response.Body.transformToString();
        console.log(`Downloaded: ${downloadedContent}`);
        
        // List objects
        const listResponse = await s3Client.send(new ListObjectsV2Command({
            Bucket: bucketName
        }));
        if (listResponse.Contents) {
            listResponse.Contents.forEach(obj => {
                console.log(`Object: ${obj.Key}, Size: ${obj.Size}`);
            });
        }
        
        // Clean up
        await s3Client.send(new DeleteObjectCommand({
            Bucket: bucketName,
            Key: 'hello.txt'
        }));
        await s3Client.send(new DeleteBucketCommand({
            Bucket: bucketName
        }));
        console.log('Cleanup completed');
        
    } catch (error) {
        console.error('Error:', error);
    }
}

demo();
```

### Using axios (Direct HTTP)

```javascript
import axios from 'axios';

class Simple8fsClient {
    constructor(endpoint, accessKey, secretKey) {
        this.endpoint = endpoint.replace(/\/$/, '');
        this.accessKey = accessKey;
        this.secretKey = secretKey;
    }
    
    getHeaders(method, path) {
        // Simplified headers - in production use proper AWS Signature v4
        return {
            'Authorization': `AWS ${this.accessKey}:signature`,
            'Date': new Date().toUTCString(),
            'Content-Type': 'application/octet-stream'
        };
    }
    
    async createBucket(bucketName) {
        const url = `${this.endpoint}/${bucketName}`;
        const headers = this.getHeaders('PUT', `/${bucketName}`);
        
        try {
            const response = await axios.put(url, '', { headers });
            return response.status === 200;
        } catch (error) {
            console.error('Error creating bucket:', error.message);
            return false;
        }
    }
    
    async putObject(bucketName, key, content) {
        const url = `${this.endpoint}/${bucketName}/${key}`;
        const headers = this.getHeaders('PUT', `/${bucketName}/${key}`);
        
        try {
            const response = await axios.put(url, content, { headers });
            return response.status === 200;
        } catch (error) {
            console.error('Error putting object:', error.message);
            return false;
        }
    }
    
    async getObject(bucketName, key) {
        const url = `${this.endpoint}/${bucketName}/${key}`;
        const headers = this.getHeaders('GET', `/${bucketName}/${key}`);
        
        try {
            const response = await axios.get(url, { headers });
            return response.data;
        } catch (error) {
            console.error('Error getting object:', error.message);
            return null;
        }
    }
}

// Usage
const client = new Simple8fsClient(
    'http://localhost:8080',
    'AKIAIOSFODNN7EXAMPLE',
    'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'
);

async function demo() {
    await client.createBucket('js-test-bucket');
    await client.putObject('js-test-bucket', 'test.txt', 'Hello from JavaScript!');
    const content = await client.getObject('js-test-bucket', 'test.txt');
    console.log('Retrieved:', content);
}

demo();
```

---

## cURL Examples

### Basic Operations

```bash
# Set credentials
ACCESS_KEY="AKIAIOSFODNN7EXAMPLE"
SECRET_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
ENDPOINT="http://localhost:8080"

# List buckets (requires proper AWS Signature v4)
curl -X GET "$ENDPOINT/" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=$ACCESS_KEY/$(date +%Y%m%d)/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"

# Create bucket
curl -X PUT "$ENDPOINT/test-bucket" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=$ACCESS_KEY/$(date +%Y%m%d)/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"

# Upload file
curl -X PUT "$ENDPOINT/test-bucket/hello.txt" \
  -H "Content-Type: text/plain" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=$ACCESS_KEY/$(date +%Y%m%d)/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy" \
  -d "Hello from cURL!"

# Download file
curl -X GET "$ENDPOINT/test-bucket/hello.txt" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=$ACCESS_KEY/$(date +%Y%m%d)/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"

# List objects in bucket
curl -X GET "$ENDPOINT/test-bucket" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=$ACCESS_KEY/$(date +%Y%m%d)/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"

# Delete object
curl -X DELETE "$ENDPOINT/test-bucket/hello.txt" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=$ACCESS_KEY/$(date +%Y%m%d)/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"

# Delete bucket
curl -X DELETE "$ENDPOINT/test-bucket" \
  -H "Authorization: AWS4-HMAC-SHA256 Credential=$ACCESS_KEY/$(date +%Y%m%d)/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy"
```

---

## AWS CLI Configuration

You can use the standard AWS CLI with your 8fs server:

```bash
# Configure AWS CLI profile
aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE --profile 8fs
aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --profile 8fs
aws configure set region us-east-1 --profile 8fs

# Use with endpoint URL
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls
aws --profile 8fs --endpoint-url http://localhost:8080 s3 mb s3://test-bucket
aws --profile 8fs --endpoint-url http://localhost:8080 s3 cp README.md s3://test-bucket/
aws --profile 8fs --endpoint-url http://localhost:8080 s3 ls s3://test-bucket/
```

---

## MinIO Client (mc)

```bash
# Add your 8fs server as an alias
mc alias set 8fs http://localhost:8080 AKIAIOSFODNN7EXAMPLE wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# Use MinIO client commands
mc ls 8fs
mc mb 8fs/test-bucket
mc cp README.md 8fs/test-bucket/
mc ls 8fs/test-bucket
mc rm 8fs/test-bucket/README.md
mc rb 8fs/test-bucket
```

---

## Go Example

```go
package main

import (
    "bytes"
    "fmt"
    "log"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

func main() {
    // Configure AWS session
    sess, err := session.NewSession(&aws.Config{
        Endpoint:         aws.String("http://localhost:8080"),
        Region:           aws.String("us-east-1"),
        Credentials:      credentials.NewStaticCredentials("AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", ""),
        S3ForcePathStyle: aws.Bool(true),
    })
    if err != nil {
        log.Fatal(err)
    }
    
    svc := s3.New(sess)
    bucketName := "go-test-bucket"
    
    // Create bucket
    _, err = svc.CreateBucket(&s3.CreateBucketInput{
        Bucket: aws.String(bucketName),
    })
    if err != nil {
        log.Printf("Error creating bucket: %v", err)
    } else {
        fmt.Printf("Created bucket: %s\n", bucketName)
    }
    
    // Put object
    content := "Hello from Go!"
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String("hello.txt"),
        Body:   bytes.NewReader([]byte(content)),
    })
    if err != nil {
        log.Printf("Error putting object: %v", err)
    } else {
        fmt.Println("Uploaded hello.txt")
    }
    
    // Get object
    result, err := svc.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String("hello.txt"),
    })
    if err != nil {
        log.Printf("Error getting object: %v", err)
    } else {
        buf := new(bytes.Buffer)
        buf.ReadFrom(result.Body)
        fmt.Printf("Downloaded: %s\n", buf.String())
    }
}
```

---

## Environment Configuration

For easier client setup, you can set environment variables:

```bash
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_ENDPOINT_URL=http://localhost:8080
export AWS_DEFAULT_REGION=us-east-1
```

Then most AWS SDK clients will automatically pick up these credentials.

---

## Testing Your Setup

To verify your 8fs server is working correctly:

```bash
# 1. Health check
curl http://localhost:8080/healthz

# 2. List buckets (should return empty list initially)
curl -H "Authorization: AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/$(date +%Y%m%d)/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=dummy" http://localhost:8080/

# 3. Check metrics
curl http://localhost:8080/metrics
```

---

## Production Considerations

1. **HTTPS**: Use a reverse proxy (nginx/caddy) for HTTPS in production
2. **Authentication**: Consider implementing proper AWS Signature v4 validation
3. **Monitoring**: Use the `/metrics` endpoint with Prometheus
4. **Backup**: Regularly backup your storage directory
5. **Performance**: Monitor using the structured logs and metrics

---

## Troubleshooting

### Common Issues:

1. **Connection refused**: Make sure 8fs server is running on port 8080
2. **Authorization errors**: Verify your access key and secret key match the server configuration
3. **Path style**: Most clients need `forcePathStyle: true` or `s3ForcePathStyle: true`
4. **Region**: Use `us-east-1` as the region

### Debug Mode:

Start the server with debug logging:
```bash
GIN_MODE=debug ./bin/8fs
```

Check the server logs for detailed request/response information.
