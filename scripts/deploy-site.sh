#!/bin/bash
# Deploy static site to hatch.surf server
# Usage: ./scripts/deploy-site.sh [--dry-run]
#
# Environment variables required:
#   DEPLOY_HOST  - SSH host (e.g., 46.250.250.48)
#   DEPLOY_USER  - SSH user (e.g., root)
#   DEPLOY_KEY   - SSH private key (base64 encoded, optional if using ssh-agent)
#
# The script deploys the contents of the site/ directory to /var/www/hatch.surf

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SITE_DIR="$REPO_ROOT/site"
REMOTE_DIR="/var/www/hatch.surf"

# Configuration
DEPLOY_HOST="${DEPLOY_HOST:-46.250.250.48}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_KEY="${DEPLOY_KEY:-}"

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

# Setup SSH key if provided
SSH_OPTS="-o StrictHostKeyChecking=no -o ConnectTimeout=10"
if [ -n "$DEPLOY_KEY" ]; then
    SSH_KEY_FILE=$(mktemp)
    echo "$DEPLOY_KEY" | base64 -d > "$SSH_KEY_FILE"
    chmod 600 "$SSH_KEY_FILE"
    SSH_OPTS="$SSH_OPTS -i $SSH_KEY_FILE"
fi

# Deploy
DEPLOY_TARGET="${DEPLOY_USER}@${DEPLOY_HOST}"

if [ "$DRY_RUN" = true ]; then
    warn "Dry run - would deploy to $DEPLOY_TARGET:$REMOTE_DIR"
    rsync -avz --delete --dry-run $SSH_OPTS "$SITE_DIR/" "$DEPLOY_TARGET:$REMOTE_DIR/"
else
    log "Deploying to $DEPLOY_TARGET:$REMOTE_DIR"
    rsync -avz --delete $SSH_OPTS "$SITE_DIR/" "$DEPLOY_TARGET:$REMOTE_DIR/"
    
    # Reload nginx
    log "Reloading nginx..."
    ssh $SSH_OPTS "$DEPLOY_TARGET" "nginx -t && systemctl reload nginx"
    
    log "Deployment complete!"
    log "Site live at https://hatch.surf"
fi

# Cleanup SSH key if created
if [ -n "$SSH_KEY_FILE" ]; then
    rm -f "$SSH_KEY_FILE"
fi
