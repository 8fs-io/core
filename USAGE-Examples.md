# Example S3 Usage with 8fs Docker Container

## AWS CLI Configuration

Configure AWS CLI to use the 8fs service:

```bash
# Configure AWS CLI with test credentials
aws configure set aws_access_key_id AKIAIOSFODNN7EXAMPLE
aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
aws configure set default.region us-east-1
```

## Basic Operations

### Using AWS CLI

```bash
# Create a bucket
aws --endpoint-url=http://localhost:8080 s3 mb s3://my-test-bucket

# List buckets
aws --endpoint-url=http://localhost:8080 s3 ls

# Upload a file
echo "Hello, S3!" > test-file.txt
aws --endpoint-url=http://localhost:8080 s3 cp test-file.txt s3://my-test-bucket/

# List objects in bucket
aws --endpoint-url=http://localhost:8080 s3 ls s3://my-test-bucket/

# Download a file
aws --endpoint-url=http://localhost:8080 s3 cp s3://my-test-bucket/test-file.txt downloaded-file.txt

# Delete an object
aws --endpoint-url=http://localhost:8080 s3 rm s3://my-test-bucket/test-file.txt

# Delete a bucket
aws --endpoint-url=http://localhost:8080 s3 rb s3://my-test-bucket
```

### Using curl

```bash
# Health check
curl http://localhost:8080/healthz

# Create bucket
curl -X PUT http://localhost:8080/my-bucket

# List buckets
curl http://localhost:8080/

# Upload object (requires proper AWS v4 signature)
curl -X PUT http://localhost:8080/my-bucket/my-object -d "Hello World"

# Get object
curl http://localhost:8080/my-bucket/my-object
```

### Using Python boto3

```python
import boto3
from botocore.config import Config

# Configure boto3 client
s3_client = boto3.client(
    's3',
    endpoint_url='http://localhost:8080',
    aws_access_key_id='AKIAIOSFODNN7EXAMPLE',
    aws_secret_access_key='wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
    config=Config(signature_version='s3v4'),
    region_name='us-east-1'
)

# Create bucket
s3_client.create_bucket(Bucket='my-python-bucket')

# Upload object
s3_client.put_object(
    Bucket='my-python-bucket',
    Key='hello.txt',
    Body=b'Hello from Python!'
)

# List objects
response = s3_client.list_objects_v2(Bucket='my-python-bucket')
for obj in response.get('Contents', []):
    print(f"Object: {obj['Key']}, Size: {obj['Size']}")

# Download object
response = s3_client.get_object(Bucket='my-python-bucket', Key='hello.txt')
content = response['Body'].read()
print(f"Content: {content.decode('utf-8')}")
```

## Production Deployment

### Docker Compose for Production

Create a `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  8fs:
    image: 8fs:latest
    ports:
      - "8080:8080"
    volumes:
      - /opt/8fs/data:/app/data
      - /opt/8fs/logs:/app/logs
    environment:
      - GIN_MODE=release
    restart: always
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Behind a Reverse Proxy (nginx)

```nginx
upstream 8fs {
    server localhost:8080;
}

server {
    listen 80;
    server_name your-s3-domain.com;
    
    location / {
        proxy_pass http://8fs;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Important for large file uploads
        client_max_body_size 100M;
        proxy_request_buffering off;
    }
}
```

## Monitoring Setup

### Prometheus + Grafana

Start with monitoring:

```bash
./docker.sh run-monitoring
```

### Grafana Dashboard

Connect to Prometheus at `http://prometheus:9090` and create dashboards for:

- HTTP request rates and latencies
- S3 operation metrics
- Storage usage
- System metrics (memory, CPU)

### Key Metrics

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration
- `s3_operations_total` - S3 operations counter
- `buckets_total` - Number of buckets
- `objects_total` - Number of objects
- `storage_bytes_total` - Storage usage

## Security Considerations

1. **Change Default Credentials**: Replace the default AWS credentials
2. **TLS/HTTPS**: Use TLS termination at load balancer
3. **Network Security**: Run in private networks
4. **Resource Limits**: Set appropriate memory/CPU limits
5. **Audit Logs**: Monitor audit logs for suspicious activity

## Backup and Recovery

```bash
# Backup data directory
tar -czf 8fs-backup-$(date +%Y%m%d).tar.gz data/

# Restore data
tar -xzf 8fs-backup-20250913.tar.gz
```
