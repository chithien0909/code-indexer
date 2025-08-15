#!/bin/bash

echo "Testing MCP Code Indexer API..."

# Test health endpoint
echo "1. Testing health endpoint:"
curl -s http://localhost:8080/api/health | jq .

echo -e "\n2. Testing tools endpoint:"
curl -s http://localhost:8080/api/tools | jq '.total, .server_info'

echo -e "\n3. Testing sessions endpoint:"
curl -s http://localhost:8080/api/sessions | jq .

echo -e "\n4. Testing simple tool call (list_sessions):"
curl -v -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{"tool": "list_sessions", "arguments": {}}'

echo -e "\n5. Testing get_session_info tool call:"
curl -v -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -d '{"tool": "get_session_info", "arguments": {}}'
