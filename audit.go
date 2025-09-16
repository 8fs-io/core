package eightfs

import (
	"encoding/json"
	"time"

	appLogger "github.com/8fs-io/core/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditEvent represents a security/compliance audit event
type AuditEvent struct {
	Timestamp    time.Time   `json:"timestamp"`
	RequestID    string      `json:"request_id"`
	EventType    string      `json:"event_type"` // "auth", "s3_operation", "error"
	Action       string      `json:"action"`     // "login", "PUT", "GET", etc.
	Resource     string      `json:"resource"`   // bucket/object path
	UserID       string      `json:"user_id"`    // access key or user identifier
	SourceIP     string      `json:"source_ip"`
	UserAgent    string      `json:"user_agent"`
	Success      bool        `json:"success"`
	ErrorMessage string      `json:"error_message,omitempty"`
	ResponseCode int         `json:"response_code"`
	Duration     int64       `json:"duration_ms"` // request duration in milliseconds
	Metadata     interface{} `json:"metadata,omitempty"`
}

// AuditContext holds request-specific audit information
type AuditContext struct {
	RequestID string
	SourceIP  string
	UserAgent string
	StartTime time.Time
}

// AuditLogger handles structured audit logging
type AuditLogger struct {
	// Could be extended with file output, syslog, etc.
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

// LogEvent logs an audit event as structured JSON
func (a *AuditLogger) LogEvent(event *AuditEvent) {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Generate request ID if not provided
	if event.RequestID == "" {
		event.RequestID = uuid.New().String()
	}

	// Marshal to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		appLogger.Default().Error("audit marshal failed", "error", err)
		return
	}
	appLogger.Default().Info("audit event", "event", string(eventJSON))
}

// Middleware to set audit context for requests
func auditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create audit context
		auditCtx := &AuditContext{
			RequestID: uuid.New().String(),
			SourceIP:  c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			StartTime: time.Now(),
		}

		// Store in gin context
		c.Set("audit_context", auditCtx)

		// Process request
		c.Next()
	}
}

// Helper function to extract audit context from request
func getAuditContext(ctx *gin.Context) *AuditContext {
	auditCtx, exists := ctx.Get("audit_context")
	if !exists {
		return &AuditContext{
			RequestID: uuid.New().String(),
			SourceIP:  ctx.ClientIP(),
			UserAgent: ctx.GetHeader("User-Agent"),
			StartTime: time.Now(),
		}
	}
	if ac, ok := auditCtx.(*AuditContext); ok {
		return ac
	}
	return &AuditContext{
		RequestID: uuid.New().String(),
		SourceIP:  ctx.ClientIP(),
		UserAgent: ctx.GetHeader("User-Agent"),
		StartTime: time.Now(),
	}
}

// Global audit logger instance
var GlobalAuditLogger = NewAuditLogger()

// LogS3Operation logs S3 API operations
func LogS3Operation(c *gin.Context, action, resource string, success bool, responseCode int, errorMsg string, metadata interface{}) {
	auditCtx := getAuditContext(c)
	duration := time.Since(auditCtx.StartTime).Milliseconds()

	event := &AuditEvent{
		Timestamp:    time.Now().UTC(),
		RequestID:    auditCtx.RequestID,
		EventType:    "s3_operation",
		Action:       action,
		Resource:     resource,
		SourceIP:     auditCtx.SourceIP,
		UserAgent:    auditCtx.UserAgent,
		Success:      success,
		ErrorMessage: errorMsg,
		ResponseCode: responseCode,
		Duration:     duration,
		Metadata:     metadata,
	}

	// Get user from auth context if available
	if userID, exists := c.Get("access_key"); exists {
		if uid, ok := userID.(string); ok {
			event.UserID = uid
		}
	}

	GlobalAuditLogger.LogEvent(event)
}

// LogAuthEvent logs authentication events
func LogAuthEvent(c *gin.Context, accessKey string, success bool, errorMsg string) {
	auditCtx := getAuditContext(c)
	duration := time.Since(auditCtx.StartTime).Milliseconds()

	event := &AuditEvent{
		Timestamp:    time.Now().UTC(),
		RequestID:    auditCtx.RequestID,
		EventType:    "auth",
		Action:       "authenticate",
		Resource:     c.Request.URL.Path,
		UserID:       accessKey,
		SourceIP:     auditCtx.SourceIP,
		UserAgent:    auditCtx.UserAgent,
		Success:      success,
		ErrorMessage: errorMsg,
		ResponseCode: func() int {
			if success {
				return 200
			}
			return 403
		}(),
		Duration: duration,
		Metadata: map[string]interface{}{
			"method": c.Request.Method,
		},
	}

	GlobalAuditLogger.LogEvent(event)
}
