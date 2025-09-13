#!/usr/bin/env python3
"""
8fs Python Client Example
This script demonstrates how to use the 8fs S3-compatible server with boto3.

Usage:
    pip install boto3
    python3 test_8fs_client.py
"""

import boto3
import io
import json
from botocore.config import Config
from botocore.exceptions import ClientError

def main():
    print("🚀 8fs Python Client Test")
    print("=" * 40)
    
    # Configure boto3 client for 8fs
    s3_client = boto3.client(
        's3',
        endpoint_url='http://localhost:8080',
        aws_access_key_id='AKIAIOSFODNN7EXAMPLE',
        aws_secret_access_key='wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
        region_name='us-east-1',
        config=Config(
            signature_version='s3v4',
            s3={
                'addressing_style': 'path'  # Use path-style addressing
            }
        )
    )
    
    bucket_name = 'python-test-bucket'
    
    try:
        print(f"📦 Creating bucket: {bucket_name}")
        s3_client.create_bucket(Bucket=bucket_name)
        print("✅ Bucket created successfully")
        
        # Test 1: Upload text file
        print("\n📝 Uploading text file...")
        text_content = "Hello from Python! 🐍\nThis is a test file for 8fs."
        s3_client.put_object(
            Bucket=bucket_name,
            Key='hello.txt',
            Body=text_content.encode('utf-8'),
            ContentType='text/plain',
            Metadata={
                'author': 'python-client',
                'test': 'true'
            }
        )
        print("✅ Text file uploaded: hello.txt")
        
        # Test 2: Upload binary file (simulate image)
        print("\n🖼️  Uploading binary file...")
        binary_content = b'\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x04\x00\x00\x00\xb5\x1c\x0c\x02\x00\x00\x00\x0bIDATx\x9cc\xfa\x00\x00\x00\x02\x00\x01\xe5\'\xde\xfc\x00\x00\x00\x00IEND\xaeB`\x82'
        s3_client.put_object(
            Bucket=bucket_name,
            Key='images/test.png',
            Body=binary_content,
            ContentType='image/png'
        )
        print("✅ Binary file uploaded: images/test.png")
        
        # Test 3: Upload JSON data
        print("\n📊 Uploading JSON data...")
        json_data = {
            'name': '8fs-test',
            'version': '0.2.0',
            'features': ['s3-compatible', 'lightweight', 'golang'],
            'metrics': {
                'files_uploaded': 3,
                'total_size_mb': 0.1
            }
        }
        s3_client.put_object(
            Bucket=bucket_name,
            Key='data/config.json',
            Body=json.dumps(json_data, indent=2),
            ContentType='application/json'
        )
        print("✅ JSON data uploaded: data/config.json")
        
        # Test 4: List all objects
        print(f"\n📋 Listing objects in bucket: {bucket_name}")
        response = s3_client.list_objects_v2(Bucket=bucket_name)
        if 'Contents' in response:
            for obj in response['Contents']:
                size_kb = obj['Size'] / 1024
                print(f"  📄 {obj['Key']} ({size_kb:.2f} KB) - {obj['LastModified']}")
        else:
            print("  (No objects found)")
        
        # Test 5: Download and verify text file
        print(f"\n📥 Downloading text file...")
        response = s3_client.get_object(Bucket=bucket_name, Key='hello.txt')
        downloaded_content = response['Body'].read().decode('utf-8')
        print(f"✅ Downloaded content matches: {downloaded_content == text_content}")
        print(f"   Content preview: {downloaded_content[:50]}...")
        
        # Test 6: Get object metadata
        print(f"\n🔍 Getting object metadata...")
        response = s3_client.head_object(Bucket=bucket_name, Key='hello.txt')
        print(f"  Content-Type: {response.get('ContentType', 'N/A')}")
        print(f"  Content-Length: {response.get('ContentLength', 'N/A')} bytes")
        print(f"  Last-Modified: {response.get('LastModified', 'N/A')}")
        print(f"  Metadata: {response.get('Metadata', {})}")
        
        # Test 7: Download JSON and parse
        print(f"\n📊 Downloading and parsing JSON...")
        response = s3_client.get_object(Bucket=bucket_name, Key='data/config.json')
        json_content = json.loads(response['Body'].read().decode('utf-8'))
        print(f"  Parsed JSON - Name: {json_content['name']}")
        print(f"  Features: {', '.join(json_content['features'])}")
        
        # Test 8: Test file streaming
        print(f"\n🌊 Testing file streaming...")
        large_content = "Line {}\n".format(i) * 1000  # Create a larger file
        s3_client.put_object(
            Bucket=bucket_name,
            Key='data/large.txt',
            Body=large_content,
            ContentType='text/plain'
        )
        
        # Stream download
        response = s3_client.get_object(Bucket=bucket_name, Key='data/large.txt')
        chunks = []
        for chunk in response['Body'].iter_chunks(chunk_size=1024):
            chunks.append(chunk)
        streamed_content = b''.join(chunks).decode('utf-8')
        print(f"✅ Streamed {len(streamed_content)} characters successfully")
        
        # Test 9: List buckets
        print(f"\n🪣 Listing all buckets...")
        response = s3_client.list_buckets()
        for bucket in response['Buckets']:
            print(f"  📦 {bucket['Name']} (created: {bucket['CreationDate']})")
        
        print("\n🧹 Cleaning up...")
        # Delete all objects first
        objects_to_delete = []
        response = s3_client.list_objects_v2(Bucket=bucket_name)
        if 'Contents' in response:
            for obj in response['Contents']:
                objects_to_delete.append({'Key': obj['Key']})
            
            # Delete all objects
            s3_client.delete_objects(
                Bucket=bucket_name,
                Delete={'Objects': objects_to_delete}
            )
            print(f"✅ Deleted {len(objects_to_delete)} objects")
        
        # Delete the bucket
        s3_client.delete_bucket(Bucket=bucket_name)
        print(f"✅ Deleted bucket: {bucket_name}")
        
        print("\n🎉 All tests completed successfully!")
        print("\nYour 8fs server is working perfectly with Python boto3! ✨")
        
    except ClientError as e:
        error_code = e.response['Error']['Code']
        error_message = e.response['Error']['Message']
        print(f"❌ AWS Client Error: {error_code} - {error_message}")
        
    except Exception as e:
        print(f"❌ Unexpected Error: {str(e)}")
        
    print("\n" + "=" * 40)
    print("Test completed!")

if __name__ == '__main__':
    print("Make sure your 8fs server is running on http://localhost:8080")
    print("Start it with: ./bin/8fs")
    print()
    
    try:
        import boto3
        main()
    except ImportError:
        print("❌ boto3 not found. Install it with:")
        print("   pip install boto3")
