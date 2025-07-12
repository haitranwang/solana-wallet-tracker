#!/bin/bash

set -e

echo "Installing Solana Wallet Token Tracker..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.18 or newer."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)

if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 18 ]); then
    echo "Go version 1.18 or newer is required. Current version: $GO_VERSION"
    exit 1
fi

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Build the application
echo "Building application..."
go build -o tracker ./cmd/tracker

# Create default configuration if not exists
if [ ! -f config.json ]; then
    echo "Creating default configuration file..."
    cp config.example.json config.json
    echo "Please edit config.json with your Solana wallet addresses and token mint addresses."
fi

echo "Installation complete! Run ./tracker to start the application."