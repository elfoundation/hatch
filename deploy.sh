#!/bin/bash
# Deploy hatch.surf - triggered by Gitea webhook or manually

set -e

echo "=== Deploying hatch.surf ==="
cd /home/nara/apps/web

# Pull latest changes from GitHub (primary source)
git pull github main

# Push to Gitea (push-mirror)
git push origin main

# Build and deploy from site directory
cd site
docker build -t hatch-web:latest .
docker stop hatch-web 2>/dev/null || true
docker rm hatch-web 2>/dev/null || true
docker run -d --name hatch-web --restart unless-stopped -p 8080:80 hatch-web:latest
cd ..

# Verify
sleep 5
if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080 | grep -q "200"; then
    echo "✅ hatch.surf deployed successfully!"
else
    echo "❌ Deployment failed!"
    docker logs hatch-web
    exit 1
fi
