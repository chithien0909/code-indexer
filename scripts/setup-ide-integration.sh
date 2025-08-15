#!/bin/bash

# IDE Integration Setup Script
# Automatically configures MCP Code Indexer with Augment Code and Cursor IDE

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVER_PORT=8080
SERVER_HOST="localhost"

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

# Detect operating system
detect_os() {
    case "$(uname -s)" in
        Darwin*)    echo "macos" ;;
        Linux*)     echo "linux" ;;
        CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
        *)          echo "unknown" ;;
    esac
}

# Get configuration directories for different IDEs
get_config_dirs() {
    local os=$(detect_os)
    
    case $os in
        "macos")
            CURSOR_CONFIG_DIR="$HOME/Library/Application Support/Cursor/User"
            AUGMENT_CONFIG_DIR="$HOME/Library/Application Support/Augment"
            VSCODE_CONFIG_DIR="$HOME/Library/Application Support/Code/User"
            ;;
        "linux")
            CURSOR_CONFIG_DIR="$HOME/.config/Cursor/User"
            AUGMENT_CONFIG_DIR="$HOME/.config/augment"
            VSCODE_CONFIG_DIR="$HOME/.config/Code/User"
            ;;
        "windows")
            CURSOR_CONFIG_DIR="$APPDATA/Cursor/User"
            AUGMENT_CONFIG_DIR="$APPDATA/Augment"
            VSCODE_CONFIG_DIR="$APPDATA/Code/User"
            ;;
        *)
            log_error "Unsupported operating system"
            exit 1
            ;;
    esac
}

# Check and install uvx if needed
check_uvx() {
    log_info "Checking for uvx installation..."

    if command -v uvx >/dev/null 2>&1; then
        log_success "uvx is already installed"

        # Check for MCP clients
        if uvx list | grep -q "mcp-client-http"; then
            log_success "mcp-client-http is installed"
        else
            log_info "Installing mcp-client-http..."
            if uvx install mcp-client-http; then
                log_success "mcp-client-http installed successfully"
            else
                log_warning "Failed to install mcp-client-http, will use curl fallback"
            fi
        fi
    else
        log_warning "uvx not found. Install it for better performance:"
        log_info "  macOS: brew install uvx"
        log_info "  Linux: pip install uvx"
        log_info "Will use curl as fallback"
    fi
}

# Build the MCP Code Indexer
build_indexer() {
    log_info "Building MCP Code Indexer..."

    cd "$PROJECT_ROOT"
    if make build; then
        log_success "MCP Code Indexer built successfully"
    else
        log_error "Failed to build MCP Code Indexer"
        exit 1
    fi
}

# Create Cursor IDE configuration
setup_cursor() {
    log_info "Setting up Cursor IDE integration..."

    # Create configuration directory if it doesn't exist
    mkdir -p "$CURSOR_CONFIG_DIR"

    # Create MCP configuration
    local cursor_config="$CURSOR_CONFIG_DIR/mcp_settings.json"

    # Check if uvx is available
    if command -v uvx >/dev/null 2>&1; then
        log_info "Using uvx for Cursor IDE configuration"
        cat > "$cursor_config" << EOF
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:$SERVER_PORT/api/call",
        "--header", "Content-Type: application/json",
        "--header", "X-Session-ID: cursor-\${workspaceFolder}"
      ],
      "env": {
        "MCP_TIMEOUT": "30"
      }
    }
  }
}
EOF
    else
        log_warning "uvx not found, using curl fallback for Cursor IDE"
        cat > "$cursor_config" << EOF
{
  "mcpServers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:$SERVER_PORT/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: cursor-\${workspaceFolder}",
        "-d", "@-"
      ],
      "env": {
        "CURL_TIMEOUT": "30"
      }
    }
  }
}
EOF
    fi

    log_success "Cursor IDE configuration created at: $cursor_config"
}

# Create Augment Code configuration
setup_augment() {
    log_info "Setting up Augment Code integration..."

    # Create configuration directory if it doesn't exist
    mkdir -p "$AUGMENT_CONFIG_DIR"

    # Create MCP configuration
    local augment_config="$AUGMENT_CONFIG_DIR/mcp_settings.json"

    # Check if uvx is available
    if command -v uvx >/dev/null 2>&1; then
        log_info "Using uvx for Augment Code configuration"
        cat > "$augment_config" << EOF
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:$SERVER_PORT/api/call",
        "--header", "Content-Type: application/json",
        "--header", "X-Session-ID: augment-\${workspaceFolder}"
      ]
    }
  }
}
EOF
    else
        log_warning "uvx not found, using curl fallback for Augment Code"
        cat > "$augment_config" << EOF
{
  "mcpServers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:$SERVER_PORT/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: augment-\${workspaceFolder}",
        "-d", "@-"
      ]
    }
  }
}
EOF
    fi

    log_success "Augment Code configuration created at: $augment_config"
}

# Create workspace configuration template
create_workspace_config() {
    log_info "Creating workspace configuration template..."
    
    local workspace_dir="$PROJECT_ROOT/.mcp"
    mkdir -p "$workspace_dir"
    
    # Cursor workspace config
    if command -v uvx >/dev/null 2>&1; then
        cat > "$workspace_dir/cursor_mcp_settings.json" << EOF
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:$SERVER_PORT/api/call",
        "--header", "Content-Type: application/json",
        "--header", "X-Session-ID: cursor-\${workspaceFolder}"
      ],
      "env": {
        "MCP_TIMEOUT": "30"
      }
    }
  }
}
EOF
    else
        cat > "$workspace_dir/cursor_mcp_settings.json" << EOF
{
  "mcpServers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:$SERVER_PORT/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: cursor-\${workspaceFolder}",
        "-d", "@-"
      ],
      "env": {
        "CURL_TIMEOUT": "30"
      }
    }
  }
}
EOF
    fi
    
    # Augment workspace config
    if command -v uvx >/dev/null 2>&1; then
        cat > "$workspace_dir/augment_mcp_settings.json" << EOF
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:$SERVER_PORT/api/call",
        "--header", "Content-Type: application/json",
        "--header", "X-Session-ID: augment-\${workspaceFolder}"
      ]
    }
  }
}
EOF
    else
        cat > "$workspace_dir/augment_mcp_settings.json" << EOF
{
  "mcpServers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:$SERVER_PORT/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: augment-\${workspaceFolder}",
        "-d", "@-"
      ]
    }
  }
}
EOF
    fi
    
    log_success "Workspace configuration templates created in: $workspace_dir"
}

# Create startup script
create_startup_script() {
    log_info "Creating startup script..."
    
    local startup_script="$PROJECT_ROOT/start-mcp-daemon.sh"
    
    cat > "$startup_script" << EOF
#!/bin/bash

# MCP Code Indexer Daemon Startup Script

SCRIPT_DIR="\$(cd "\$(dirname "\${BASH_SOURCE[0]}")" && pwd)"
cd "\$SCRIPT_DIR"

echo "Starting MCP Code Indexer daemon..."
echo "Server will be available at: http://localhost:$SERVER_PORT"
echo "Health check: curl http://localhost:$SERVER_PORT/api/health"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start the daemon
./bin/code-indexer daemon --port $SERVER_PORT --host $SERVER_HOST
EOF
    
    chmod +x "$startup_script"
    log_success "Startup script created: $startup_script"
}

# Test the setup
test_setup() {
    log_info "Testing the setup..."
    
    # Start server in background for testing
    cd "$PROJECT_ROOT"
    ./bin/code-indexer daemon --port $SERVER_PORT --host $SERVER_HOST > /dev/null 2>&1 &
    local server_pid=$!
    
    # Wait for server to start
    sleep 3
    
    # Test health endpoint
    if curl -s -f "http://$SERVER_HOST:$SERVER_PORT/api/health" > /dev/null; then
        log_success "Server health check passed"
    else
        log_error "Server health check failed"
        kill $server_pid 2>/dev/null || true
        return 1
    fi
    
    # Test tool call
    local response
    response=$(curl -s -X POST "http://$SERVER_HOST:$SERVER_PORT/api/call" \
        -H "Content-Type: application/json" \
        -H "X-Session-ID: test-session" \
        -d '{"tool": "list_repositories", "arguments": {}}' 2>/dev/null)
    
    if [[ "$response" == *"repositories"* ]]; then
        log_success "Tool call test passed"
    else
        log_warning "Tool call test returned unexpected response"
    fi
    
    # Stop test server
    kill $server_pid 2>/dev/null || true
    wait $server_pid 2>/dev/null || true
    
    log_success "Setup test completed"
}

# Print usage instructions
print_instructions() {
    log_info "Setup completed! Here's how to use it:"
    echo ""
    echo "1. Start the MCP daemon:"
    echo "   ./start-mcp-daemon.sh"
    echo ""
    echo "2. In Cursor IDE:"
    echo "   - Restart Cursor IDE"
    echo "   - Open a project"
    echo "   - Try: 'Index this repository for code search'"
    echo ""
    echo "3. In Augment Code:"
    echo "   - Restart Augment Code"
    echo "   - Open a project"
    echo "   - Try: 'Search for authentication functions'"
    echo ""
    echo "4. Health check:"
    echo "   curl http://localhost:$SERVER_PORT/api/health"
    echo ""
    echo "5. Manual test:"
    echo "   curl -X POST http://localhost:$SERVER_PORT/api/call \\"
    echo "     -H 'Content-Type: application/json' \\"
    echo "     -H 'X-Session-ID: test' \\"
    echo "     -d '{\"tool\": \"list_repositories\", \"arguments\": {}}'"
    echo ""
    echo "Configuration files created:"
    echo "  - Cursor: $CURSOR_CONFIG_DIR/mcp_settings.json"
    echo "  - Augment: $AUGMENT_CONFIG_DIR/mcp_settings.json"
    echo "  - Workspace templates: $PROJECT_ROOT/.mcp/"
    echo ""
    echo "For troubleshooting, see:"
    echo "  - docs/CURSOR_MCP_SETUP.md"
    echo "  - docs/AUGMENT_CODE_SETUP.md"
    echo "  - docs/MULTI_IDE_SETUP.md"
}

# Main setup function
main() {
    log_info "MCP Code Indexer IDE Integration Setup"
    log_info "======================================"
    
    # Detect OS and set config directories
    get_config_dirs

    # Check uvx installation
    check_uvx

    # Build the indexer
    build_indexer
    
    # Setup IDE configurations
    setup_cursor
    setup_augment
    
    # Create workspace templates
    create_workspace_config
    
    # Create startup script
    create_startup_script
    
    # Test the setup
    test_setup
    
    # Print instructions
    print_instructions
    
    log_success "🎉 IDE integration setup completed successfully!"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --port)
            SERVER_PORT="$2"
            shift 2
            ;;
        --host)
            SERVER_HOST="$2"
            shift 2
            ;;
        --cursor-only)
            SETUP_CURSOR_ONLY=true
            shift
            ;;
        --augment-only)
            SETUP_AUGMENT_ONLY=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --port PORT        Server port (default: 8080)"
            echo "  --host HOST        Server host (default: localhost)"
            echo "  --cursor-only      Setup Cursor IDE only"
            echo "  --augment-only     Setup Augment Code only"
            echo "  --help             Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main setup
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
