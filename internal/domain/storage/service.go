package storage

import (
	"context"
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/8fs/8fs/pkg/errors"
	"github.com/8fs/8fs/pkg/logger"
)

// service implements the Service interface
type service struct {
	repo      Repository
	validator Validator
	logger    logger.Logger
}

// NewService creates a new storage service
func NewService(repo Repository, validator Validator, logger logger.Logger) Service {
	return &service{
		repo:      repo,
		validator: validator,
		logger:    logger,
	}
}

// CreateBucket creates a new bucket
func (s *service) CreateBucket(ctx context.Context, name string, metadata map[string]string) (*Bucket, error) {
	// Validate bucket name
	if err := s.validator.ValidateBucketName(name); err != nil {
		return nil, err
	}

	// Validate metadata
	if err := s.validator.ValidateMetadata(metadata); err != nil {
		return nil, err
	}

	// Check if bucket already exists
	exists, err := s.repo.BucketExists(ctx, name)
	if err != nil {
		s.logger.Error("Failed to check bucket existence", "bucket", name, "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to check bucket existence", err)
	}
	if exists {
		return nil, errors.ErrBucketExists.WithContext("bucket", name)
	}

	// Create bucket
	bucket := &Bucket{
		Name:        name,
		CreatedAt:   time.Now().UTC(),
		Metadata:    metadata,
		ObjectCount: 0,
		Size:        0,
	}

	if err := s.repo.CreateBucket(ctx, bucket); err != nil {
		s.logger.Error("Failed to create bucket", "bucket", name, "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to create bucket", err)
	}

	s.logger.Info("Bucket created successfully", "bucket", name)
	return bucket, nil
}

// DeleteBucket deletes a bucket
func (s *service) DeleteBucket(ctx context.Context, name string) error {
	// Validate bucket name
	if err := s.validator.ValidateBucketName(name); err != nil {
		return err
	}

	// Check if bucket exists
	exists, err := s.repo.BucketExists(ctx, name)
	if err != nil {
		s.logger.Error("Failed to check bucket existence", "bucket", name, "error", err)
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to check bucket existence", err)
	}
	if !exists {
		return errors.ErrBucketNotFound.WithContext("bucket", name)
	}

	// Check if bucket is empty
	listResult, err := s.repo.ListObjects(ctx, name, ListOptions{MaxKeys: 1})
	if err != nil {
		s.logger.Error("Failed to list objects", "bucket", name, "error", err)
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to check if bucket is empty", err)
	}
	if len(listResult.Objects) > 0 {
		return errors.ErrBucketNotEmpty.WithContext("bucket", name)
	}

	// Delete bucket
	if err := s.repo.DeleteBucket(ctx, name); err != nil {
		s.logger.Error("Failed to delete bucket", "bucket", name, "error", err)
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to delete bucket", err)
	}

	s.logger.Info("Bucket deleted successfully", "bucket", name)
	return nil
}

// GetBucket retrieves bucket information
func (s *service) GetBucket(ctx context.Context, name string) (*Bucket, error) {
	if err := s.validator.ValidateBucketName(name); err != nil {
		return nil, err
	}

	bucket, err := s.repo.GetBucket(ctx, name)
	if err != nil {
		if errors.IsErrorCode(err, errors.ErrCodeBucketNotFound) {
			return nil, err
		}
		s.logger.Error("Failed to get bucket", "bucket", name, "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to get bucket", err)
	}

	return bucket, nil
}

// ListBuckets lists all buckets
func (s *service) ListBuckets(ctx context.Context) ([]*Bucket, error) {
	buckets, err := s.repo.ListBuckets(ctx)
	if err != nil {
		s.logger.Error("Failed to list buckets", "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to list buckets", err)
	}

	return buckets, nil
}

// PutObject stores an object
func (s *service) PutObject(ctx context.Context, bucket, key string, data []byte, contentType string, metadata map[string]string) (*Object, error) {
	// Validate inputs
	if err := s.validator.ValidateBucketName(bucket); err != nil {
		return nil, err
	}
	if err := s.validator.ValidateObjectKey(key); err != nil {
		return nil, err
	}
	if err := s.validator.ValidateMetadata(metadata); err != nil {
		return nil, err
	}

	// Check if bucket exists
	exists, err := s.repo.BucketExists(ctx, bucket)
	if err != nil {
		s.logger.Error("Failed to check bucket existence", "bucket", bucket, "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to check bucket existence", err)
	}
	if !exists {
		return nil, errors.ErrBucketNotFound.WithContext("bucket", bucket)
	}

	// Generate ETag (MD5 hash of content)
	etag := fmt.Sprintf("\"%x\"", md5.Sum(data))

	// Create object
	object := &Object{
		Key:          key,
		Bucket:       bucket,
		Size:         int64(len(data)),
		ContentType:  contentType,
		ETag:         etag,
		LastModified: time.Now().UTC(),
		Metadata:     metadata,
		Data:         data,
	}

	if err := s.repo.PutObject(ctx, object); err != nil {
		s.logger.Error("Failed to put object", "bucket", bucket, "key", key, "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to put object", err)
	}

	s.logger.Info("Object stored successfully", "bucket", bucket, "key", key, "size", len(data))
	return object, nil
}

// GetObject retrieves an object
func (s *service) GetObject(ctx context.Context, bucket, key string) (*Object, error) {
	if err := s.validator.ValidateBucketName(bucket); err != nil {
		return nil, err
	}
	if err := s.validator.ValidateObjectKey(key); err != nil {
		return nil, err
	}

	object, err := s.repo.GetObject(ctx, bucket, key)
	if err != nil {
		if errors.IsErrorCode(err, errors.ErrCodeObjectNotFound) {
			return nil, err
		}
		s.logger.Error("Failed to get object", "bucket", bucket, "key", key, "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to get object", err)
	}

	return object, nil
}

// GetObjectInfo retrieves object metadata
func (s *service) GetObjectInfo(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	if err := s.validator.ValidateBucketName(bucket); err != nil {
		return nil, err
	}
	if err := s.validator.ValidateObjectKey(key); err != nil {
		return nil, err
	}

	objectInfo, err := s.repo.GetObjectInfo(ctx, bucket, key)
	if err != nil {
		if errors.IsErrorCode(err, errors.ErrCodeObjectNotFound) {
			return nil, err
		}
		s.logger.Error("Failed to get object info", "bucket", bucket, "key", key, "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to get object info", err)
	}

	return objectInfo, nil
}

// DeleteObject deletes an object
func (s *service) DeleteObject(ctx context.Context, bucket, key string) error {
	if err := s.validator.ValidateBucketName(bucket); err != nil {
		return err
	}
	if err := s.validator.ValidateObjectKey(key); err != nil {
		return err
	}

	// Check if object exists
	exists, err := s.repo.ObjectExists(ctx, bucket, key)
	if err != nil {
		s.logger.Error("Failed to check object existence", "bucket", bucket, "key", key, "error", err)
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to check object existence", err)
	}
	if !exists {
		return errors.ErrObjectNotFound.WithContext("bucket", bucket).WithContext("key", key)
	}

	if err := s.repo.DeleteObject(ctx, bucket, key); err != nil {
		s.logger.Error("Failed to delete object", "bucket", bucket, "key", key, "error", err)
		return errors.Wrap(errors.ErrCodeInternalError, "Failed to delete object", err)
	}

	s.logger.Info("Object deleted successfully", "bucket", bucket, "key", key)
	return nil
}

// ListObjects lists objects in a bucket
func (s *service) ListObjects(ctx context.Context, bucket string, opts ListOptions) (*ListResult, error) {
	if err := s.validator.ValidateBucketName(bucket); err != nil {
		return nil, err
	}

	// Check if bucket exists
	exists, err := s.repo.BucketExists(ctx, bucket)
	if err != nil {
		s.logger.Error("Failed to check bucket existence", "bucket", bucket, "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to check bucket existence", err)
	}
	if !exists {
		return nil, errors.ErrBucketNotFound.WithContext("bucket", bucket)
	}

	result, err := s.repo.ListObjects(ctx, bucket, opts)
	if err != nil {
		s.logger.Error("Failed to list objects", "bucket", bucket, "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to list objects", err)
	}

	return result, nil
}

// GetStorageStats retrieves storage statistics
func (s *service) GetStorageStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := s.repo.GetStorageStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get storage stats", "error", err)
		return nil, errors.Wrap(errors.ErrCodeInternalError, "Failed to get storage stats", err)
	}

	return stats, nil
}

// HealthCheck performs a health check on the storage system
func (s *service) HealthCheck(ctx context.Context) error {
	if err := s.repo.HealthCheck(ctx); err != nil {
		s.logger.Error("Storage health check failed", "error", err)
		return errors.Wrap(errors.ErrCodeServiceUnavailable, "Storage service unavailable", err)
	}

	return nil
}

// validator implements the Validator interface
type validator struct{}

// NewValidator creates a new validator
func NewValidator() Validator {
	return &validator{}
}

// ValidateBucketName validates bucket name according to S3 rules
func (v *validator) ValidateBucketName(name string) error {
	if name == "" {
		return errors.ErrInvalidBucketName.WithContext("reason", "bucket name cannot be empty")
	}

	if len(name) < 3 || len(name) > 63 {
		return errors.ErrInvalidBucketName.WithContext("reason", "bucket name must be between 3 and 63 characters")
	}

	// Basic S3 bucket naming rules
	validName := regexp.MustCompile(`^[a-z0-9][a-z0-9\-]*[a-z0-9]$`)
	if !validName.MatchString(name) {
		return errors.ErrInvalidBucketName.WithContext("reason", "bucket name contains invalid characters")
	}

	// Cannot start or end with hyphen
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return errors.ErrInvalidBucketName.WithContext("reason", "bucket name cannot start or end with hyphen")
	}

	// Cannot contain consecutive hyphens
	if strings.Contains(name, "--") {
		return errors.ErrInvalidBucketName.WithContext("reason", "bucket name cannot contain consecutive hyphens")
	}

	return nil
}

// ValidateObjectKey validates object key
func (v *validator) ValidateObjectKey(key string) error {
	if key == "" {
		return errors.ErrInvalidObjectName.WithContext("reason", "object key cannot be empty")
	}

	if len(key) > 1024 {
		return errors.ErrInvalidObjectName.WithContext("reason", "object key too long (max 1024 characters)")
	}

	// Check for invalid characters (very basic validation)
	if strings.Contains(key, "\x00") {
		return errors.ErrInvalidObjectName.WithContext("reason", "object key contains null character")
	}

	return nil
}

// ValidateMetadata validates object metadata
func (v *validator) ValidateMetadata(metadata map[string]string) error {
	if metadata == nil {
		return nil
	}

	if len(metadata) > 10 {
		return errors.New(errors.ErrCodeInvalidParameter, "Too many metadata entries (max 10)")
	}

	for key, value := range metadata {
		if len(key) > 128 {
			return errors.New(errors.ErrCodeInvalidParameter, "Metadata key too long (max 128 characters)")
		}
		if len(value) > 256 {
			return errors.New(errors.ErrCodeInvalidParameter, "Metadata value too long (max 256 characters)")
		}
	}

	return nil
}
