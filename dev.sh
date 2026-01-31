#!/bin/bash
# Development script with hot reload using Air

# # Check if air is installed
# if ! command -v air &> /dev/null; then
#     echo "Installing air for hot reload..."
#     go install github.com/air-verse/air@latest
# fi

# # Run with air
# echo "Starting development server with hot reload..."
# air


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
