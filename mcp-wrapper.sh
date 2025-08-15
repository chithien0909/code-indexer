#!/bin/bash

# MCP Code Indexer Wrapper Script for Cursor IDE
# This script ensures proper environment and working directory

# Set working directory to the MCP project directory
cd /home/hp/Documents/personal/my-mcp

# Log startup for debugging (optional)
echo "$(date): Starting MCP Code Indexer from $(pwd)" >> /tmp/mcp-startup.log
echo "$(date): Args: $@" >> /tmp/mcp-startup.log

# Execute the MCP server with all arguments passed through
exec /home/hp/Documents/personal/my-mcp/bin/code-indexer "$@"
