package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Storage errors
	ErrCodeBucketExists         ErrorCode = "BUCKET_ALREADY_EXISTS"
	ErrCodeBucketNotFound       ErrorCode = "BUCKET_NOT_FOUND"
	ErrCodeBucketNotEmpty       ErrorCode = "BUCKET_NOT_EMPTY"
	ErrCodeObjectNotFound       ErrorCode = "OBJECT_NOT_FOUND"
	ErrCodeInvalidBucketName    ErrorCode = "INVALID_BUCKET_NAME"
	ErrCodeInvalidObjectName    ErrorCode = "INVALID_OBJECT_NAME"
	ErrCodeStorageQuotaExceeded ErrorCode = "STORAGE_QUOTA_EXCEEDED"

	// Authentication errors
	ErrCodeAuthenticationRequired ErrorCode = "AUTHENTICATION_REQUIRED"
	ErrCodeInvalidCredentials     ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeInvalidSignature       ErrorCode = "INVALID_SIGNATURE"
	ErrCodeAccessDenied           ErrorCode = "ACCESS_DENIED"
	ErrCodeTokenExpired           ErrorCode = "TOKEN_EXPIRED"

	// Request errors
	ErrCodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	ErrCodeMalformedXML     ErrorCode = "MALFORMED_XML"
	ErrCodeMissingHeaders   ErrorCode = "MISSING_REQUIRED_HEADERS"
	ErrCodeInvalidParameter ErrorCode = "INVALID_PARAMETER"
	ErrCodeRequestTooLarge  ErrorCode = "REQUEST_TOO_LARGE"

	// System errors
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeNotImplemented     ErrorCode = "NOT_IMPLEMENTED"
	ErrCodeConfigurationError ErrorCode = "CONFIGURATION_ERROR"
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	HTTPStatus int                    `json:"-"`
	Cause      error                  `json:"-"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause sets the underlying cause of the error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		Context:    make(map[string]interface{}),
	}
}

// Newf creates a new AppError with formatted message
func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return New(code, fmt.Sprintf(format, args...))
}

// Wrap wraps an existing error with an AppError
func Wrap(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		Cause:      cause,
		Context:    make(map[string]interface{}),
	}
}

// Wrapf wraps an existing error with a formatted message
func Wrapf(code ErrorCode, cause error, format string, args ...interface{}) *AppError {
	return Wrap(code, fmt.Sprintf(format, args...), cause)
}

// IsErrorCode checks if an error is a specific AppError code
func IsErrorCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// As is a wrapper around errors.As for convenience
func As(err error, target interface{}) bool {
	switch target := target.(type) {
	case **AppError:
		if appErr, ok := err.(*AppError); ok {
			*target = appErr
			return true
		}
	}
	return false
}

// getHTTPStatus maps error codes to HTTP status codes
func getHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrCodeBucketExists:
		return http.StatusConflict
	case ErrCodeBucketNotFound, ErrCodeObjectNotFound:
		return http.StatusNotFound
	case ErrCodeBucketNotEmpty:
		return http.StatusConflict
	case ErrCodeInvalidBucketName, ErrCodeInvalidObjectName, ErrCodeInvalidRequest, ErrCodeMalformedXML, ErrCodeMissingHeaders, ErrCodeInvalidParameter:
		return http.StatusBadRequest
	case ErrCodeAuthenticationRequired:
		return http.StatusUnauthorized
	case ErrCodeInvalidCredentials, ErrCodeInvalidSignature, ErrCodeTokenExpired:
		return http.StatusUnauthorized
	case ErrCodeAccessDenied:
		return http.StatusForbidden
	case ErrCodeRequestTooLarge:
		return http.StatusRequestEntityTooLarge
	case ErrCodeStorageQuotaExceeded:
		return http.StatusInsufficientStorage
	case ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeNotImplemented:
		return http.StatusNotImplemented
	case ErrCodeInternalError, ErrCodeConfigurationError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Predefined common errors for convenience
var (
	ErrBucketExists       = New(ErrCodeBucketExists, "The requested bucket name is not available")
	ErrBucketNotFound     = New(ErrCodeBucketNotFound, "The specified bucket does not exist")
	ErrBucketNotEmpty     = New(ErrCodeBucketNotEmpty, "The bucket you tried to delete is not empty")
	ErrObjectNotFound     = New(ErrCodeObjectNotFound, "The specified key does not exist")
	ErrInvalidBucketName  = New(ErrCodeInvalidBucketName, "The specified bucket is not valid")
	ErrInvalidObjectName  = New(ErrCodeInvalidObjectName, "The specified key is not valid")
	ErrAuthRequired       = New(ErrCodeAuthenticationRequired, "Authentication required")
	ErrInvalidCredentials = New(ErrCodeInvalidCredentials, "The AWS access key ID you provided does not exist in our records")
	ErrInvalidSignature   = New(ErrCodeInvalidSignature, "The request signature we calculated does not match the signature you provided")
	ErrAccessDenied       = New(ErrCodeAccessDenied, "Access denied")
	ErrInternalError      = New(ErrCodeInternalError, "We encountered an internal error. Please try again")
	ErrNotImplemented     = New(ErrCodeNotImplemented, "A header you provided implies functionality that is not implemented")
)
