package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

var Logger zerolog.Logger

// Init initializes the global logger
func Init(serviceName string, isDevelopment bool) {
	// Set global log level
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var output io.Writer = os.Stdout

	if isDevelopment {
		// Pretty print for development
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
	}

	// Create logger
	Logger = zerolog.New(output).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	// Set as global logger
	log.Logger = Logger
}

// WithContext returns a logger with trace information from context
func WithContext(ctx context.Context) *zerolog.Logger {
	logger := Logger.With().Logger()

	// Add trace ID if available
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		logger = logger.With().
			Str("trace_id", span.SpanContext().TraceID().String()).
			Str("span_id", span.SpanContext().SpanID().String()).
			Logger()
	}

	return &logger
}

// Info logs at info level with context
func Info(ctx context.Context) *zerolog.Event {
	return WithContext(ctx).Info()
}

// Error logs at error level with context
func Error(ctx context.Context) *zerolog.Event {
	return WithContext(ctx).Error()
}

// Debug logs at debug level with context
func Debug(ctx context.Context) *zerolog.Event {
	return WithContext(ctx).Debug()
}

// Warn logs at warn level with context
func Warn(ctx context.Context) *zerolog.Event {
	return WithContext(ctx).Warn()
}

// Fatal logs at fatal level with context
func Fatal(ctx context.Context) *zerolog.Event {
	return WithContext(ctx).Fatal()
}

// SetLevel sets the global log level
func SetLevel(level string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

