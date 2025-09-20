package indexing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/8fs-io/core/internal/domain/ai"
	"github.com/8fs-io/core/pkg/logger"
)

// JobStatus represents the status of an indexing job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// IndexingJob represents a document indexing job
type IndexingJob struct {
	ID          string                 `json:"id"`
	ObjectID    string                 `json:"object_id"`
	Text        string                 `json:"text"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      JobStatus              `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Retries     int                    `json:"retries"`
}

// Config holds async indexing configuration
type Config struct {
	Enabled       bool          `yaml:"enabled"`
	Workers       int           `yaml:"workers"`
	QueueSize     int           `yaml:"queue_size"`
	MaxRetries    int           `yaml:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
	CleanupAfter  time.Duration `yaml:"cleanup_after"`
	StatusEnabled bool          `yaml:"status_enabled"`
}

// DefaultConfig returns default async indexing configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		Workers:       3,
		QueueSize:     1000,
		MaxRetries:    3,
		RetryDelay:    5 * time.Second,
		CleanupAfter:  24 * time.Hour,
		StatusEnabled: true,
	}
}

// Service provides async document indexing
type Service interface {
	// SubmitJob submits a document for async indexing
	SubmitJob(ctx context.Context, objectID, text string, metadata map[string]interface{}) (*IndexingJob, error)

	// GetJobStatus returns the status of an indexing job
	GetJobStatus(ctx context.Context, jobID string) (*IndexingJob, error)

	// GetJobsByObjectID returns all jobs for a specific object
	GetJobsByObjectID(ctx context.Context, objectID string) ([]*IndexingJob, error)

	// Start starts the async indexing workers
	Start(ctx context.Context) error

	// Stop stops the async indexing workers
	Stop() error

	// Stats returns indexing statistics
	Stats() *Stats
}

// Stats represents indexing statistics
type Stats struct {
	TotalJobs      int64     `json:"total_jobs"`
	PendingJobs    int64     `json:"pending_jobs"`
	ProcessingJobs int64     `json:"processing_jobs"`
	CompletedJobs  int64     `json:"completed_jobs"`
	FailedJobs     int64     `json:"failed_jobs"`
	QueueLength    int       `json:"queue_length"`
	WorkersActive  int       `json:"workers_active"`
	LastProcessed  time.Time `json:"last_processed"`
}

// service implements Service interface
type service struct {
	config    *Config
	aiService ai.Service
	logger    logger.Logger

	// Job management
	jobs      map[string]*IndexingJob
	jobsMutex sync.RWMutex
	jobQueue  chan *IndexingJob

	// Worker management
	workers []worker
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	// Statistics
	stats      *Stats
	statsMutex sync.RWMutex
}

// worker represents a background indexing worker
type worker struct {
	id      int
	service *service
}

// NewService creates a new async indexing service
func NewService(config *Config, aiService ai.Service, logger logger.Logger) Service {
	if config == nil {
		config = DefaultConfig()
	}

	return &service{
		config:    config,
		aiService: aiService,
		logger:    logger,
		jobs:      make(map[string]*IndexingJob),
		jobQueue:  make(chan *IndexingJob, config.QueueSize),
		stats: &Stats{
			LastProcessed: time.Now(),
		},
	}
}

// SubmitJob submits a document for async indexing
func (s *service) SubmitJob(ctx context.Context, objectID, text string, metadata map[string]interface{}) (*IndexingJob, error) {
	if !s.config.Enabled {
		// If async indexing is disabled, process synchronously
		if err := s.aiService.ProcessAndStoreDocument(ctx, objectID, text, metadata); err != nil {
			return nil, fmt.Errorf("failed to process document synchronously: %w", err)
		}

		// Return a completed job for compatibility
		job := &IndexingJob{
			ID:          s.generateJobID(objectID),
			ObjectID:    objectID,
			Text:        text,
			Metadata:    metadata,
			Status:      JobStatusCompleted,
			CreatedAt:   time.Now(),
			CompletedAt: &[]time.Time{time.Now()}[0],
		}

		return job, nil
	}

	job := &IndexingJob{
		ID:        s.generateJobID(objectID),
		ObjectID:  objectID,
		Text:      text,
		Metadata:  metadata,
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
		Retries:   0,
	}

	s.jobsMutex.Lock()
	s.jobs[job.ID] = job
	s.jobsMutex.Unlock()

	// Try to enqueue the job
	select {
	case s.jobQueue <- job:
		s.updateStats(func(stats *Stats) {
			stats.TotalJobs++
			stats.PendingJobs++
			stats.QueueLength = len(s.jobQueue)
		})
		s.logger.Info("Submitted indexing job", "job_id", job.ID, "object_id", objectID)
		return job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, fmt.Errorf("indexing queue is full")
	}
}

// GetJobStatus returns the status of an indexing job
func (s *service) GetJobStatus(ctx context.Context, jobID string) (*IndexingJob, error) {
	s.jobsMutex.RLock()
	job, exists := s.jobs[jobID]
	s.jobsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	// Return a copy to avoid race conditions
	jobCopy := *job
	return &jobCopy, nil
}

// GetJobsByObjectID returns all jobs for a specific object
func (s *service) GetJobsByObjectID(ctx context.Context, objectID string) ([]*IndexingJob, error) {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()

	var jobs []*IndexingJob
	for _, job := range s.jobs {
		if job.ObjectID == objectID {
			jobCopy := *job
			jobs = append(jobs, &jobCopy)
		}
	}

	return jobs, nil
}

// Start starts the async indexing workers
func (s *service) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Async indexing disabled, skipping worker startup")
		return nil
	}

	s.ctx, s.cancel = context.WithCancel(ctx)

	// Start workers
	for i := 0; i < s.config.Workers; i++ {
		worker := worker{
			id:      i,
			service: s,
		}
		s.workers = append(s.workers, worker)

		s.wg.Add(1)
		go worker.run(s.ctx)
	}

	// Start cleanup routine
	s.wg.Add(1)
	go s.cleanupRoutine(s.ctx)

	s.logger.Info("Started async indexing service", "workers", s.config.Workers, "queue_size", s.config.QueueSize)
	return nil
}

// Stop stops the async indexing workers
func (s *service) Stop() error {
	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
	}

	s.logger.Info("Stopped async indexing service")
	return nil
}

// Stats returns indexing statistics
func (s *service) Stats() *Stats {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	statsCopy := *s.stats
	statsCopy.QueueLength = len(s.jobQueue)

	// Count active workers
	activeWorkers := 0
	for range s.workers {
		// This is a simple approximation; in production you might track worker status
		activeWorkers++
	}
	statsCopy.WorkersActive = activeWorkers

	return &statsCopy
}

// worker.run processes indexing jobs
func (w *worker) run(ctx context.Context) {
	defer w.service.wg.Done()

	w.service.logger.Info("Started indexing worker", "worker_id", w.id)

	for {
		select {
		case <-ctx.Done():
			w.service.logger.Info("Stopping indexing worker", "worker_id", w.id)
			return
		case job := <-w.service.jobQueue:
			w.processJob(ctx, job)
		}
	}
}

// processJob processes a single indexing job
func (w *worker) processJob(ctx context.Context, job *IndexingJob) {
	w.service.logger.Info("Processing indexing job", "worker_id", w.id, "job_id", job.ID, "object_id", job.ObjectID)

	// Update job status to processing
	w.service.jobsMutex.Lock()
	job.Status = JobStatusProcessing
	now := time.Now()
	job.StartedAt = &now
	w.service.jobsMutex.Unlock()

	w.service.updateStats(func(stats *Stats) {
		stats.PendingJobs--
		stats.ProcessingJobs++
	})

	// Process the document
	err := w.service.aiService.ProcessAndStoreDocument(ctx, job.ObjectID, job.Text, job.Metadata)

	w.service.jobsMutex.Lock()
	defer w.service.jobsMutex.Unlock()

	now = time.Now()
	if err != nil {
		job.Error = err.Error()
		job.Retries++

		if job.Retries < w.service.config.MaxRetries {
			// Retry the job
			job.Status = JobStatusPending
			w.service.logger.Warn("Indexing job failed, retrying", "worker_id", w.id, "job_id", job.ID, "retries", job.Retries, "error", err)

			// Re-enqueue with delay
			go func() {
				time.Sleep(w.service.config.RetryDelay)
				select {
				case w.service.jobQueue <- job:
					w.service.updateStats(func(stats *Stats) {
						stats.ProcessingJobs--
						stats.PendingJobs++
					})
				case <-ctx.Done():
				}
			}()
		} else {
			// Max retries reached
			job.Status = JobStatusFailed
			job.CompletedAt = &now
			w.service.logger.Error("Indexing job failed permanently", "worker_id", w.id, "job_id", job.ID, "retries", job.Retries, "error", err)

			w.service.updateStats(func(stats *Stats) {
				stats.ProcessingJobs--
				stats.FailedJobs++
				stats.LastProcessed = now
			})
		}
	} else {
		// Success
		job.Status = JobStatusCompleted
		job.CompletedAt = &now
		w.service.logger.Info("Indexing job completed successfully", "worker_id", w.id, "job_id", job.ID, "object_id", job.ObjectID)

		w.service.updateStats(func(stats *Stats) {
			stats.ProcessingJobs--
			stats.CompletedJobs++
			stats.LastProcessed = now
		})
	}
}

// cleanupRoutine periodically cleans up old completed jobs
func (s *service) cleanupRoutine(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupOldJobs()
		}
	}
}

// cleanupOldJobs removes old completed/failed jobs from memory
func (s *service) cleanupOldJobs() {
	cutoff := time.Now().Add(-s.config.CleanupAfter)

	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	cleaned := 0
	for id, job := range s.jobs {
		if (job.Status == JobStatusCompleted || job.Status == JobStatusFailed) &&
			job.CreatedAt.Before(cutoff) {
			delete(s.jobs, id)
			cleaned++
		}
	}

	if cleaned > 0 {
		s.logger.Info("Cleaned up old indexing jobs", "count", cleaned)
	}
}

// generateJobID generates a unique job ID
func (s *service) generateJobID(objectID string) string {
	return fmt.Sprintf("idx_%d_%s", time.Now().UnixNano(), objectID)
}

// updateStats safely updates statistics
func (s *service) updateStats(updater func(*Stats)) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	updater(s.stats)
}
