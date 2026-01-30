#!/bin/bash
# Development script with hot reload using Air

# Check if air is installed
if ! command -v air &> /dev/null; then
    echo "Installing air for hot reload..."
    go install github.com/air-verse/air@latest
fi

# Run with air
echo "Starting development server with hot reload..."
air
