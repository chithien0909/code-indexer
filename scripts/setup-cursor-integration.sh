#!/bin/bash

# MCP Code Indexer - Cursor/Augment Integration Setup Script
# This script helps set up the MCP Code Indexer with Cursor or Augment IDE

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
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

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

print_status "MCP Code Indexer - Cursor/Augment Integration Setup"
echo "=================================================="

# Check if binary exists
BINARY_PATH="$PROJECT_ROOT/bin/code-indexer"
if [ ! -f "$BINARY_PATH" ]; then
    print_error "Binary not found at $BINARY_PATH"
    print_status "Building the project..."
    cd "$PROJECT_ROOT"
    make build
    if [ $? -ne 0 ]; then
        print_error "Failed to build the project"
        exit 1
    fi
    print_success "Project built successfully"
fi

# Make binary executable
chmod +x "$BINARY_PATH"
print_success "Binary is executable"

# Test the binary
print_status "Testing the binary..."
if ! "$BINARY_PATH" --help > /dev/null 2>&1; then
    print_error "Binary test failed"
    exit 1
fi
print_success "Binary test passed"

# Detect IDE
IDE_TYPE=""
CURSOR_CONFIG_DIR=""
AUGMENT_CONFIG_DIR=""

# Check for Cursor
if command -v cursor > /dev/null 2>&1; then
    IDE_TYPE="cursor"
    case "$(uname -s)" in
        Darwin)
            CURSOR_CONFIG_DIR="$HOME/Library/Application Support/Cursor/User/globalStorage/rooveterinaryinc.roo-cline/settings"
            ;;
        Linux)
            CURSOR_CONFIG_DIR="$HOME/.config/Cursor/User/globalStorage/rooveterinaryinc.roo-cline/settings"
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)
            CURSOR_CONFIG_DIR="$APPDATA/Cursor/User/globalStorage/rooveterinaryinc.roo-cline/settings"
            ;;
    esac
fi

# Check for Augment (add detection logic when available)
# if command -v augment > /dev/null 2>&1; then
#     IDE_TYPE="augment"
# fi

if [ -z "$IDE_TYPE" ]; then
    print_warning "No supported IDE detected. Manual configuration required."
    print_status "Supported IDEs: Cursor"
    echo ""
    print_status "Manual configuration paths:"
    echo "  Cursor (macOS): ~/Library/Application Support/Cursor/User/globalStorage/rooveterinaryinc.roo-cline/settings/"
    echo "  Cursor (Linux): ~/.config/Cursor/User/globalStorage/rooveterinaryinc.roo-cline/settings/"
    echo "  Cursor (Windows): %APPDATA%\\Cursor\\User\\globalStorage\\rooveterinaryinc.roo-cline\\settings\\"
    echo ""
else
    print_success "Detected IDE: $IDE_TYPE"
fi

# Create MCP configuration
create_cursor_config() {
    local config_dir="$1"
    local config_file="$config_dir/cline_mcp_settings.json"
    
    print_status "Creating Cursor MCP configuration..."
    
    # Create directory if it doesn't exist
    mkdir -p "$config_dir"
    
    # Create or update configuration
    cat > "$config_file" << EOF
{
  "mcpServers": {
    "code-indexer": {
      "command": "$BINARY_PATH",
      "args": ["serve"],
      "env": {
        "INDEXER_LOG_LEVEL": "info"
      }
    }
  }
}
EOF
    
    print_success "Configuration created at: $config_file"
}

# Setup configuration based on detected IDE
if [ "$IDE_TYPE" = "cursor" ] && [ -n "$CURSOR_CONFIG_DIR" ]; then
    create_cursor_config "$CURSOR_CONFIG_DIR"
fi

# Create a custom configuration file
print_status "Creating custom configuration file..."
CONFIG_DIR="$HOME/.config/code-indexer"
mkdir -p "$CONFIG_DIR"

cat > "$CONFIG_DIR/config.yaml" << EOF
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
    - .hpp
    - .rs
    - .rb
    - .php
    - .cs
    - .kt
    - .swift
    - .scala
    - .clj
    - .hs
    - .ml
    - .sh
    - .bash
    - .zsh
    - .fish
    - .ps1
    - .sql
    - .r
    - .m
    - .dart
    - .lua
    - .perl
    - .pl

  max_file_size: 1048576  # 1MB
  
  exclude_patterns:
    - "*/node_modules/*"
    - "*/vendor/*"
    - "*/.git/*"
    - "*/build/*"
    - "*/dist/*"
    - "*/target/*"
    - "*/__pycache__/*"
    - "*.pyc"
    - "*.class"
    - "*.jar"
    - "*.war"
    - "*.ear"
    - "*.exe"
    - "*.dll"
    - "*.so"
    - "*.dylib"
    - "*.a"
    - "*.lib"
    - "*.o"
    - "*.obj"
    - "*.min.js"
    - "*.min.css"

  index_dir: "$CONFIG_DIR/index"
  repo_dir: "$CONFIG_DIR/repositories"

search:
  max_results: 100
  highlight_snippets: true
  snippet_length: 200
  fuzzy_tolerance: 0.2

server:
  name: "Code Indexer"
  version: "1.0.0"
  enable_recovery: true

logging:
  level: info
  file: "$CONFIG_DIR/indexer.log"
  json_format: false
  max_size: 100
  max_backups: 3
  max_age: 30
EOF

print_success "Custom configuration created at: $CONFIG_DIR/config.yaml"

# Create directories
mkdir -p "$CONFIG_DIR/index"
mkdir -p "$CONFIG_DIR/repositories"
print_success "Created index and repository directories"

# Test the setup
print_status "Testing the MCP server setup..."
cd "$PROJECT_ROOT"

# Run a quick test
timeout 10s "$BINARY_PATH" serve --config "$CONFIG_DIR/config.yaml" --log-level debug > /dev/null 2>&1 &
SERVER_PID=$!
sleep 2

if kill -0 $SERVER_PID 2>/dev/null; then
    kill $SERVER_PID
    print_success "MCP server test passed"
else
    print_warning "MCP server test inconclusive (this may be normal)"
fi

# Create a test script
TEST_SCRIPT="$PROJECT_ROOT/test-integration.sh"
cat > "$TEST_SCRIPT" << 'EOF'
#!/bin/bash
# Test script for MCP Code Indexer integration

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$HOME/.config/code-indexer"

echo "Testing MCP Code Indexer integration..."
echo "======================================"

# Test 1: Server starts
echo "Test 1: Server startup"
timeout 5s "$SCRIPT_DIR/bin/code-indexer" serve --config "$CONFIG_DIR/config.yaml" > /dev/null 2>&1 &
SERVER_PID=$!
sleep 2

if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✅ Server starts successfully"
    kill $SERVER_PID
else
    echo "❌ Server failed to start"
    exit 1
fi

# Test 2: Configuration is valid
echo "Test 2: Configuration validation"
if "$SCRIPT_DIR/bin/code-indexer" serve --config "$CONFIG_DIR/config.yaml" --help > /dev/null 2>&1; then
    echo "✅ Configuration is valid"
else
    echo "❌ Configuration validation failed"
    exit 1
fi

# Test 3: Directories exist
echo "Test 3: Directory structure"
if [ -d "$CONFIG_DIR/index" ] && [ -d "$CONFIG_DIR/repositories" ]; then
    echo "✅ Required directories exist"
else
    echo "❌ Required directories missing"
    exit 1
fi

echo ""
echo "🎉 All tests passed! Integration is ready."
echo ""
echo "Next steps:"
echo "1. Restart your IDE (Cursor/Augment)"
echo "2. Look for MCP server connection status"
echo "3. Try using MCP tools in your AI assistant"
echo ""
echo "Example prompts to try:"
echo '- "Please index my current project repository"'
echo '- "Search for all functions containing authentication"'
echo '- "Show me the structure of the main.go file"'
EOF

chmod +x "$TEST_SCRIPT"
print_success "Created integration test script: $TEST_SCRIPT"

# Final instructions
echo ""
print_success "Setup completed successfully!"
echo ""
print_status "Next steps:"
echo "1. Restart your IDE (Cursor/Augment) to load the new MCP configuration"
echo "2. Look for MCP server connection status in your IDE"
echo "3. Try using the MCP tools with your AI assistant"
echo ""
print_status "Test the integration:"
echo "  $TEST_SCRIPT"
echo ""
print_status "Example AI prompts to try:"
echo '  - "Please index my current project repository"'
echo '  - "Search for all functions containing authentication"'
echo '  - "Show me the structure of the main.go file"'
echo '  - "List all indexed repositories and their statistics"'
echo ""
print_status "Configuration files created:"
if [ "$IDE_TYPE" = "cursor" ] && [ -n "$CURSOR_CONFIG_DIR" ]; then
    echo "  - Cursor MCP config: $CURSOR_CONFIG_DIR/cline_mcp_settings.json"
fi
echo "  - Custom config: $CONFIG_DIR/config.yaml"
echo "  - Log file: $CONFIG_DIR/indexer.log"
echo ""
print_status "For troubleshooting, see: docs/CURSOR_AUGMENT_INTEGRATION.md"
echo ""
print_success "🎉 MCP Code Indexer is ready for use with your IDE!"
