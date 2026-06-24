#!/bin/bash

# Hatch CLI Setup and Verification Script
# This script helps users install and verify the Hatch CLI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    
    case $status in
        "OK")
            echo -e "${GREEN}✓${NC} $message"
            ;;
        "WARN")
            echo -e "${YELLOW}⚠${NC} $message"
            ;;
        "ERROR")
            echo -e "${RED}✗${NC} $message"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ${NC} $message"
            ;;
    esac
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to get latest release version from GitHub
get_latest_version() {
    if command_exists curl; then
        curl -s https://api.github.com/repos/elfoundation/hatch/releases/latest | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4
    elif command_exists wget; then
        wget -qO- https://api.github.com/repos/elfoundation/hatch/releases/latest | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4
    else
        echo "unknown"
    fi
}

# Function to detect platform
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux)
            case $arch in
                x86_64|amd64)
                    echo "linux-amd64"
                    ;;
                aarch64|arm64)
                    echo "linux-arm64"
                    ;;
                *)
                    echo "linux-unknown"
                    ;;
            esac
            ;;
        darwin)
            case $arch in
                x86_64|amd64)
                    echo "darwin-amd64"
                    ;;
                arm64)
                    echo "darwin-arm64"
                    ;;
                *)
                    echo "darwin-unknown"
                    ;;
            esac
            ;;
        mingw*|msys*|cygwin*)
            echo "windows-amd64"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Function to install hatch
install_hatch() {
    local platform=$(detect_platform)
    local version=$(get_latest_version)
    
    if [ "$platform" = "unknown" ]; then
        print_status "ERROR" "Unsupported platform: $(uname -s) $(uname -m)"
        echo "Please install manually from https://github.com/elfoundation/hatch/releases"
        exit 1
    fi
    
    if [ "$version" = "unknown" ] || [ -z "$version" ]; then
        print_status "WARN" "Could not determine latest version"
        print_status "INFO" "Please check https://github.com/elfoundation/hatch/releases"
        exit 1
    fi
    
    local binary_name="hatch-${platform}"
    local download_url="https://github.com/elfoundation/hatch/releases/download/${version}/${binary_name}"
    
    if [ "$platform" = "windows-amd64" ]; then
        binary_name="hatch-windows-amd64.exe"
        download_url="https://github.com/elfoundation/hatch/releases/download/${version}/${binary_name}"
    fi
    
    print_status "INFO" "Detected platform: ${platform}"
    print_status "INFO" "Latest version: ${version}"
    print_status "INFO" "Download URL: ${download_url}"
    
    # Create temporary directory
    local temp_dir=$(mktemp -d)
    local temp_file="${temp_dir}/${binary_name}"
    
    print_status "INFO" "Downloading hatch..."
    
    # Download binary
    if command_exists curl; then
        curl -L -o "$temp_file" "$download_url"
    elif command_exists wget; then
        wget -O "$temp_file" "$download_url"
    else
        print_status "ERROR" "Neither curl nor wget found. Please install one."
        exit 1
    fi
    
    if [ ! -f "$temp_file" ]; then
        print_status "ERROR" "Failed to download binary"
        exit 1
    fi
    
    # Make executable (Unix only)
    if [[ "$platform" != windows-* ]]; then
        chmod +x "$temp_file"
    fi
    
    # Determine installation directory
    local install_dir=""
    if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
        install_dir="/usr/local/bin"
    elif [ -d "$HOME/.local/bin" ]; then
        install_dir="$HOME/.local/bin"
        # Ensure PATH includes ~/.local/bin
        if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
            print_status "WARN" "Adding ~/.local/bin to PATH"
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc" 2>/dev/null || true
        fi
    else
        install_dir="$HOME/.local/bin"
        mkdir -p "$install_dir"
    fi
    
    local install_path="${install_dir}/hatch"
    
    # Install binary
    if [ -w "$install_dir" ]; then
        mv "$temp_file" "$install_path"
    else
        print_status "INFO" "Using sudo to install to ${install_dir}"
        sudo mv "$temp_file" "$install_path"
    fi
    
    # Clean up
    rm -rf "$temp_dir"
    
    print_status "OK" "Hatch installed to ${install_path}"
    
    # Verify installation
    if command_exists hatch; then
        print_status "OK" "Hatch is in PATH"
        hatch version
    else
        print_status "WARN" "Hatch installed but not in PATH"
        print_status "INFO" "You may need to restart your shell or add ${install_dir} to PATH"
    fi
}

# Function to verify installation
verify_installation() {
    print_status "INFO" "Verifying Hatch installation..."
    
    # Check if hatch is installed
    if ! command_exists hatch; then
        print_status "ERROR" "Hatch is not installed or not in PATH"
        echo "Please install hatch first:"
        echo "  $0 install"
        exit 1
    fi
    
    # Check version
    print_status "INFO" "Checking version..."
    if hatch version; then
        print_status "OK" "Version check passed"
    else
        print_status "ERROR" "Version check failed"
        exit 1
    fi
    
    # Check help
    print_status "INFO" "Checking help..."
    if hatch --help > /dev/null 2>&1; then
        print_status "OK" "Help command works"
    else
        print_status "ERROR" "Help command failed"
        exit 1
    fi
    
    # Check if server is running
    print_status "INFO" "Checking server connectivity..."
    if curl -s http://localhost:8080/healthz > /dev/null 2>&1; then
        print_status "OK" "Server is running and accessible"
        
        # Test capture command (dry run)
        print_status "INFO" "Testing CLI commands..."
        
        # Test inspect (will fail if no endpoints, but that's OK)
        if hatch inspect default -limit 1 > /dev/null 2>&1; then
            print_status "OK" "CLI commands working"
        else
            print_status "WARN" "CLI commands may have issues (server might not have data)"
        fi
    else
        print_status "WARN" "Server not running at http://localhost:8080"
        print_status "INFO" "Start server with: hatch serve"
    fi
    
    print_status "OK" "Installation verification complete"
}

# Function to show usage
show_usage() {
    echo "Hatch CLI Setup and Verification Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  install     Download and install the latest hatch binary"
    echo "  verify      Verify hatch installation and connectivity"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 install   # Install hatch"
    echo "  $0 verify    # Verify installation"
    echo ""
    echo "Manual installation:"
    echo "  1. Download from https://github.com/elfoundation/hatch/releases"
    echo "  2. chmod +x hatch-*"
    echo "  3. sudo mv hatch-* /usr/local/bin/hatch"
}

# Main script
main() {
    local command=${1:-help}
    
    case $command in
        install)
            install_hatch
            ;;
        verify)
            verify_installation
            ;;
        help|*)
            show_usage
            ;;
    esac
}

main "$@"