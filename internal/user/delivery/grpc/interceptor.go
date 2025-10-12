package grpc

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/tair/full-observability/pkg/auth"
	"github.com/tair/full-observability/pkg/logger"
)

var grpcTracer = otel.Tracer("grpc-server")

// gRPC Prometheus metrics
var (
	grpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status_code"},
	)

	grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_service_grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	grpcRequestSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "user_service_grpc_request_duration_summary",
			Help: "Summary of gRPC request durations with percentiles",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.95: 0.01,
				0.99: 0.001,
			},
			MaxAge: 10 * time.Minute,
		},
		[]string{"method"},
	)

	grpcErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_grpc_errors_total",
			Help: "Total number of gRPC errors",
		},
		[]string{"method", "error_code"},
	)
)

func init() {
	// Register gRPC metrics
	prometheus.MustRegister(grpcRequestsTotal)
	prometheus.MustRegister(grpcRequestDuration)
	prometheus.MustRegister(grpcRequestSummary)
	prometheus.MustRegister(grpcErrorsTotal)
}

// TracingInterceptor adds distributed tracing to gRPC calls
func TracingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Start a new span
	ctx, span := grpcTracer.Start(ctx, info.FullMethod,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", info.FullMethod),
		),
	)
	defer span.End()

	// Call the handler
	resp, err := handler(ctx, req)

	// Record error if any
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		
		// Add gRPC status code
		if st, ok := status.FromError(err); ok {
			span.SetAttributes(attribute.String("rpc.grpc.status_code", st.Code().String()))
		}
	} else {
		span.SetStatus(codes.Ok, "success")
	}

	return resp, err
}

// MetricsInterceptor collects Prometheus metrics for gRPC calls
func MetricsInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// Call the handler
	resp, err := handler(ctx, req)

	// Calculate duration
	duration := time.Since(start).Seconds()

	// Get gRPC status code
	statusCode := "OK"
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code().String()
			// Track errors separately
			grpcErrorsTotal.WithLabelValues(info.FullMethod, statusCode).Inc()
		} else {
			statusCode = "Unknown"
		}
	}

	// Record metrics
	grpcRequestsTotal.WithLabelValues(info.FullMethod, statusCode).Inc()
	grpcRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)
	grpcRequestSummary.WithLabelValues(info.FullMethod).Observe(duration)

	return resp, err
}

// LoggingInterceptor logs gRPC requests with structured logging
func LoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// Extract trace ID
	traceID := "no-trace"
	if span := oteltrace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		traceID = span.SpanContext().TraceID().String()
	}

	// Log request start
	logger.Info(ctx).
		Str("method", info.FullMethod).
		Str("protocol", "grpc").
		Str("trace_id", traceID).
		Msg("gRPC request started")

	// Call the handler
	resp, err := handler(ctx, req)

	// Calculate duration
	duration := time.Since(start)

	// Log request completion
	if err != nil {
		// Get gRPC status code
		grpcStatus := "unknown"
		if st, ok := status.FromError(err); ok {
			grpcStatus = st.Code().String()
		}

		logger.Error(ctx).
			Str("method", info.FullMethod).
			Str("protocol", "grpc").
			Dur("duration", duration).
			Int64("duration_ms", duration.Milliseconds()).
			Str("trace_id", traceID).
			Str("grpc_status", grpcStatus).
			Err(err).
			Msg("gRPC request failed")
	} else {
		logger.Info(ctx).
			Str("method", info.FullMethod).
			Str("protocol", "grpc").
			Dur("duration", duration).
			Int64("duration_ms", duration.Milliseconds()).
			Str("trace_id", traceID).
			Msg("gRPC request completed")
	}

	return resp, err
}

// AuthInterceptor validates JWT tokens for protected methods
func AuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Skip auth for public methods
	publicMethods := map[string]bool{
		"/user.v1.UserService/Register": true,
		"/user.v1.UserService/Login":    true,
	}

	if publicMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(grpccodes.Unauthenticated, "metadata not provided")
	}

	// Get authorization token
	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Errorf(grpccodes.Unauthenticated, "authorization token not provided")
	}

	// Validate token
	token := tokens[0]
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		return nil, status.Errorf(grpccodes.Unauthenticated, "invalid token: %v", err)
	}

	// Add claims to context
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "username", claims.Username)
	ctx = context.WithValue(ctx, "role", claims.Role)

	// Check admin methods
	adminMethods := map[string]bool{
		"/user.v1.UserService/ChangeRole":   true,
		"/user.v1.UserService/ToggleActive": true,
		"/user.v1.UserService/GetStats":     true,
	}

	if adminMethods[info.FullMethod] && claims.Role != "admin" {
		return nil, status.Errorf(grpccodes.PermissionDenied, "admin access required")
	}

	return handler(ctx, req)
}

