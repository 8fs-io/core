package handlers

import (
	"strings"

	"github.com/8fs-io/core/internal/container"
	"github.com/8fs-io/core/pkg/errors"
	"github.com/8fs-io/core/pkg/logger"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication middleware
type AuthHandler struct {
	container *container.Container
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(c *container.Container) *AuthHandler {
	return &AuthHandler{container: c}
}

// AWSSignatureMiddleware validates AWS Signature v4 authentication
func (h *AuthHandler) AWSSignatureMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health and metrics endpoints
		path := c.Request.URL.Path
		if path == "/healthz" || path == h.container.Config.Metrics.Path {
			c.Next()
			return
		}

		// Skip auth if disabled
		if !h.container.Config.Auth.Enabled {
			c.Next()
			return
		}

		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			h.respondAuthError(c, errors.ErrAuthRequired, "Missing Authorization header")
			return
		}

		// Parse AWS Signature v4 header
		if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256") {
			h.respondAuthError(c, errors.ErrInvalidSignature, "Invalid signature format")
			return
		}

		// Extract credential from authorization header
		credential := h.extractCredential(authHeader)
		if credential == "" {
			h.respondAuthError(c, errors.ErrInvalidCredentials, "Missing credential")
			return
		}

		// Simple credential validation (in production, this should be more sophisticated)
		if !h.validateCredential(credential) {
			h.respondAuthError(c, errors.ErrInvalidCredentials, "Invalid access key")
			authRequestsTotal.WithLabelValues(credential, "failure").Inc()
			return
		}

		authRequestsTotal.WithLabelValues(credential, "success").Inc()

		// Set user context
		ctx := logger.WithUserID(c.Request.Context(), credential)
		c.Request = c.Request.WithContext(ctx)
		c.Set("user_id", credential)

		c.Next()
	}
}

// extractCredential extracts the access key from the Authorization header
func (h *AuthHandler) extractCredential(authHeader string) string {
	// Parse: AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host, Signature=...
	parts := strings.Split(authHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "Credential=") {
			credential := strings.TrimPrefix(part, "Credential=")
			// Extract access key (before first slash)
			if idx := strings.Index(credential, "/"); idx > 0 {
				return credential[:idx]
			}
			return credential
		}
	}

	// Also try parsing the first part after AWS4-HMAC-SHA256
	if strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256 Credential=") {
		remaining := strings.TrimPrefix(authHeader, "AWS4-HMAC-SHA256 Credential=")
		if commaIdx := strings.Index(remaining, ","); commaIdx > 0 {
			credential := remaining[:commaIdx]
			if slashIdx := strings.Index(credential, "/"); slashIdx > 0 {
				return credential[:slashIdx]
			}
			return credential
		}
	}

	return ""
}

// validateCredential validates the access key (simplified)
func (h *AuthHandler) validateCredential(accessKey string) bool {
	// In a real implementation, this would:
	// 1. Look up the access key in a database
	// 2. Validate the complete signature
	// 3. Check expiration and permissions

	// For now, just check against the default key from config
	return accessKey == h.container.Config.Auth.DefaultKey.AccessKey
}

// respondAuthError sends an authentication error response
func (h *AuthHandler) respondAuthError(c *gin.Context, err *errors.AppError, message string) {
	// Log the auth failure
	h.container.Logger.Warn("Authentication failed",
		"error", message,
		"ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
		"path", c.Request.URL.Path,
	)

	// For S3 compatibility, return XML error for S3 endpoints
	if h.isS3Request(c) {
		errorResponse := struct {
			XMLName   string `xml:"Error"`
			Code      string `xml:"Code"`
			Message   string `xml:"Message"`
			Resource  string `xml:"Resource"`
			RequestID string `xml:"RequestId"`
		}{
			Code:      string(err.Code),
			Message:   message,
			Resource:  c.Request.URL.Path,
			RequestID: c.GetHeader("X-Request-ID"),
		}
		c.XML(err.HTTPStatus, errorResponse)
	} else {
		c.JSON(err.HTTPStatus, gin.H{
			"error": gin.H{
				"code":    err.Code,
				"message": message,
			},
		})
	}

	c.Abort()
}

// isS3Request checks if the request is for S3-compatible API
func (h *AuthHandler) isS3Request(c *gin.Context) bool {
	path := c.Request.URL.Path
	// S3 requests are typically bucket/object paths or root
	return !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/healthz") && path != h.container.Config.Metrics.Path
}
