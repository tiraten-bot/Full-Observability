package grpc

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/tair/full-observability/pkg/auth"
)

var grpcTracer = otel.Tracer("grpc-server")

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

// LoggingInterceptor logs gRPC requests
func LoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// Call the handler
	resp, err := handler(ctx, req)

	// Log the request
	duration := time.Since(start)
	
	// Extract trace ID from context if available
	traceID := "no-trace"
	if span := oteltrace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		traceID = span.SpanContext().TraceID().String()
	}
	
	if err != nil {
		log.Printf("gRPC [%s] trace_id=%s duration=%v error=%v", info.FullMethod, traceID, duration, err)
	} else {
		log.Printf("gRPC [%s] trace_id=%s duration=%v", info.FullMethod, traceID, duration)
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

