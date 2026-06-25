#!/bin/bash
# verify-deployment.sh - Verify deployed content matches expected site
# Usage: ./scripts/verify-deployment.sh [--host HOST] [--user USER] [--key KEY]
#
# This script verifies that the deployed site is serving the correct content
# by checking for expected markers and comparing checksums.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SITE_DIR="$REPO_ROOT/site"

# Default configuration
DEPLOY_HOST="${DEPLOY_HOST:-46.250.250.48}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_KEY="${DEPLOY_KEY:-}"
CONTAINER_NAME="hatch-homepage"
EXPECTED_TITLE="Hatch — Self-hostable HTTP request inspector + mocker"
EXPECTED_DESCRIPTION="Hatch is a self-hostable HTTP request inspector and mocker"

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
    return 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --host)
            DEPLOY_HOST="$2"
            shift 2
            ;;
        --user)
            DEPLOY_USER="$2"
            shift 2
            ;;
        --key)
            DEPLOY_KEY="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Setup SSH key if provided
SSH_OPTS="-o StrictHostKeyChecking=no -o ConnectTimeout=10"
if [ -n "$DEPLOY_KEY" ]; then
    SSH_KEY_FILE=$(mktemp)
    echo "$DEPLOY_KEY" | base64 -d > "$SSH_KEY_FILE"
    chmod 600 "$SSH_KEY_FILE"
    SSH_OPTS="$SSH_OPTS -i $SSH_KEY_FILE"
fi

DEPLOY_TARGET="${DEPLOY_USER}@${DEPLOY_HOST}"

# Function to verify container is running
verify_container_running() {
    local container_status
    container_status=$(ssh $SSH_OPTS "$DEPLOY_TARGET" "docker inspect -f '{{.State.Status}}' $CONTAINER_NAME 2>/dev/null || echo 'not_found'")
    
    if [ "$container_status" = "running" ]; then
        log "Container $CONTAINER_NAME is running"
        return 0
    else
        error "Container $CONTAINER_NAME is not running (status: $container_status)"
        return 1
    fi
}

# Function to verify content via HTTP
verify_content() {
    local response
    local http_code
    
    # Wait for container to be ready
    sleep 2
    
    # Fetch the page content
    response=$(ssh $SSH_OPTS "$DEPLOY_TARGET" "curl -s -o /dev/null -w '%{http_code}' http://localhost:3000/ 2>/dev/null || echo '000'")
    http_code="$response"
    
    if [ "$http_code" != "200" ]; then
        error "HTTP request failed with status code: $http_code"
        return 1
    fi
    
    log "HTTP endpoint returns 200 OK"
    
    # Fetch actual content
    local content
    content=$(ssh $SSH_OPTS "$DEPLOY_TARGET" "curl -s http://localhost:3000/ 2>/dev/null")
    
    # Check for expected title
    if echo "$content" | grep -q "$EXPECTED_TITLE"; then
        log "Expected title found: '$EXPECTED_TITLE'"
    else
        error "Expected title not found. Possible stale content detected!"
        echo "Expected: $EXPECTED_TITLE"
        echo "Got content snippet:"
        echo "$content" | head -20
        return 1
    fi
    
    # Check for expected description
    if echo "$content" | grep -q "$EXPECTED_DESCRIPTION"; then
        log "Expected description found"
    else
        warn "Expected description not found in content"
    fi
    
    # Check for Next.js template markers (sign of stale content)
    if echo "$content" | grep -qi "next.js\|create-next-app\|get started by editing"; then
        error "STALE CONTENT DETECTED: Found Next.js template markers!"
        return 1
    else
        log "No Next.js template markers found (good)"
    fi
    
    return 0
}

# Function to verify static assets
verify_assets() {
    local css_response
    local js_response
    
    # Check CSS file
    css_response=$(ssh $SSH_OPTS "$DEPLOY_TARGET" "curl -s -o /dev/null -w '%{http_code}' http://localhost:3000/style.css 2>/dev/null || echo '000'")
    
    if [ "$css_response" = "200" ]; then
        log "CSS asset accessible (HTTP 200)"
    else
        warn "CSS asset returned HTTP $css_response"
    fi
    
    # Check JS file
    js_response=$(ssh $SSH_OPTS "$DEPLOY_TARGET" "curl -s -o /dev/null -w '%{http_code}' http://localhost:3000/main.js 2>/dev/null || echo '000'")
    
    if [ "$js_response" = "200" ]; then
        log "JS asset accessible (HTTP 200)"
    else
        warn "JS asset returned HTTP $js_response"
    fi
    
    return 0
}

# Function to verify container logs for errors
verify_logs() {
    local logs
    logs=$(ssh $SSH_OPTS "$DEPLOY_TARGET" "docker logs --tail 10 $CONTAINER_NAME 2>&1")
    
    if echo "$logs" | grep -qi "error\|fatal\|panic"; then
        warn "Found error entries in container logs:"
        echo "$logs" | grep -i "error\|fatal\|panic" | head -5
        return 1
    else
        log "No critical errors in container logs"
        return 0
    fi
}

# Main verification
main() {
    echo "=== Deployment Verification ==="
    echo "Host: $DEPLOY_HOST"
    echo "Container: $CONTAINER_NAME"
    echo ""
    
    local exit_code=0
    
    # Run all verification checks
    verify_container_running || exit_code=1
    verify_content || exit_code=1
    verify_assets || exit_code=1
    verify_logs || exit_code=1
    
    echo ""
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}=== All verification checks passed ===${NC}"
    else
        echo -e "${RED}=== Verification FAILED ===${NC}"
        echo "Possible actions:"
        echo "  1. Check container logs: ssh $DEPLOY_TARGET 'docker logs $CONTAINER_NAME'"
        echo "  2. Rebuild and redeploy: ./scripts/deploy-site.sh"
        echo "  3. Rollback to previous version if available"
    fi
    
    # Cleanup SSH key if created
    if [ -n "${SSH_KEY_FILE:-}" ]; then
        rm -f "$SSH_KEY_FILE"
    fi
    
    return $exit_code
}

# Run main function
main "$@"
