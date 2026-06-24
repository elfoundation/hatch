#!/bin/bash
# Deploy hatch.surf
# Run this script to build and deploy the Docker container

set -e

echo "=== Deploying hatch.surf ==="

# Build the Docker image
echo "Building Docker image..."
docker build -t hatch-web:latest .

# Stop existing container
echo "Stopping existing container..."
docker stop hatch-web 2>/dev/null || true
docker rm hatch-web 2>/dev/null || true

# Start new container
echo "Starting new container..."
docker compose up -d

# Wait for health check
echo "Waiting for container to be healthy..."
sleep 5

# Verify
if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080 | grep -q "200"; then
    echo "✅ hatch.surf deployed successfully!"
    echo "   Container: hatch-web"
    echo "   Port: 8080"
    echo "   Status: $(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080)"
else
    echo "❌ Deployment failed!"
    docker logs hatch-web
    exit 1
fi
