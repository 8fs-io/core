#!/usr/bin/env node

/**
 * 8fs Node.js Client Example
 * This script demonstrates how to use the 8fs S3-compatible server with AWS SDK v3.
 * 
 * Usage:
 *   npm install @aws-sdk/client-s3
 *   node test_nodejs_client.js
 */

import { 
    S3Client, 
    CreateBucketCommand, 
    PutObjectCommand, 
    GetObjectCommand,
    ListObjectsV2Command,
    HeadObjectCommand,
    DeleteObjectCommand,
    DeleteBucketCommand,
    ListBucketsCommand
} from "@aws-sdk/client-s3";

// Configure S3 client for 8fs
const s3Client = new S3Client({
    endpoint: "http://localhost:8080",
    region: "us-east-1",
    credentials: {
        accessKeyId: "AKIAIOSFODNN7EXAMPLE",
        secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    },
    forcePathStyle: true, // Use path-style URLs
});

async function main() {
    console.log("🚀 8fs Node.js Client Test");
    console.log("=".repeat(40));
    
    const bucketName = 'nodejs-test-bucket';
    
    try {
        // Test 1: Create bucket
        console.log(`📦 Creating bucket: ${bucketName}`);
        await s3Client.send(new CreateBucketCommand({ 
            Bucket: bucketName 
        }));
        console.log("✅ Bucket created successfully");
        
        // Test 2: Upload text file
        console.log("\n📝 Uploading text file...");
        const textContent = "Hello from Node.js! 🟢\nThis is a test file for 8fs.";
        await s3Client.send(new PutObjectCommand({
            Bucket: bucketName,
            Key: 'hello.txt',
            Body: textContent,
            ContentType: 'text/plain',
            Metadata: {
                'author': 'nodejs-client',
                'test': 'true'
            }
        }));
        console.log("✅ Text file uploaded: hello.txt");
        
        // Test 3: Upload JSON data
        console.log("\n📊 Uploading JSON data...");
        const jsonData = {
            name: '8fs-nodejs-test',
            version: '0.2.0',
            runtime: 'Node.js',
            features: ['s3-compatible', 'async/await', 'AWS SDK v3'],
            timestamp: new Date().toISOString()
        };
        await s3Client.send(new PutObjectCommand({
            Bucket: bucketName,
            Key: 'data/config.json',
            Body: JSON.stringify(jsonData, null, 2),
            ContentType: 'application/json'
        }));
        console.log("✅ JSON data uploaded: data/config.json");
        
        // Test 4: Upload binary data (simulate small image)
        console.log("\n🖼️  Uploading binary data...");
        const binaryData = Buffer.from([
            0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
            0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
            0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
            0x08, 0x04, 0x00, 0x00, 0x00, 0xB5, 0x1C, 0x0C
        ]);
        await s3Client.send(new PutObjectCommand({
            Bucket: bucketName,
            Key: 'images/tiny.png',
            Body: binaryData,
            ContentType: 'image/png'
        }));
        console.log("✅ Binary data uploaded: images/tiny.png");
        
        // Test 5: List objects
        console.log(`\n📋 Listing objects in bucket: ${bucketName}`);
        const listResponse = await s3Client.send(new ListObjectsV2Command({
            Bucket: bucketName
        }));
        
        if (listResponse.Contents && listResponse.Contents.length > 0) {
            listResponse.Contents.forEach(obj => {
                const sizeKB = (obj.Size / 1024).toFixed(2);
                console.log(`  📄 ${obj.Key} (${sizeKB} KB) - ${obj.LastModified}`);
            });
        } else {
            console.log("  (No objects found)");
        }
        
        // Test 6: Download and verify text file
        console.log(`\n📥 Downloading text file...`);
        const getResponse = await s3Client.send(new GetObjectCommand({
            Bucket: bucketName,
            Key: 'hello.txt'
        }));
        
        const downloadedContent = await getResponse.Body.transformToString();
        const contentMatches = downloadedContent === textContent;
        console.log(`✅ Downloaded content matches: ${contentMatches}`);
        console.log(`   Content preview: ${downloadedContent.substring(0, 50)}...`);
        
        // Test 7: Get object metadata
        console.log(`\n🔍 Getting object metadata...`);
        const headResponse = await s3Client.send(new HeadObjectCommand({
            Bucket: bucketName,
            Key: 'hello.txt'
        }));
        
        console.log(`  Content-Type: ${headResponse.ContentType || 'N/A'}`);
        console.log(`  Content-Length: ${headResponse.ContentLength || 'N/A'} bytes`);
        console.log(`  Last-Modified: ${headResponse.LastModified || 'N/A'}`);
        console.log(`  Metadata:`, headResponse.Metadata || {});
        
        // Test 8: Download and parse JSON
        console.log(`\n📊 Downloading and parsing JSON...`);
        const jsonResponse = await s3Client.send(new GetObjectCommand({
            Bucket: bucketName,
            Key: 'data/config.json'
        }));
        
        const jsonContent = JSON.parse(await jsonResponse.Body.transformToString());
        console.log(`  Parsed JSON - Name: ${jsonContent.name}`);
        console.log(`  Runtime: ${jsonContent.runtime}`);
        console.log(`  Features: ${jsonContent.features.join(', ')}`);
        
        // Test 9: Upload large text file for streaming test
        console.log(`\n🌊 Testing with larger file...`);
        const largeContent = Array.from({length: 1000}, (_, i) => `Line ${i + 1}: This is a test line for streaming.\n`).join('');
        await s3Client.send(new PutObjectCommand({
            Bucket: bucketName,
            Key: 'data/large.txt',
            Body: largeContent,
            ContentType: 'text/plain'
        }));
        
        const largeResponse = await s3Client.send(new GetObjectCommand({
            Bucket: bucketName,
            Key: 'data/large.txt'
        }));
        const largeDownloaded = await largeResponse.Body.transformToString();
        console.log(`✅ Large file test: ${largeDownloaded.length} characters downloaded`);
        
        // Test 10: List all buckets
        console.log(`\n🪣 Listing all buckets...`);
        const bucketsResponse = await s3Client.send(new ListBucketsCommand({}));
        if (bucketsResponse.Buckets) {
            bucketsResponse.Buckets.forEach(bucket => {
                console.log(`  📦 ${bucket.Name} (created: ${bucket.CreationDate})`);
            });
        }
        
        // Cleanup
        console.log("\n🧹 Cleaning up...");
        
        // Delete all objects first
        if (listResponse.Contents) {
            for (const obj of listResponse.Contents) {
                await s3Client.send(new DeleteObjectCommand({
                    Bucket: bucketName,
                    Key: obj.Key
                }));
            }
            // Also delete the large file we created later
            await s3Client.send(new DeleteObjectCommand({
                Bucket: bucketName,
                Key: 'data/large.txt'
            }));
            
            console.log(`✅ Deleted ${listResponse.Contents.length + 1} objects`);
        }
        
        // Delete the bucket
        await s3Client.send(new DeleteBucketCommand({
            Bucket: bucketName
        }));
        console.log(`✅ Deleted bucket: ${bucketName}`);
        
        console.log("\n🎉 All tests completed successfully!");
        console.log("\nYour 8fs server is working perfectly with Node.js AWS SDK v3! ✨");
        
    } catch (error) {
        console.error("❌ Error occurred:");
        
        if (error.name === 'NoSuchBucket') {
            console.error("   Bucket does not exist");
        } else if (error.name === 'BucketAlreadyExists') {
            console.error("   Bucket already exists");
        } else if (error.$metadata) {
            console.error(`   ${error.name}: ${error.message}`);
            console.error(`   HTTP Status: ${error.$metadata.httpStatusCode}`);
        } else {
            console.error(`   ${error.message}`);
        }
        
        process.exit(1);
    }
    
    console.log("\n" + "=".repeat(40));
    console.log("Test completed!");
}

// Check if running as main module
if (import.meta.url === `file://${process.argv[1]}`) {
    console.log("Make sure your 8fs server is running on http://localhost:8080");
    console.log("Start it with: ./bin/8fs");
    console.log();
    
    main().catch(error => {
        if (error.code === 'MODULE_NOT_FOUND' && error.message.includes('@aws-sdk/client-s3')) {
            console.error("❌ AWS SDK not found. Install it with:");
            console.error("   npm install @aws-sdk/client-s3");
        } else {
            console.error("❌ Unexpected error:", error.message);
        }
        process.exit(1);
    });
}

export { main };
