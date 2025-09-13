package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

// Logger interface defines logging methods
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	With(args ...interface{}) Logger
	WithContext(ctx context.Context) Logger
}

// logger implements Logger interface using slog
type logger struct {
	slog *slog.Logger
}

// Config for logger initialization
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	Output     string // stdout, file
	FilePath   string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

// contextKey is used for context keys
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	TraceIDKey   contextKey = "trace_id"
)

var defaultLogger Logger

// New creates a new logger instance
func New(cfg Config) (Logger, error) {
	level := parseLevel(cfg.Level)

	var writer io.Writer
	switch cfg.Output {
	case "file":
		// For simplicity, use os.OpenFile here
		// In production, you might want to use lumberjack for log rotation
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writer = file
	case "stdout":
		fallthrough
	default:
		writer = os.Stdout
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	default:
		handler = slog.NewJSONHandler(writer, opts)
	}

	return &logger{
		slog: slog.New(handler),
	}, nil
}

func (l *logger) Debug(msg string, args ...interface{}) {
	l.slog.Debug(msg, args...)
}

func (l *logger) Info(msg string, args ...interface{}) {
	l.slog.Info(msg, args...)
}

func (l *logger) Warn(msg string, args ...interface{}) {
	l.slog.Warn(msg, args...)
}

func (l *logger) Error(msg string, args ...interface{}) {
	l.slog.Error(msg, args...)
}

func (l *logger) With(args ...interface{}) Logger {
	return &logger{
		slog: l.slog.With(args...),
	}
}

func (l *logger) WithContext(ctx context.Context) Logger {
	newLogger := l.slog

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		newLogger = newLogger.With("request_id", requestID)
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		newLogger = newLogger.With("user_id", userID)
	}

	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		newLogger = newLogger.With("trace_id", traceID)
	}

	return &logger{slog: newLogger}
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Setup initializes the default logger
func Setup(cfg Config) error {
	logger, err := New(cfg)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// Default returns the default logger instance
func Default() Logger {
	if defaultLogger == nil {
		// Fallback to a basic logger if not initialized
		defaultLogger = &logger{
			slog: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		}
	}
	return defaultLogger
}

// Context helpers
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func GenerateRequestID() string {
	return uuid.New().String()
}

// Audit logger for structured audit events
type AuditLogger struct {
	logger Logger
}

type AuditEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	RequestID    string                 `json:"request_id"`
	EventType    string                 `json:"event_type"`
	Action       string                 `json:"action"`
	Resource     string                 `json:"resource"`
	UserID       string                 `json:"user_id"`
	SourceIP     string                 `json:"source_ip"`
	UserAgent    string                 `json:"user_agent"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	ResponseCode int                    `json:"response_code"`
	DurationMS   int64                  `json:"duration_ms"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

func NewAuditLogger(logger Logger) *AuditLogger {
	return &AuditLogger{logger: logger}
}

func (al *AuditLogger) Log(event AuditEvent) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		al.logger.Error("Failed to marshal audit event", "error", err)
		return
	}

	al.logger.Info("AUDIT", "event", string(eventJSON))
}

// Global audit logger instance
var globalAuditLogger *AuditLogger

func SetupAudit(logger Logger) {
	globalAuditLogger = NewAuditLogger(logger)
}

func Audit() *AuditLogger {
	if globalAuditLogger == nil {
		globalAuditLogger = NewAuditLogger(Default())
	}
	return globalAuditLogger
}
