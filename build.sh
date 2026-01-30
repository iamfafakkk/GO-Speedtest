#!/bin/bash
# Build script for GO-Speedtest

set -e

APP_NAME="speedtest"
BUILD_DIR="./build"

echo "Building $APP_NAME..."

# Clean build directory
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# Build for current platform
echo "Building for current platform..."
go build -ldflags="-s -w" -o $BUILD_DIR/$APP_NAME main.go

# Cross-compile for common platforms (optional)
if [ "$1" == "--all" ]; then
    echo "Cross-compiling for multiple platforms..."
    
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-linux-amd64 main.go
    GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-linux-arm64 main.go
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-darwin-amd64 main.go
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-darwin-arm64 main.go
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $BUILD_DIR/${APP_NAME}-windows-amd64.exe main.go
    
    echo "Cross-compile complete!"
fi

echo ""
echo "Build complete!"
ls -lh $BUILD_DIR/

echo ""
echo "Run with: $BUILD_DIR/$APP_NAME"
