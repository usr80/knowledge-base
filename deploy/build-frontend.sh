#!/bin/bash
# Build frontend only (Vite/Vuetify)
# Output: build/frontend/dist/

set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
FRONTEND_BUILD="$BUILD_DIR/frontend"

echo "=========================================="
echo "  Build Frontend (Vite)"
echo "=========================================="

# Create build directory
mkdir -p "$FRONTEND_BUILD"

# Build
cd "$PROJECT_ROOT/frontend"

echo "Installing dependencies..."
npm install --silent

echo "Building..."
npm run build

# Show result
if [ -d "$FRONTEND_BUILD/dist" ]; then
    TOTAL_SIZE=$(du -sh "$FRONTEND_BUILD/dist" | cut -f1)
    FILE_COUNT=$(find "$FRONTEND_BUILD/dist" -type f | wc -l)
    echo ""
    echo "Done: build/frontend/dist/ ($TOTAL_SIZE, $FILE_COUNT files)"
else
    echo "Build failed!"
    exit 1
fi