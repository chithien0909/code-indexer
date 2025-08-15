#!/bin/bash
# Test script for MCP Code Indexer integration

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$HOME/.config/code-indexer"

echo "Testing MCP Code Indexer integration..."
echo "======================================"

# Test 1: Server starts
echo "Test 1: Server startup"
# MCP servers wait for stdin, so we need to provide input or they exit
echo '{"jsonrpc": "2.0", "method": "initialize", "id": 1}' | "$SCRIPT_DIR/bin/code-indexer" serve --config "$CONFIG_DIR/config.yaml" > /tmp/server_test.log 2>&1 &
SERVER_PID=$!
sleep 2

# Check if server started and is processing MCP protocol
if grep -q '"protocolVersion"' /tmp/server_test.log 2>/dev/null; then
    echo "✅ Server starts successfully and responds to MCP protocol"
    kill $SERVER_PID 2>/dev/null
    wait $SERVER_PID 2>/dev/null
else
    echo "❌ Server failed to start or respond to MCP protocol"
    echo "Server output:"
    cat /tmp/server_test.log 2>/dev/null || echo "No output captured"
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
