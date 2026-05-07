#!/bin/bash
# Build backend only (Linux amd64)
# Output: build/backend/knowledge-base

set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
BACKEND_BUILD="$BUILD_DIR/backend"

echo "=========================================="
echo "  Build Backend (Linux amd64)"
echo "=========================================="

# Create build directory
mkdir -p "$BACKEND_BUILD"

# Build
cd "$PROJECT_ROOT/backend"
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOPROXY="https://goproxy.cn,direct"

echo "Compiling..."
go build -ldflags="-s -w" -o "$BACKEND_BUILD/knowledge-base" main.go

# Copy config template
cp "$PROJECT_ROOT/backend/.env.example" "$BACKEND_BUILD/.env.example" 2>/dev/null || true

# Show result
if [ -f "$BACKEND_BUILD/knowledge-base" ]; then
    SIZE=$(stat -f%z "$BACKEND_BUILD/knowledge-base" 2>/dev/null || stat -c%s "$BACKEND_BUILD/knowledge-base")
    echo ""
    echo "Done: build/backend/knowledge-base ($SIZE bytes)"
else
    echo "Build failed!"
    exit 1
fi