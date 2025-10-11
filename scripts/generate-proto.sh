#!/bin/bash

# Script to generate Go code from proto files

set -e

echo "🔧 Checking for protoc..."
if ! command -v protoc &> /dev/null; then
    echo "❌ protoc not found. Please install Protocol Buffers compiler."
    echo "   Ubuntu/Debian: sudo apt-get install protobuf-compiler"
    echo "   macOS: brew install protobuf"
    exit 1
fi

echo "✅ protoc found: $(protoc --version)"

echo "🔧 Checking for Go plugins..."
if ! command -v protoc-gen-go &> /dev/null; then
    echo "📦 Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "📦 Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

echo "✅ Go plugins installed"

echo "🚀 Generating Go code from proto files..."

# Create output directory if it doesn't exist
mkdir -p api/proto/user

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/user/user.proto

echo "✅ Proto generation complete!"
echo "📁 Generated files:"
ls -la api/proto/user/*.pb.go 2>/dev/null || echo "   (files will appear after running this script)"

