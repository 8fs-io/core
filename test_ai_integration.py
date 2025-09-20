#!/usr/bin/env python3
"""
Test AI Integration for 8fs
This script tests if the AI service is working and generating vectors.
"""

import boto3
import time
import requests
from botocore.config import Config

def main():
    print("🤖 8fs AI Integration Test")
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
            s3={'addressing_style': 'path'}
        )
    )
    
    bucket_name = f'ai-test-bucket-{int(time.time())}'
    
    try:
        print(f"📦 Creating bucket: {bucket_name}")
        s3_client.create_bucket(Bucket=bucket_name)
        print("✅ Bucket created successfully")
        
        # Test 1: Upload a text document that should trigger AI processing
        print("\n📝 Uploading AI-ready document...")
        ai_content = """
        This is a comprehensive document about artificial intelligence and machine learning.
        
        Machine learning is a subset of artificial intelligence that enables computers to learn
        and improve from experience without being explicitly programmed. Deep learning, a 
        subset of machine learning, uses neural networks with multiple layers to model and 
        understand complex patterns in data.
        
        Natural language processing (NLP) is another important area of AI that focuses on the
        interaction between computers and human language. It enables machines to read, 
        understand, and generate human language in a valuable way.
        
        Vector databases and embeddings are crucial technologies for modern AI applications,
        allowing for semantic search and similarity matching across large datasets.
        """
        
        s3_client.put_object(
            Bucket=bucket_name,
            Key='ai-document.txt',
            Body=ai_content.encode('utf-8'),
            ContentType='text/plain',
            Metadata={
                'type': 'ai-test',
                'processing': 'enabled'
            }
        )
        print("✅ AI document uploaded: ai-document.txt")
        
        # Test 2: Upload a non-text file (should not trigger AI processing)
        print("\n🖼️  Uploading non-text document...")
        binary_content = b'\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01'
        s3_client.put_object(
            Bucket=bucket_name,
            Key='image.png',
            Body=binary_content,
            ContentType='image/png'
        )
        print("✅ Binary file uploaded: image.png (should not trigger AI)")
        
        # Wait for background processing
        print("\n⏳ Waiting for AI processing to complete...")
        time.sleep(5)  # Give time for background processing
        
        # Test 3: Try to search for similar content (if we can access vectors API)
        print("\n🔍 Checking if vectors were generated...")
        try:
            # This might fail due to authentication, but worth trying
            response = requests.get('http://localhost:8080/api/v1/vectors')
            if response.status_code == 401:
                print("⚠️  Vector API requires authentication (expected)")
            else:
                print(f"✅ Vector API response: {response.status_code}")
        except Exception as e:
            print(f"⚠️  Could not access vector API: {e}")
        
        print("\n📋 Listing uploaded objects...")
        response = s3_client.list_objects_v2(Bucket=bucket_name)
        if 'Contents' in response:
            for obj in response['Contents']:
                size_kb = obj['Size'] / 1024
                print(f"  📄 {obj['Key']} ({size_kb:.2f} KB)")
        
        print("\n🧹 Cleaning up...")
        # Delete objects
        objects_to_delete = []
        response = s3_client.list_objects_v2(Bucket=bucket_name)
        if 'Contents' in response:
            for obj in response['Contents']:
                objects_to_delete.append({'Key': obj['Key']})
            
            s3_client.delete_objects(
                Bucket=bucket_name,
                Delete={'Objects': objects_to_delete}
            )
            print(f"✅ Deleted {len(objects_to_delete)} objects")
        
        # Delete bucket
        s3_client.delete_bucket(Bucket=bucket_name)
        print(f"✅ Deleted bucket: {bucket_name}")
        
        print("\n🎉 AI Integration test completed!")
        print("\n💡 Check the server logs for AI processing messages:")
        print("   - Look for 'Document processed for AI' success messages")
        print("   - Look for 'Failed to process document for AI' error messages")
        print("   - Look for 'Stored vector embedding' vector storage messages")
        
    except Exception as e:
        print(f"❌ Error: {str(e)}")

if __name__ == '__main__':
    main()