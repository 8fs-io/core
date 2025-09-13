package storage

import (
	"context"
	"time"
)

// Bucket represents a storage bucket
type Bucket struct {
	Name        string            `json:"name"`
	CreatedAt   time.Time         `json:"created_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ObjectCount int64             `json:"object_count"`
	Size        int64             `json:"size"` // total size in bytes
}

// Object represents a storage object
type Object struct {
	Key          string            `json:"key"`
	Bucket       string            `json:"bucket"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	ETag         string            `json:"etag"`
	LastModified time.Time         `json:"last_modified"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Data         []byte            `json:"-"` // Actual object data
}

// ObjectInfo represents object metadata without data
type ObjectInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	ETag         string            `json:"etag"`
	LastModified time.Time         `json:"last_modified"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ListOptions represents options for listing operations
type ListOptions struct {
	Prefix    string `json:"prefix,omitempty"`
	Delimiter string `json:"delimiter,omitempty"`
	MaxKeys   int    `json:"max_keys,omitempty"`
	Marker    string `json:"marker,omitempty"`
}

// ListResult represents the result of a list operation
type ListResult struct {
	Objects        []ObjectInfo `json:"objects"`
	CommonPrefixes []string     `json:"common_prefixes,omitempty"`
	IsTruncated    bool         `json:"is_truncated"`
	NextMarker     string       `json:"next_marker,omitempty"`
}

// Repository defines the storage repository interface
type Repository interface {
	// Bucket operations
	CreateBucket(ctx context.Context, bucket *Bucket) error
	DeleteBucket(ctx context.Context, name string) error
	GetBucket(ctx context.Context, name string) (*Bucket, error)
	ListBuckets(ctx context.Context) ([]*Bucket, error)
	BucketExists(ctx context.Context, name string) (bool, error)

	// Object operations
	PutObject(ctx context.Context, object *Object) error
	GetObject(ctx context.Context, bucket, key string) (*Object, error)
	GetObjectInfo(ctx context.Context, bucket, key string) (*ObjectInfo, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, bucket string, opts ListOptions) (*ListResult, error)
	ObjectExists(ctx context.Context, bucket, key string) (bool, error)

	// Utility operations
	GetStorageStats(ctx context.Context) (map[string]interface{}, error)
	HealthCheck(ctx context.Context) error
}

// Service defines the storage service interface
type Service interface {
	// Bucket operations
	CreateBucket(ctx context.Context, name string, metadata map[string]string) (*Bucket, error)
	DeleteBucket(ctx context.Context, name string) error
	GetBucket(ctx context.Context, name string) (*Bucket, error)
	ListBuckets(ctx context.Context) ([]*Bucket, error)

	// Object operations
	PutObject(ctx context.Context, bucket, key string, data []byte, contentType string, metadata map[string]string) (*Object, error)
	GetObject(ctx context.Context, bucket, key string) (*Object, error)
	GetObjectInfo(ctx context.Context, bucket, key string) (*ObjectInfo, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, bucket string, opts ListOptions) (*ListResult, error)

	// Utility operations
	GetStorageStats(ctx context.Context) (map[string]interface{}, error)
	HealthCheck(ctx context.Context) error
}

// Validator defines validation interface
type Validator interface {
	ValidateBucketName(name string) error
	ValidateObjectKey(key string) error
	ValidateMetadata(metadata map[string]string) error
}
