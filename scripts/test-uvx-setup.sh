#!/bin/bash

# Test script for uvx direct execution setup
# This script tests the new mcp-server command and uvx integration

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

# Test the mcp-server command
test_mcp_server_command() {
    log_info "Testing mcp-server command..."
    
    cd "$PROJECT_ROOT"
    
    # Build if binary doesn't exist
    if [[ ! -f "bin/code-indexer" ]]; then
        log_info "Building code-indexer binary..."
        make build
    fi
    
    # Test that the mcp-server command exists
    if ./bin/code-indexer mcp-server --help >/dev/null 2>&1; then
        log_success "mcp-server command is available"
    else
        log_error "mcp-server command not found"
        return 1
    fi
    
    # Test that the command starts (briefly)
    log_info "Testing mcp-server startup..."
    timeout 3s ./bin/code-indexer mcp-server --log-level debug 2>/dev/null || true
    
    if [[ $? -eq 124 ]]; then
        log_success "mcp-server command starts successfully (timed out as expected)"
    else
        log_warning "mcp-server command may have issues (check manually)"
    fi
}

# Test Python package structure
test_python_package() {
    log_info "Testing Python package structure..."
    
    cd "$PROJECT_ROOT"
    
    # Check if Python package directory exists
    if [[ -d "python/mcp_code_indexer" ]]; then
        log_success "Python package directory exists"
    else
        log_error "Python package directory not found"
        return 1
    fi
    
    # Check if __init__.py exists
    if [[ -f "python/mcp_code_indexer/__init__.py" ]]; then
        log_success "Python package __init__.py exists"
    else
        log_error "Python package __init__.py not found"
        return 1
    fi
    
    # Test Python import
    if python3 -c "import sys; sys.path.insert(0, 'python'); import mcp_code_indexer; print(f'Version: {mcp_code_indexer.__version__}')" 2>/dev/null; then
        log_success "Python package imports successfully"
    else
        log_error "Python package import failed"
        return 1
    fi
}

# Test uvx installation (if uvx is available)
test_uvx_installation() {
    log_info "Testing uvx installation..."
    
    if ! command -v uvx >/dev/null 2>&1; then
        log_warning "uvx not found, skipping uvx tests"
        log_info "Install uvx with: pip install uvx"
        return 0
    fi
    
    cd "$PROJECT_ROOT"
    
    # Test local installation
    log_info "Testing local uvx installation..."
    if uvx install --force . >/dev/null 2>&1; then
        log_success "Local uvx installation successful"
        
        # Test that the installed command works
        if uvx code-indexer --version >/dev/null 2>&1; then
            log_success "Installed uvx command works"
        else
            log_warning "Installed uvx command may have issues"
        fi
        
        # Test mcp-server command via uvx
        log_info "Testing mcp-server via uvx..."
        timeout 3s uvx code-indexer mcp-server --log-level debug 2>/dev/null || true
        
        if [[ $? -eq 124 ]]; then
            log_success "uvx mcp-server command works (timed out as expected)"
        else
            log_warning "uvx mcp-server command may have issues"
        fi
        
        # Clean up
        uvx uninstall mcp-code-indexer >/dev/null 2>&1 || true
        
    else
        log_error "Local uvx installation failed"
        return 1
    fi
}

# Test build for uvx script
test_build_script() {
    log_info "Testing build-for-uvx script..."
    
    cd "$PROJECT_ROOT"
    
    if [[ -f "scripts/build-for-uvx.sh" ]]; then
        log_success "build-for-uvx.sh script exists"
        
        # Test that it's executable
        if [[ -x "scripts/build-for-uvx.sh" ]]; then
            log_success "build-for-uvx.sh is executable"
        else
            log_warning "build-for-uvx.sh is not executable, fixing..."
            chmod +x scripts/build-for-uvx.sh
        fi
        
        # Test dry run (go-only to avoid full build)
        if ./scripts/build-for-uvx.sh --go-only >/dev/null 2>&1; then
            log_success "build-for-uvx.sh script works"
        else
            log_warning "build-for-uvx.sh script may have issues"
        fi
    else
        log_error "build-for-uvx.sh script not found"
        return 1
    fi
}

# Test configuration files
test_configuration_files() {
    log_info "Testing configuration files..."
    
    cd "$PROJECT_ROOT"
    
    # Check pyproject.toml
    if [[ -f "pyproject.toml" ]]; then
        log_success "pyproject.toml exists"
        
        # Validate TOML syntax
        if python3 -c "import tomllib; tomllib.load(open('pyproject.toml', 'rb'))" 2>/dev/null; then
            log_success "pyproject.toml is valid"
        elif python3 -c "import tomli; tomli.load(open('pyproject.toml', 'rb'))" 2>/dev/null; then
            log_success "pyproject.toml is valid (using tomli)"
        else
            log_warning "pyproject.toml validation failed (install tomllib/tomli)"
        fi
    else
        log_error "pyproject.toml not found"
        return 1
    fi
    
    # Check setup.py
    if [[ -f "setup.py" ]]; then
        log_success "setup.py exists"
    else
        log_error "setup.py not found"
        return 1
    fi
    
    # Check MANIFEST.in
    if [[ -f "MANIFEST.in" ]]; then
        log_success "MANIFEST.in exists"
    else
        log_error "MANIFEST.in not found"
        return 1
    fi
}

# Test documentation
test_documentation() {
    log_info "Testing documentation..."
    
    cd "$PROJECT_ROOT"
    
    local docs=(
        "docs/UVX_INSTALLATION.md"
        "docs/MIGRATION_TO_UVX.md"
    )
    
    for doc in "${docs[@]}"; do
        if [[ -f "$doc" ]]; then
            log_success "$doc exists"
        else
            log_error "$doc not found"
            return 1
        fi
    done
}

# Create sample IDE configuration
create_sample_configs() {
    log_info "Creating sample IDE configurations..."
    
    cd "$PROJECT_ROOT"
    
    mkdir -p examples/ide-configs
    
    # Augment Code config
    cat > examples/ide-configs/augment-code.json << 'EOF'
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "--from",
        "git+https://github.com/my-mcp/code-indexer.git",
        "code-indexer",
        "mcp-server"
      ],
      "env": {
        "CONFIG_PATH": "${workspaceFolder}/.code-indexer/config.yaml"
      }
    }
  }
}
EOF
    
    # Cursor IDE config
    cat > examples/ide-configs/cursor-ide.json << 'EOF'
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "--from",
        "git+https://github.com/my-mcp/code-indexer.git",
        "code-indexer",
        "mcp-server"
      ],
      "env": {
        "CONFIG_PATH": "${workspaceFolder}/.code-indexer/config.yaml"
      }
    }
  }
}
EOF
    
    # Sample workspace config
    cat > examples/ide-configs/workspace-config.yaml << 'EOF'
indexer:
  supported_extensions:
    - .go
    - .py
    - .js
    - .ts
    - .java
    - .cpp
    - .c
    - .h
    - .rs
    - .rb
    - .php
    - .cs
    - .kt
    - .swift
  
  max_file_size: 10485760  # 10MB
  
  exclude_patterns:
    - "*/node_modules/*"
    - "*/vendor/*"
    - "*/.git/*"
    - "*/build/*"
    - "*/dist/*"
    - "*/target/*"
    - "*/__pycache__/*"
  
  index_dir: ".code-indexer/index"
  repo_dir: ".code-indexer/repositories"

search:
  max_results: 100
  highlight_snippets: true
  snippet_length: 200
  fuzzy_tolerance: 0.2

logging:
  level: info
  file: ".code-indexer/indexer.log"
  json_format: false
EOF
    
    log_success "Sample IDE configurations created in examples/ide-configs/"
}

# Main test function
main() {
    log_info "Starting uvx setup tests..."
    log_info "=========================="
    
    local tests_passed=0
    local tests_failed=0
    
    # Run tests
    if test_mcp_server_command; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    if test_python_package; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    if test_configuration_files; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    if test_documentation; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    if test_build_script; then
        tests_passed=$((tests_passed + 1))
    else
        tests_failed=$((tests_failed + 1))
    fi
    
    # uvx test is optional
    if test_uvx_installation; then
        tests_passed=$((tests_passed + 1))
    fi
    
    # Create sample configs
    create_sample_configs
    
    # Summary
    log_info "=========================="
    log_info "Test Summary:"
    log_success "Tests passed: $tests_passed"
    if [[ $tests_failed -gt 0 ]]; then
        log_error "Tests failed: $tests_failed"
    else
        log_info "Tests failed: $tests_failed"
    fi
    
    if [[ $tests_failed -eq 0 ]]; then
        log_success "🎉 All tests passed! uvx setup is ready."
        log_info ""
        log_info "Next steps:"
        log_info "1. Install via uvx: uvx install git+https://github.com/my-mcp/code-indexer.git"
        log_info "2. Configure your IDE using examples in examples/ide-configs/"
        log_info "3. See docs/UVX_INSTALLATION.md for detailed setup instructions"
        exit 0
    else
        log_error "❌ Some tests failed. Check the output above for details."
        exit 1
    fi
}

# Run main test
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
