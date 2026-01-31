#!/bin/bash
set -e

# Determine GOBIN path
GOBIN="${GOBIN:-${GOPATH:-$HOME/go}/bin}"

# Check if air is installed
if ! command -v air &> /dev/null && [ ! -f "$GOBIN/air" ]; then
    echo "Installing air..."
    go install github.com/air-verse/air@latest
fi

# Run with air for hot reload
echo "Starting development server with hot reload..."
"$GOBIN/air" -c .air.toml
