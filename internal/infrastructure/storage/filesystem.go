package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/8fs/8fs/internal/domain/storage"
	"github.com/8fs/8fs/pkg/errors"
	"github.com/8fs/8fs/pkg/logger"
)

// filesystemRepository implements storage.Repository using filesystem
type filesystemRepository struct {
	basePath string
	logger   logger.Logger
}

// NewFilesystemRepository creates a new filesystem-based storage repository
func NewFilesystemRepository(basePath string, logger logger.Logger) (storage.Repository, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return &filesystemRepository{
		basePath: basePath,
		logger:   logger,
	}, nil
}

// CreateBucket creates a new bucket directory
func (r *filesystemRepository) CreateBucket(ctx context.Context, bucket *storage.Bucket) error {
	bucketPath := r.bucketPath(bucket.Name)

	// Check if bucket already exists
	if _, err := os.Stat(bucketPath); !os.IsNotExist(err) {
		return errors.ErrBucketExists.WithContext("bucket", bucket.Name)
	}

	// Create bucket directory
	if err := os.MkdirAll(bucketPath, 0755); err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to create bucket directory", err)
	}

	// Create metadata directory
	metadataDir := filepath.Join(bucketPath, ".metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to create metadata directory", err)
	}

	// Save bucket metadata
	bucketMetadataPath := filepath.Join(metadataDir, "bucket.json")
	bucketData, err := json.Marshal(bucket)
	if err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to marshal bucket metadata", err)
	}

	if err := ioutil.WriteFile(bucketMetadataPath, bucketData, 0644); err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to write bucket metadata", err)
	}

	return nil
}

// DeleteBucket removes a bucket directory
func (r *filesystemRepository) DeleteBucket(ctx context.Context, name string) error {
	bucketPath := r.bucketPath(name)

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		return errors.ErrBucketNotFound.WithContext("bucket", name)
	}

	if err := os.RemoveAll(bucketPath); err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to remove bucket directory", err)
	}

	return nil
}

// GetBucket retrieves bucket information
func (r *filesystemRepository) GetBucket(ctx context.Context, name string) (*storage.Bucket, error) {
	bucketPath := r.bucketPath(name)
	bucketMetadataPath := filepath.Join(bucketPath, ".metadata", "bucket.json")

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		return nil, errors.ErrBucketNotFound.WithContext("bucket", name)
	}

	// Try to read bucket metadata
	data, err := ioutil.ReadFile(bucketMetadataPath)
	if err != nil {
		// If metadata file doesn't exist, create a basic bucket object
		if os.IsNotExist(err) {
			return &storage.Bucket{
				Name:        name,
				CreatedAt:   time.Now().UTC(), // We can't determine actual creation time
				Metadata:    make(map[string]string),
				ObjectCount: 0,
				Size:        0,
			}, nil
		}
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to read bucket metadata", err)
	}

	var bucket storage.Bucket
	if err := json.Unmarshal(data, &bucket); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to unmarshal bucket metadata", err)
	}

	// Update bucket stats
	if err := r.updateBucketStats(ctx, &bucket); err != nil {
		r.logger.Warn("Failed to update bucket stats", "bucket", name, "error", err)
	}

	return &bucket, nil
}

// ListBuckets lists all buckets
func (r *filesystemRepository) ListBuckets(ctx context.Context) ([]*storage.Bucket, error) {
	entries, err := ioutil.ReadDir(r.basePath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to read base directory", err)
	}

	var buckets []*storage.Bucket
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			bucket, err := r.GetBucket(ctx, entry.Name())
			if err != nil {
				r.logger.Warn("Failed to get bucket info", "bucket", entry.Name(), "error", err)
				continue
			}
			buckets = append(buckets, bucket)
		}
	}

	return buckets, nil
}

// BucketExists checks if a bucket exists
func (r *filesystemRepository) BucketExists(ctx context.Context, name string) (bool, error) {
	bucketPath := r.bucketPath(name)
	_, err := os.Stat(bucketPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(errors.ErrCodeInternalError, "Failed to check bucket existence", err)
	}
	return true, nil
}

// PutObject stores an object
func (r *filesystemRepository) PutObject(ctx context.Context, object *storage.Object) error {
	objectPath := r.objectPath(object.Bucket, object.Key)
	metadataPath := r.metadataPath(object.Bucket, object.Key)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(objectPath), 0755); err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to create object directory", err)
	}
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0755); err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to create metadata directory", err)
	}

	// Write object data
	if err := ioutil.WriteFile(objectPath, object.Data, 0644); err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to write object data", err)
	}

	// Write object metadata
	metadata := storage.ObjectInfo{
		Key:          object.Key,
		Size:         object.Size,
		ContentType:  object.ContentType,
		ETag:         object.ETag,
		LastModified: object.LastModified,
		Metadata:     object.Metadata,
	}

	metadataData, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to marshal object metadata", err)
	}

	if err := ioutil.WriteFile(metadataPath, metadataData, 0644); err != nil {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to write object metadata", err)
	}

	return nil
}

// GetObject retrieves an object
func (r *filesystemRepository) GetObject(ctx context.Context, bucket, key string) (*storage.Object, error) {
	objectPath := r.objectPath(bucket, key)
	metadataPath := r.metadataPath(bucket, key)

	// Check if object exists
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return nil, errors.ErrObjectNotFound.WithContext("bucket", bucket).WithContext("key", key)
	}

	// Read object data
	data, err := ioutil.ReadFile(objectPath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to read object data", err)
	}

	// Read object metadata
	metadataData, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		// If metadata doesn't exist, create basic metadata
		if os.IsNotExist(err) {
			return &storage.Object{
				Key:          key,
				Bucket:       bucket,
				Size:         int64(len(data)),
				ContentType:  "application/octet-stream",
				ETag:         "\"unknown\"",
				LastModified: time.Now().UTC(),
				Metadata:     make(map[string]string),
				Data:         data,
			}, nil
		}
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to read object metadata", err)
	}

	var objectInfo storage.ObjectInfo
	if err := json.Unmarshal(metadataData, &objectInfo); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to unmarshal object metadata", err)
	}

	return &storage.Object{
		Key:          objectInfo.Key,
		Bucket:       bucket,
		Size:         objectInfo.Size,
		ContentType:  objectInfo.ContentType,
		ETag:         objectInfo.ETag,
		LastModified: objectInfo.LastModified,
		Metadata:     objectInfo.Metadata,
		Data:         data,
	}, nil
}

// GetObjectInfo retrieves object metadata only
func (r *filesystemRepository) GetObjectInfo(ctx context.Context, bucket, key string) (*storage.ObjectInfo, error) {
	objectPath := r.objectPath(bucket, key)
	metadataPath := r.metadataPath(bucket, key)

	// Check if object exists
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return nil, errors.ErrObjectNotFound.WithContext("bucket", bucket).WithContext("key", key)
	}

	// Read object metadata
	metadataData, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		// If metadata doesn't exist, create basic metadata from file stats
		if os.IsNotExist(err) {
			stat, err := os.Stat(objectPath)
			if err != nil {
				return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to stat object file", err)
			}

			return &storage.ObjectInfo{
				Key:          key,
				Size:         stat.Size(),
				ContentType:  "application/octet-stream",
				ETag:         "\"unknown\"",
				LastModified: stat.ModTime().UTC(),
				Metadata:     make(map[string]string),
			}, nil
		}
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to read object metadata", err)
	}

	var objectInfo storage.ObjectInfo
	if err := json.Unmarshal(metadataData, &objectInfo); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to unmarshal object metadata", err)
	}

	return &objectInfo, nil
}

// DeleteObject removes an object
func (r *filesystemRepository) DeleteObject(ctx context.Context, bucket, key string) error {
	objectPath := r.objectPath(bucket, key)
	metadataPath := r.metadataPath(bucket, key)

	// Check if object exists
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return errors.ErrObjectNotFound.WithContext("bucket", bucket).WithContext("key", key)
	}

	// Remove object file
	if err := os.Remove(objectPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to remove object file", err)
	}

	// Remove metadata file
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		r.logger.Warn("Failed to remove object metadata", "bucket", bucket, "key", key, "error", err)
	}

	return nil
}

// ListObjects lists objects in a bucket
func (r *filesystemRepository) ListObjects(ctx context.Context, bucket string, opts storage.ListOptions) (*storage.ListResult, error) {
	bucketPath := r.bucketPath(bucket)

	// Check if bucket exists
	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		return nil, errors.ErrBucketNotFound.WithContext("bucket", bucket)
	}

	var objects []storage.ObjectInfo
	err := r.walkObjects(bucketPath, opts.Prefix, func(key string, info storage.ObjectInfo) {
		objects = append(objects, info)
	})
	if err != nil {
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to walk objects", err)
	}

	// Apply max keys limit
	maxKeys := opts.MaxKeys
	if maxKeys <= 0 {
		maxKeys = 1000 // Default limit
	}

	result := &storage.ListResult{
		Objects:        objects,
		CommonPrefixes: []string{}, // TODO: Implement common prefixes for delimiter
		IsTruncated:    len(objects) > maxKeys,
	}

	if result.IsTruncated {
		result.Objects = objects[:maxKeys]
		if len(objects) > maxKeys {
			result.NextMarker = objects[maxKeys].Key
		}
	}

	return result, nil
}

// ObjectExists checks if an object exists
func (r *filesystemRepository) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	objectPath := r.objectPath(bucket, key)
	_, err := os.Stat(objectPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(errors.ErrCodeInternalError, "Failed to check object existence", err)
	}
	return true, nil
}

// GetStorageStats retrieves storage statistics
func (r *filesystemRepository) GetStorageStats(ctx context.Context) (map[string]interface{}, error) {
	buckets, err := r.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	var totalObjects int64
	var totalSize int64

	for _, bucket := range buckets {
		totalObjects += bucket.ObjectCount
		totalSize += bucket.Size
	}

	stats := map[string]interface{}{
		"buckets_count":  len(buckets),
		"objects_count":  totalObjects,
		"storage_bytes":  totalSize,
		"storage_driver": "filesystem",
		"base_path":      r.basePath,
	}

	return stats, nil
}

// HealthCheck performs a health check
func (r *filesystemRepository) HealthCheck(ctx context.Context) error {
	// Check if base path is accessible
	if _, err := os.Stat(r.basePath); err != nil {
		return errors.Wrap(errors.ErrCodeServiceUnavailable, "Storage path not accessible", err)
	}

	// Try to create a test file
	testPath := filepath.Join(r.basePath, ".health_check")
	if err := ioutil.WriteFile(testPath, []byte("test"), 0644); err != nil {
		return errors.Wrap(errors.ErrCodeServiceUnavailable, "Cannot write to storage", err)
	}

	// Clean up test file
	os.Remove(testPath)

	return nil
}

// Helper methods
func (r *filesystemRepository) bucketPath(bucket string) string {
	return filepath.Join(r.basePath, bucket)
}

func (r *filesystemRepository) objectPath(bucket, object string) string {
	return filepath.Join(r.basePath, bucket, object)
}

func (r *filesystemRepository) metadataPath(bucket, object string) string {
	return filepath.Join(r.basePath, bucket, ".metadata", object+".json")
}

func (r *filesystemRepository) updateBucketStats(ctx context.Context, bucket *storage.Bucket) error {
	bucketPath := r.bucketPath(bucket.Name)

	var objectCount int64
	var totalSize int64

	err := r.walkObjects(bucketPath, "", func(key string, info storage.ObjectInfo) {
		objectCount++
		totalSize += info.Size
	})
	if err != nil {
		return err
	}

	bucket.ObjectCount = objectCount
	bucket.Size = totalSize

	return nil
}

func (r *filesystemRepository) walkObjects(bucketPath, prefix string, fn func(string, storage.ObjectInfo)) error {
	return filepath.Walk(bucketPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and metadata files
		if info.IsDir() || strings.Contains(path, ".metadata") {
			return nil
		}

		// Calculate relative key
		relPath, err := filepath.Rel(bucketPath, path)
		if err != nil {
			return err
		}

		// Convert filepath separators to forward slashes for S3 compatibility
		key := filepath.ToSlash(relPath)

		// Apply prefix filter
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			return nil
		}

		// Try to get metadata
		objectInfo, err := r.GetObjectInfo(context.Background(), filepath.Base(bucketPath), key)
		if err != nil {
			// If metadata doesn't exist, create basic info
			objectInfo = &storage.ObjectInfo{
				Key:          key,
				Size:         info.Size(),
				ContentType:  "application/octet-stream",
				ETag:         "\"unknown\"",
				LastModified: info.ModTime().UTC(),
				Metadata:     make(map[string]string),
			}
		}

		fn(key, *objectInfo)
		return nil
	})
}
