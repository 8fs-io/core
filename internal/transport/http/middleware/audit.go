package middleware

import (
	"time"

	"github.com/8fs/8fs/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Audit creates a middleware that logs audit events
func Audit(auditLogger *logger.AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Create audit event
		duration := time.Since(start)

		event := logger.AuditEvent{
			Timestamp:    start.UTC(),
			RequestID:    getRequestID(c),
			EventType:    "http_request",
			Action:       c.Request.Method,
			Resource:     c.Request.URL.Path,
			UserID:       getUserID(c),
			SourceIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Success:      c.Writer.Status() < 400,
			ResponseCode: c.Writer.Status(),
			DurationMS:   duration.Milliseconds(),
			Metadata: map[string]interface{}{
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"query":     c.Request.URL.RawQuery,
				"body_size": c.Writer.Size(),
			},
		}

		// Set error message if request failed
		if !event.Success {
			if errors := c.Errors; len(errors) > 0 {
				event.ErrorMessage = errors[0].Error()
			} else {
				event.ErrorMessage = "HTTP error"
			}
		}

		auditLogger.Log(event)
	}
}

// Helper functions
func getRequestID(c *gin.Context) string {
	return c.GetHeader(RequestIDHeader)
}

func getUserID(c *gin.Context) string {
	// Try to get user ID from context or headers
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	return ""
}
