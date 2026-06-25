#!/bin/bash
# Deploy static site to hatch.surf server via Docker
# Usage: ./scripts/deploy-site.sh [--dry-run]
#
# Environment variables required:
#   DEPLOY_HOST  - SSH host (e.g., 46.250.250.48)
#   DEPLOY_USER  - SSH user (e.g., root)
#   DEPLOY_KEY   - SSH private key (base64 encoded, optional if using ssh-agent)
#
# This script builds a Docker image with static files baked in and deploys it.
# No /var/www mounting required - the image is self-contained.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SITE_DIR="$REPO_ROOT/site"

# Configuration
DEPLOY_HOST="${DEPLOY_HOST:-46.250.250.48}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_KEY="${DEPLOY_KEY:-}"
CONTAINER_NAME="hatch-homepage"
IMAGE_NAME="hatch-homepage:latest"

# Parse arguments
DRY_RUN=false
for arg in "$@"; do
    case $arg in
        --dry-run)
            DRY_RUN=true
            ;;
    esac
done

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}✓${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1" >&2
    exit 1
}

# Validate
if [ ! -d "$SITE_DIR" ]; then
    error "Site directory not found: $SITE_DIR"
fi

if [ ! -f "$SITE_DIR/index.html" ]; then
    error "index.html not found in $SITE_DIR"
fi

if [ ! -f "$SITE_DIR/Dockerfile" ]; then
    error "Dockerfile not found in $SITE_DIR"
fi

# Setup SSH key if provided
SSH_OPTS="-o StrictHostKeyChecking=no -o ConnectTimeout=10"
if [ -n "$DEPLOY_KEY" ]; then
    SSH_KEY_FILE=$(mktemp)
    echo "$DEPLOY_KEY" | base64 -d > "$SSH_KEY_FILE"
    chmod 600 "$SSH_KEY_FILE"
    SSH_OPTS="$SSH_OPTS -i $SSH_KEY_FILE"
fi

DEPLOY_TARGET="${DEPLOY_USER}@${DEPLOY_HOST}"

if [ "$DRY_RUN" = true ]; then
    warn "Dry run - would build and deploy Docker image to $DEPLOY_TARGET"
    warn "Image: $IMAGE_NAME"
    warn "Container: $CONTAINER_NAME"
else
    image_content=""
    log "Building Docker image..."
    docker build --no-cache -t "$IMAGE_NAME" "$SITE_DIR"
    
    # Verify the built image contains correct content
    log "Verifying built image content..."
    image_content=$(docker run --rm "$IMAGE_NAME" cat /usr/share/nginx/html/index.html 2>/dev/null | head -20)
    
    if echo "$image_content" | grep -q "Hatch.*Self-hostable HTTP request inspector"; then
        log "Built image contains correct content"
    else
        error "Built image may contain stale content!"
        echo "Expected: Hatch - Self-hostable HTTP request inspector"
        echo "Got: $image_content"
        exit 1
    fi
    
    log "Saving Docker image..."
    docker save "$IMAGE_NAME" | gzip > /tmp/hatch-homepage.tar.gz
    
    log "Uploading image to $DEPLOY_TARGET..."
    scp $SSH_OPTS /tmp/hatch-homepage.tar.gz "$DEPLOY_TARGET:/tmp/hatch-homepage.tar.gz"
    
    log "Deploying container on $DEPLOY_TARGET..."
    ssh $SSH_OPTS "$DEPLOY_TARGET" << 'ENDSSH'
        # Backup current image before loading new one
        if docker images hatch-homepage:latest -q | grep -q .; then
            echo "Backing up current image as hatch-homepage:backup..."
            docker tag hatch-homepage:latest hatch-homepage:backup 2>/dev/null || true
        fi
        
        # Load the image
        docker load < /tmp/hatch-homepage.tar.gz
        
        # Stop and remove old container if exists
        docker stop hatch-homepage 2>/dev/null || true
        docker rm hatch-homepage 2>/dev/null || true
        
        # Run new container (static files baked into image, no host mount)
        docker run -d \
          --name hatch-homepage \
          --restart unless-stopped \
          -p 127.0.0.1:3000:80 \
          hatch-homepage:latest
        
        # Clean up
        rm -f /tmp/hatch-homepage.tar.gz
        
        # Reload nginx if it's managing the reverse proxy
        nginx -t && systemctl reload nginx 2>/dev/null || true
ENDSSH
    
    # Clean up local temp file
    rm -f /tmp/hatch-homepage.tar.gz
    
    log "Deployment complete!"
    
    # Run verification to ensure correct content is being served
    log "Running deployment verification..."
    if "$SCRIPT_DIR/verify-deployment.sh" --host "$DEPLOY_HOST" --user "$DEPLOY_USER" ${DEPLOY_KEY:+--key "$DEPLOY_KEY"}; then
        log "Verification passed - correct content is being served"
        log "Site live at https://hatch.surf"
    else
        error "Deployment verification failed!"
        warn "Rolling back deployment..."
        
        # Rollback: restore previous container if backup exists
        ssh $SSH_OPTS "$DEPLOY_TARGET" << 'ROLLBACK'
            # Check if we have a backup image
            if docker images hatch-homepage:backup -q | grep -q .; then
                echo "Restoring from backup image..."
                docker stop hatch-homepage 2>/dev/null || true
                docker rm hatch-homepage 2>/dev/null || true
                docker run -d \
                  --name hatch-homepage \
                  --restart unless-stopped \
                  -p 127.0.0.1:3000:80 \
                  hatch-homepage:backup
                echo "Rollback complete"
            else
                echo "No backup image available for rollback"
                echo "Manual intervention required"
                exit 1
            fi
ROLLBACK
        
        error "Deployment failed verification and rollback attempted"
        error "Please check the server manually"
        exit 1
    fi
fi

# Cleanup SSH key if created
if [ -n "${SSH_KEY_FILE:-}" ]; then
    rm -f "$SSH_KEY_FILE"
fi
