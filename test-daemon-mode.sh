#!/bin/bash

echo "🧪 Testing Daemon Mode..."

# Test daemon mode which we know works
echo "1. Testing daemon mode..."

# Start daemon in background
uvx --from git+https://github.com/chithien0909/code-indexer.git code-indexer daemon --port 9991 &
DAEMON_PID=$!

# Wait for daemon to start
sleep 5

# Test if daemon is responding
echo "2. Testing daemon health..."
HEALTH_RESPONSE=$(curl -s http://localhost:9991/api/health)

if [[ -n "$HEALTH_RESPONSE" ]]; then
    echo "✅ Daemon is responding"
    echo "Health: $HEALTH_RESPONSE"
else
    echo "❌ Daemon not responding"
fi

# Test tools API
echo "3. Testing tools API..."
TOOLS_RESPONSE=$(curl -s http://localhost:9991/api/tools)

if [[ -n "$TOOLS_RESPONSE" ]] && echo "$TOOLS_RESPONSE" | grep -q "index_repository"; then
    echo "✅ Tools API working"
    echo "Tools: $TOOLS_RESPONSE"
else
    echo "❌ Tools API not working"
    echo "Response: $TOOLS_RESPONSE"
fi

# Clean up
echo "4. Cleaning up..."
kill $DAEMON_PID 2>/dev/null
sleep 2

echo "🏁 Daemon test complete!"
