#!/bin/bash

# Go-to-Run installer script
# Usage: curl -sSL https://raw.githubusercontent.com/13winged/go-to-run/main/scripts/install.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variables
REPO="13winged/go-to-run"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="go-to-run"
TEMP_DIR=$(mktemp -d)

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    echo "${OS}_${ARCH}"
}

# Get latest release from GitHub
get_latest_release() {
    curl -s "https://api.github.com/repos/$REPO/releases/latest" | \
    grep '"tag_name":' | \
    sed -E 's/.*"([^"]+)".*/\1/'
}

# Download binary
download_binary() {
    local version=$1
    local platform=$2
    local url="https://github.com/$REPO/releases/download/$version/go-to-run-$platform"
    
    print_info "Downloading $BINARY_NAME $version for $platform..."
    curl -sSL -o "$TEMP_DIR/$BINARY_NAME" "$url"
    
    if [ $? -ne 0 ]; then
        print_error "Failed to download binary"
        exit 1
    fi
    
    chmod +x "$TEMP_DIR/$BINARY_NAME"
}

# Install binary
install_binary() {
    print_info "Installing to $INSTALL_DIR..."
    
    # Check if already installed
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_warning "$BINARY_NAME already installed in $INSTALL_DIR"
        read -p "Overwrite? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Installation cancelled"
            exit 0
        fi
    fi
    
    # Install
    sudo cp "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    
    if [ $? -eq 0 ]; then
        print_success "$BINARY_NAME installed successfully to $INSTALL_DIR"
    else
        print_error "Failed to install $BINARY_NAME"
        exit 1
    fi
}

# Verify installation
verify_installation() {
    if command -v $BINARY_NAME >/dev/null 2>&1; then
        print_success "Verification successful!"
        echo
        echo "Usage:"
        echo "  sudo $BINARY_NAME           # Full system setup"
        echo "  $BINARY_NAME --info        # System information"
        echo "  $BINARY_NAME --help        # Show help"
        echo
        echo "Documentation: https://github.com/$REPO"
    else
        print_error "Installation verification failed"
        exit 1
    fi
}

# Cleanup
cleanup() {
    rm -rf "$TEMP_DIR"
}

# Main installation
main() {
    print_info "Starting $BINARY_NAME installation..."
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1; then
        print_error "curl is required but not installed"
        exit 1
    fi
    
    # Detect platform
    PLATFORM=$(detect_platform)
    
    # Get latest version
    VERSION=$(get_latest_release)
    if [ -z "$VERSION" ]; then
        print_warning "Could not get latest version, using 'latest'"
        VERSION="latest"
    fi
    
    # Download and install
    download_binary "$VERSION" "$PLATFORM"
    install_binary
    verify_installation
    
    # Cleanup
    cleanup
}

# Run main function
trap cleanup EXIT
main