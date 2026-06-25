#!/bin/bash
# cleanup-home.sh — Clean up home directory clutter
# Run as: bash scripts/cleanup-home.sh [--dry-run]

set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
    DRY_RUN=true
    echo "DRY RUN MODE — nothing will be deleted"
fi

HOME_DIR="/home/nara"

echo "=== Cleaning up $HOME_DIR ==="

# Function to remove items
remove_item() {
    local item="$1"
    local full_path="$HOME_DIR/$item"
    
    if [[ -e "$full_path" ]]; then
        if $DRY_RUN; then
            echo "  [DRY RUN] Would remove: $item"
        else
            echo "  Removing: $item"
            rm -rf "$full_path"
        fi
    fi
}

# Function to move items
move_item() {
    local item="$1"
    local dest="$2"
    local full_path="$HOME_DIR/$item"
    local dest_path="$HOME_DIR/$dest"
    
    if [[ -e "$full_path" ]]; then
        if $DRY_RUN; then
            echo "  [DRY RUN] Would move: $item → $dest/"
        else
            echo "  Moving: $item → $dest/"
            mkdir -p "$dest_path"
            mv "$full_path" "$dest_path/"
        fi
    fi
}

echo ""
echo "1. Removing stale Go installations..."
remove_item "go"
remove_item "go-1.25"
remove_item "go-local"
remove_item "gopath"
remove_item "go-tools"

echo ""
echo "2. Removing Node.js artifacts..."
remove_item "node_modules"
remove_item "package.json"
remove_item "pnpm-lock.yaml"
remove_item "pnpm-workspace.yaml"

echo ""
echo "3. Removing temp bundles..."
remove_item "vibecode-bundle"
remove_item "vibecode-bundle-20260622.tar.gz"

echo ""
echo "4. Removing stray scripts..."
remove_item "add-bridge-nginx.sh"
remove_item "setup-nginx-tls.sh"

echo ""
echo "5. Removing log files..."
remove_item "paperclip-onboard.log"
remove_item "paperclip-run.log"

echo ""
echo "6. Moving project files..."
move_item "ELF-6-completion-report.md" "Documents"
move_item "README.md" "Documents"

echo ""
echo "7. Verifying clean state..."
echo "Current home directory:"
ls -la "$HOME_DIR" | grep -v "^\." | head -20

echo ""
if $DRY_RUN; then
    echo "=== Dry run complete. Run without --dry-run to apply changes. ==="
else
    echo "=== Cleanup complete! ==="
fi
