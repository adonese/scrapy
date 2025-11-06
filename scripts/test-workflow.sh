#!/bin/bash

# Test Temporal workflow setup
# This script tests the workflow by running the worker and client inside Docker containers

set -e

echo "Testing Temporal workflow setup..."

# Check if Temporal is running
if ! docker ps | grep -q cost-of-living-temporal; then
    echo "Error: Temporal container is not running. Run 'make temporal-up' first."
    exit 1
fi

echo "Temporal is running."

# Build Go binaries
echo "Building worker and client..."
go build -o /tmp/cost-worker cmd/worker/main.go
go build -o /tmp/cost-client examples/workflow_client.go

# Run worker in Docker network
echo "Starting worker in background..."
docker run --rm --network cost-of-living_default \
    -v /tmp/cost-worker:/app/worker \
    -e TEMPORAL_ADDRESS=cost-of-living-temporal:7233 \
    --name cost-worker \
    -d \
    golang:1.24-alpine \
    /app/worker

# Wait for worker to connect
sleep 5

# Run client
echo "Executing workflow..."
docker run --rm --network cost-of-living_default \
    -v /tmp/cost-client:/app/client \
    golang:1.24-alpine \
    /app/client

# Stop worker
echo "Stopping worker..."
docker stop cost-worker

echo "Test completed successfully!"
