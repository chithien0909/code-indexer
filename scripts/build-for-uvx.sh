#!/bin/bash

# Build script for uvx packaging
# This script builds the Go binary and prepares the Python package for uvx installation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Go
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go is not installed. Please install Go from https://golang.org/"
        exit 1
    fi
    
    # Check Python
    if ! command -v python3 >/dev/null 2>&1; then
        log_error "Python 3 is not installed. Please install Python 3."
        exit 1
    fi
    
    # Check if we're in the right directory
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        log_error "Not in the correct project directory. go.mod not found."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Clean previous builds
clean_build() {
    log_info "Cleaning previous builds..."
    
    cd "$PROJECT_ROOT"
    
    # Remove build artifacts
    rm -rf build/
    rm -rf dist/
    rm -rf *.egg-info/
    rm -rf python/mcp_code_indexer/bin/
    
    # Remove Python cache
    find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
    find . -type f -name "*.pyc" -delete 2>/dev/null || true
    
    log_success "Build artifacts cleaned"
}

# Build Go binary
build_go_binary() {
    log_info "Building Go binary..."
    
    cd "$PROJECT_ROOT"
    
    # Create bin directory in Python package
    mkdir -p python/mcp_code_indexer/bin
    
    # Determine platform and architecture
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $arch in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) arch="amd64" ;;  # Default fallback
    esac
    
    # Set binary name
    local binary_name="code-indexer"
    if [[ "$os" == "windows" || "$os" == "mingw"* || "$os" == "cygwin"* ]]; then
        binary_name="code-indexer.exe"
    fi
    
    # Build binary
    local output_path="python/mcp_code_indexer/bin/$binary_name"
    
    log_info "Building for $os/$arch..."
    
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build \
        -o "$output_path" \
        -ldflags "-s -w -X main.version=1.1.0" \
        ./cmd/server
    
    if [[ -f "$output_path" ]]; then
        chmod +x "$output_path"
        log_success "Binary built successfully: $output_path"
        
        # Show binary info
        local size=$(du -h "$output_path" | cut -f1)
        log_info "Binary size: $size"
    else
        log_error "Failed to build binary"
        exit 1
    fi
}

# Build Python package
build_python_package() {
    log_info "Building Python package..."
    
    cd "$PROJECT_ROOT"
    
    # Install build dependencies if needed
    if ! python3 -c "import build" 2>/dev/null; then
        log_info "Installing build dependencies..."
        python3 -m pip install build
    fi
    
    # Build the package
    python3 -m build
    
    if [[ -d "dist" ]]; then
        log_success "Python package built successfully"
        log_info "Package files:"
        ls -la dist/
    else
        log_error "Failed to build Python package"
        exit 1
    fi
}

# Test the package
test_package() {
    log_info "Testing the package..."
    
    cd "$PROJECT_ROOT"
    
    # Test that the binary works
    local binary_path="python/mcp_code_indexer/bin/code-indexer"
    if [[ -f "$binary_path" ]]; then
        log_info "Testing binary..."
        if "$binary_path" --version >/dev/null 2>&1; then
            log_success "Binary test passed"
        else
            log_warning "Binary test failed, but continuing..."
        fi
    fi
    
    # Test Python package import
    log_info "Testing Python package import..."
    if python3 -c "import sys; sys.path.insert(0, 'python'); import mcp_code_indexer; print(f'Version: {mcp_code_indexer.__version__}')" 2>/dev/null; then
        log_success "Python package import test passed"
    else
        log_error "Python package import test failed"
        exit 1
    fi
}

# Create installation instructions
create_install_instructions() {
    log_info "Creating installation instructions..."
    
    cd "$PROJECT_ROOT"
    
    cat > INSTALL_UVX.md << 'EOF'
# UVX Installation Instructions

## Quick Install

```bash
uvx install git+https://github.com/my-mcp/code-indexer.git
```

## IDE Configuration

### Augment Code
```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "--from",
        "git+https://github.com/my-mcp/code-indexer.git",
        "code-indexer",
        "mcp-server"
      ]
    }
  }
}
```

### Cursor IDE
```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx", 
      "args": [
        "--from",
        "git+https://github.com/my-mcp/code-indexer.git",
        "code-indexer",
        "mcp-server"
      ]
    }
  }
}
```

## Manual Testing

```bash
# Test installation
uvx --from git+https://github.com/my-mcp/code-indexer.git code-indexer --version

# Run MCP server
uvx --from git+https://github.com/my-mcp/code-indexer.git code-indexer mcp-server
```

For detailed instructions, see docs/UVX_INSTALLATION.md
EOF
    
    log_success "Installation instructions created: INSTALL_UVX.md"
}

# Main build function
main() {
    log_info "Starting uvx build process..."
    log_info "================================"
    
    check_prerequisites
    clean_build
    build_go_binary
    build_python_package
    test_package
    create_install_instructions
    
    log_info "================================"
    log_success "🎉 Build completed successfully!"
    log_info ""
    log_info "Next steps:"
    log_info "1. Test locally: uvx install ."
    log_info "2. Push to repository for others to install"
    log_info "3. Configure your IDE using the examples in INSTALL_UVX.md"
    log_info ""
    log_info "For detailed setup instructions, see docs/UVX_INSTALLATION.md"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --clean-only)
            clean_build
            exit 0
            ;;
        --go-only)
            check_prerequisites
            build_go_binary
            exit 0
            ;;
        --python-only)
            check_prerequisites
            build_python_package
            exit 0
            ;;
        --test-only)
            test_package
            exit 0
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --clean-only    Only clean build artifacts"
            echo "  --go-only       Only build Go binary"
            echo "  --python-only   Only build Python package"
            echo "  --test-only     Only run tests"
            echo "  --help          Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main build
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
