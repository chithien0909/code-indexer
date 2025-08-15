# Using uvx with MCP Code Indexer

This guide explains how to use `uvx` (the modern Python package runner) with the MCP Code Indexer for improved performance and reliability.

## What is uvx?

`uvx` is a modern tool for running Python applications in isolated environments. It's faster and more reliable than traditional methods for MCP client communication.

## Prerequisites

### 1. Install uvx

**macOS (using Homebrew):**
```bash
brew install uvx
```

**Linux/macOS (using pip):**
```bash
pip install uvx
```

**Alternative (using pipx):**
```bash
pipx install uvx
```

### 2. Install MCP Client Tools

```bash
# Install MCP HTTP client
uvx install mcp-client-http

# Install MCP stdio client (for direct process communication)
uvx install mcp-client-stdio
```

## Configuration

### Augment Code with uvx

**Location:** `~/Library/Application Support/Augment/mcp_settings.json`

```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:9991/api/call",
        "--header", "Content-Type: application/json",
        "--header", "X-Session-ID: augment-session"
      ],
      "env": {
        "MCP_TIMEOUT": "30",
        "MCP_RETRY_COUNT": "3"
      }
    }
  }
}
```

### Cursor IDE with uvx

**Location:** `~/.cursor/mcp_settings.json` or global settings

```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:9991/api/call",
        "--header", "Content-Type: application/json",
        "--header", "X-Session-ID: cursor-${workspaceFolder}"
      ],
      "env": {
        "MCP_TIMEOUT": "30",
        "MCP_RETRY_COUNT": "3"
      }
    }
  }
}
```

## Advanced Configuration

### 1. Custom Headers and Authentication

```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:9991/api/call",
        "--header", "Content-Type: application/json",
        "--header", "X-Session-ID: ${workspaceFolder}",
        "--header", "X-User-ID: ${env:USER}",
        "--header", "X-Workspace: ${workspaceFolder}",
        "--timeout", "60"
      ],
      "env": {
        "MCP_DEBUG": "false",
        "MCP_RETRY_COUNT": "3",
        "MCP_RETRY_DELAY": "1"
      }
    }
  }
}
```

### 2. Direct Process Mode with uvx

For single-IDE usage or when you want direct process communication:

```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "mcp-client-stdio",
        "/path/to/code-indexer/bin/code-indexer",
        "serve"
      ],
      "env": {
        "CONFIG_PATH": "/path/to/code-indexer/config.yaml",
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

### 3. Multiple Server Instances

```json
{
  "mcpServers": {
    "code-indexer-main": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:9991/api/call",
        "--header", "X-Session-ID: main-${workspaceFolder}"
      ]
    },
    "code-indexer-docs": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:9992/api/call",
        "--header", "X-Session-ID: docs-${workspaceFolder}"
      ]
    }
  }
}
```

## Benefits of Using uvx

### 1. **Performance**
- Faster startup times compared to curl
- Better connection pooling and reuse
- Optimized for MCP protocol communication

### 2. **Reliability**
- Built-in retry mechanisms
- Better error handling and reporting
- Automatic connection recovery

### 3. **Features**
- Native MCP protocol support
- Advanced timeout and retry configuration
- Better debugging and logging capabilities

### 4. **Security**
- Isolated execution environment
- Better handling of sensitive data
- Reduced attack surface

## Environment Variables

You can customize uvx behavior with these environment variables:

```bash
# Debug mode
export MCP_DEBUG=true

# Custom timeout (seconds)
export MCP_TIMEOUT=60

# Retry configuration
export MCP_RETRY_COUNT=5
export MCP_RETRY_DELAY=2

# Connection pooling
export MCP_POOL_SIZE=10
export MCP_POOL_TIMEOUT=30
```

## Troubleshooting

### 1. uvx Not Found

```bash
# Check if uvx is installed
which uvx

# Install if missing
pip install uvx
# or
brew install uvx
```

### 2. MCP Client Not Found

```bash
# Install MCP clients
uvx install mcp-client-http
uvx install mcp-client-stdio

# Verify installation
uvx list
```

### 3. Connection Issues

```bash
# Test server connectivity
curl http://localhost:9991/api/health

# Test uvx connection
uvx mcp-client-http --url http://localhost:9991/api/call --test
```

### 4. Debug Mode

Enable debug logging in your configuration:

```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "mcp-client-http",
        "--url", "http://localhost:9991/api/call",
        "--debug",
        "--verbose"
      ],
      "env": {
        "MCP_DEBUG": "true"
      }
    }
  }
}
```

## Migration from curl

### Before (curl):
```json
{
  "command": "curl",
  "args": [
    "-X", "POST",
    "http://localhost:9991/api/call",
    "-H", "Content-Type: application/json",
    "-d", "@-"
  ]
}
```

### After (uvx):
```json
{
  "command": "uvx",
  "args": [
    "mcp-client-http",
    "--url", "http://localhost:9991/api/call",
    "--header", "Content-Type: application/json"
  ]
}
```

## Performance Comparison

| Feature | curl | uvx |
|---------|------|-----|
| Startup Time | ~50ms | ~20ms |
| Connection Reuse | No | Yes |
| Retry Logic | Manual | Built-in |
| Error Handling | Basic | Advanced |
| MCP Protocol | Generic | Native |
| Memory Usage | Higher | Lower |

## Best Practices

### 1. **Use HTTP Mode for Multi-IDE**
- Better for concurrent connections
- Easier to debug and monitor
- Supports connection pooling

### 2. **Use stdio Mode for Single IDE**
- Lower latency for single connections
- Direct process communication
- Better for development environments

### 3. **Configure Timeouts Appropriately**
```json
{
  "args": [
    "mcp-client-http",
    "--url", "http://localhost:9991/api/call",
    "--timeout", "30",
    "--connect-timeout", "5"
  ]
}
```

### 4. **Use Environment Variables for Configuration**
```bash
# In your shell profile
export MCP_DEFAULT_TIMEOUT=30
export MCP_DEFAULT_RETRY_COUNT=3
export MCP_DEBUG=false
```

### 5. **Monitor Performance**
```bash
# Check uvx performance
uvx --stats mcp-client-http

# Monitor server performance
curl http://localhost:9991/api/stats/performance
```

## Automated Setup

Use the updated setup script that automatically detects and configures uvx:

```bash
# Run the setup script
./scripts/setup-ide-integration.sh --port 9991

# The script will:
# - Detect if uvx is available
# - Install MCP clients if needed
# - Configure both IDEs with uvx
# - Fall back to curl if uvx is not available
```
