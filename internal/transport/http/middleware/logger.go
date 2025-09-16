package middleware

import (
	"time"

	"github.com/8fs-io/core/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Logger creates a middleware that logs HTTP requests
func Logger(appLogger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		logLevel := "info"
		if statusCode >= 500 {
			logLevel = "error"
		} else if statusCode >= 400 {
			logLevel = "warn"
		}

		loggerWithContext := appLogger.WithContext(c.Request.Context())

		switch logLevel {
		case "error":
			loggerWithContext.Error("HTTP request",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency,
				"ip", clientIP,
				"size", bodySize,
			)
		case "warn":
			loggerWithContext.Warn("HTTP request",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency,
				"ip", clientIP,
				"size", bodySize,
			)
		default:
			loggerWithContext.Info("HTTP request",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency,
				"ip", clientIP,
				"size", bodySize,
			)
		}
	}
}
