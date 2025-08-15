#!/bin/bash

echo "=== MCP Code Indexer Startup Test ==="
echo "Date: $(date)"
echo "Working Directory: $(pwd)"
echo "User: $(whoami)"
echo ""

echo "1. Testing binary existence and permissions:"
if [ -f "/home/hp/Documents/personal/my-mcp/bin/code-indexer" ]; then
    echo "✅ Binary exists"
    ls -la /home/hp/Documents/personal/my-mcp/bin/code-indexer
else
    echo "❌ Binary not found"
    exit 1
fi

echo ""
echo "2. Testing binary execution:"
if /home/hp/Documents/personal/my-mcp/bin/code-indexer --help > /dev/null 2>&1; then
    echo "✅ Binary executes successfully"
else
    echo "❌ Binary execution failed"
    exit 1
fi

echo ""
echo "3. Testing MCP server startup (3 second test):"
echo "Starting server..."
timeout 3s /home/hp/Documents/personal/my-mcp/bin/code-indexer serve 2>&1 | head -20

echo ""
echo "4. Configuration file check:"
if [ -f "/home/hp/Documents/personal/my-mcp/config.yaml" ]; then
    echo "✅ Config file exists"
    echo "Config file contents:"
    cat /home/hp/Documents/personal/my-mcp/config.yaml
else
    echo "⚠️  No config file found (using defaults)"
fi

echo ""
echo "5. Environment check:"
echo "PATH: $PATH"
echo "HOME: $HOME"
echo "PWD: $PWD"

echo ""
echo "6. Process check:"
echo "Current MCP processes:"
ps aux | grep code-indexer | grep -v grep || echo "No MCP processes running"

echo ""
echo "=== Test Complete ==="
echo ""
echo "If you see this output, copy it and share with the developer for diagnosis."
