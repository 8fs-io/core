package middleware

import (
	"github.com/8fs-io/core/pkg/logger"
	"github.com/gin-gonic/gin"
)

const RequestIDHeader = "X-Request-ID"

// RequestID creates a middleware that generates and attaches a request ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = logger.GenerateRequestID()
			c.Header(RequestIDHeader, requestID)
		}

		// Store request ID in context
		ctx := logger.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
