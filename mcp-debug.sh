#!/bin/bash

# MCP Code Indexer Debug Wrapper for Cursor IDE
# This script logs everything for debugging

LOG_FILE="/tmp/mcp-debug-$(date +%Y%m%d-%H%M%S).log"

{
    echo "=== MCP Debug Log Started at $(date) ==="
    echo "Working Directory: $(pwd)"
    echo "User: $(whoami)"
    echo "Arguments: $@"
    echo "Environment:"
    env | grep -E "(PATH|HOME|USER|PWD)" | sort
    echo ""
    
    # Change to MCP directory
    echo "Changing to MCP directory..."
    cd /home/hp/Documents/personal/my-mcp
    echo "New working directory: $(pwd)"
    
    # Check binary
    echo "Checking binary..."
    if [ -f "/home/hp/Documents/personal/my-mcp/bin/code-indexer" ]; then
        echo "✅ Binary exists"
        ls -la /home/hp/Documents/personal/my-mcp/bin/code-indexer
    else
        echo "❌ Binary not found"
        exit 1
    fi
    
    echo ""
    echo "Starting MCP server..."
    echo "Command: /home/hp/Documents/personal/my-mcp/bin/code-indexer $@"
    echo ""
    
} >> "$LOG_FILE" 2>&1

# Execute the MCP server and log output
exec /home/hp/Documents/personal/my-mcp/bin/code-indexer "$@" 2>> "$LOG_FILE"
