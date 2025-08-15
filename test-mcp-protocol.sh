#!/bin/bash

echo "🧪 Testing MCP Protocol Communication..."

# Test 1: Basic server startup
echo "1. Testing server startup..."
timeout 5s uvx --from git+https://github.com/chithien0909/code-indexer.git code-indexer serve --log-level debug &
SERVER_PID=$!
sleep 2

if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✅ Server started successfully"
    kill $SERVER_PID 2>/dev/null
else
    echo "❌ Server failed to start"
fi

# Test 2: MCP Initialize
echo ""
echo "2. Testing MCP initialize..."
INIT_RESPONSE=$(echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}' | timeout 10s uvx --from git+https://github.com/chithien0909/code-indexer.git code-indexer serve 2>/dev/null)

if [[ -n "$INIT_RESPONSE" ]]; then
    echo "✅ Server responded to initialize"
    echo "Response: $INIT_RESPONSE"
else
    echo "❌ No response to initialize"
fi

# Test 3: Tools List
echo ""
echo "3. Testing tools/list..."
TOOLS_RESPONSE=$(
(
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}'
sleep 1
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}'
) | timeout 15s uvx --from git+https://github.com/chithien0909/code-indexer.git code-indexer serve 2>/dev/null
)

if [[ -n "$TOOLS_RESPONSE" ]] && echo "$TOOLS_RESPONSE" | grep -q "tools"; then
    echo "✅ Tools list available"
    echo "Response: $TOOLS_RESPONSE"
else
    echo "❌ No tools in response"
    echo "Response: $TOOLS_RESPONSE"
fi

# Test 4: Check for specific tools
echo ""
echo "4. Checking for specific tools..."
if echo "$TOOLS_RESPONSE" | grep -q "index_repository"; then
    echo "✅ index_repository tool found"
else
    echo "❌ index_repository tool missing"
fi

if echo "$TOOLS_RESPONSE" | grep -q "search_code"; then
    echo "✅ search_code tool found"
else
    echo "❌ search_code tool missing"
fi

echo ""
echo "🏁 Test complete!"
