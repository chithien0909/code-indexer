# Cursor IDE MCP Integration Guide

This guide explains how to integrate the MCP Code Indexer with Cursor IDE for enhanced code search and analysis capabilities.

## Prerequisites

- Cursor IDE installed (latest version recommended)
- MCP Code Indexer built and configured
- Basic understanding of JSON configuration

## Quick Setup

### Method 1: Direct uvx Execution (Recommended)

#### 1. Install via uvx

```bash
uvx install git+https://github.com/my-mcp/code-indexer.git
```

#### 2. Configure Cursor IDE

Add the following to your Cursor IDE MCP configuration:

```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "--from",
        "git+https://github.com/my-mcp/code-indexer.git",
        "code-indexer",
        "mcp-server"
      ],
      "env": {
        "CONFIG_PATH": "${workspaceFolder}/.code-indexer/config.yaml"
      }
    }
  }
}
```

### Method 2: Daemon Mode (For Multi-IDE)

#### 1. Start the MCP Code Indexer

```bash
# Navigate to your code-indexer directory
cd /path/to/code-indexer

# Start the daemon server
./bin/code-indexer daemon --port 8080 --host localhost

# Verify it's running
curl http://localhost:8080/api/health
```

#### 2. Configure Cursor IDE

#### Method A: Global Configuration (Recommended)

1. Open Cursor IDE
2. Go to **Settings** → **Extensions** → **MCP**
3. Add the following configuration:

**Option 1: Using uvx (Recommended)**
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
        "MCP_TIMEOUT": "30"
      }
    }
  }
}
```

**Option 2: Using curl (Alternative)**
```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:9991/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: cursor-${workspaceFolder}",
        "-d", "@-"
      ],
      "env": {
        "CURL_TIMEOUT": "30"
      }
    }
  }
}
```

#### Method B: Workspace Configuration

Create `.cursor/mcp_settings.json` in your project root:

**Using uvx (Recommended):**
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
        "MCP_TIMEOUT": "30"
      }
    }
  }
}
```

#### Method C: Direct Process Mode

For single-IDE usage, you can run the indexer directly:

```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "/path/to/code-indexer/bin/code-indexer",
      "args": ["serve"],
      "env": {
        "CONFIG_PATH": "/path/to/code-indexer/config.yaml"
      }
    }
  }
}
```

## Configuration File Locations

### Cursor Settings

The MCP configuration can be placed in:

**Global Settings:**
- **macOS**: `~/Library/Application Support/Cursor/User/settings.json`
- **Windows**: `%APPDATA%\Cursor\User\settings.json`
- **Linux**: `~/.config/Cursor/User/settings.json`

**Workspace Settings:**
- `.cursor/mcp_settings.json` (in project root)
- `.vscode/settings.json` (if Cursor uses VS Code settings)

### Example Global Settings

Add this to your Cursor `settings.json`:

```json
{
  "mcp.servers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:8080/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: cursor-${workspaceFolder}",
        "-d", "@-"
      ],
      "env": {
        "CURL_TIMEOUT": "30"
      }
    }
  }
}
```

## Verification Steps

### 1. Test Server Connection

```bash
# Check if daemon is running
curl http://localhost:8080/api/health

# Test a tool call
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: cursor-test" \
  -d '{
    "tool": "list_repositories",
    "arguments": {}
  }'
```

### 2. Test in Cursor IDE

1. Open Cursor IDE
2. Open a project/workspace
3. Open the AI chat panel
4. Try these commands:

```
"Index this repository for code search"
"Search for all authentication functions"
"Show me the main configuration files"
```

### 3. Check MCP Tools

In Cursor's AI chat, you can ask:

```
"What MCP tools are available?"
"List all code indexer capabilities"
```

## Available Tools in Cursor

Once configured, you'll have access to these tools:

### Core Indexing Tools
- **Index Repository**: `"Index this repository"`
- **Search Code**: `"Search for [pattern] in the code"`
- **Get Metadata**: `"Show me details about [file]"`
- **List Repositories**: `"What repositories are indexed?"`

### File Operations
- **Find Files**: `"Find all .go files"`
- **Get File Content**: `"Show me the content of main.go"`
- **List Directory**: `"List files in the src directory"`

### Advanced Search
- **Find Symbols**: `"Find all functions named 'authenticate'"`
- **Find References**: `"Find all references to this function"`
- **Get Snippets**: `"Show me lines 10-20 of config.go"`

## Usage Examples

### Initial Setup

```
User: "Index this repository so I can search through the code"
```

This will trigger the `index_repository` tool and index your current workspace.

### Code Search

```
User: "Find all error handling patterns in this codebase"
User: "Search for database connection code"
User: "Show me all API endpoint definitions"
```

### File Exploration

```
User: "What's in the main configuration file?"
User: "List all Go files in this project"
User: "Show me the structure of the internal directory"
```

### Code Analysis

```
User: "Find all functions that handle user authentication"
User: "Show me where the 'UserService' struct is defined"
User: "Find all references to the 'config' variable"
```

## Advanced Configuration

### Custom Session Management

```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:8080/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: cursor-${workspaceFolder}-${timestamp}",
        "-H", "X-User-ID: ${env:USER}",
        "-d", "@-"
      ],
      "env": {
        "CURL_TIMEOUT": "60"
      }
    }
  }
}
```

### Multiple Indexer Instances

```json
{
  "mcpServers": {
    "code-indexer-main": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:8080/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: cursor-main-${workspaceFolder}",
        "-d", "@-"
      ]
    },
    "code-indexer-docs": {
      "command": "curl", 
      "args": [
        "-X", "POST",
        "http://localhost:8081/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: cursor-docs-${workspaceFolder}",
        "-d", "@-"
      ]
    }
  }
}
```

## Troubleshooting

### Common Issues

1. **MCP Server Not Found**
   - Restart Cursor after configuration changes
   - Check configuration file syntax
   - Verify file paths and permissions

2. **Connection Timeout**
   - Increase `CURL_TIMEOUT` value
   - Check if daemon is running: `curl http://localhost:8080/api/health`
   - Verify network connectivity

3. **Tools Not Working**
   - Check server logs: `tail -f indexer.log`
   - Test tool calls manually with curl
   - Verify JSON request format

### Debug Mode

Enable debug logging:

```bash
./bin/code-indexer daemon --port 8080 --log-level debug
```

Monitor logs:

```bash
tail -f indexer.log | grep -E "(cursor|connection|error)"
```

### Health Checks

```bash
# Server health
curl http://localhost:8080/api/health

# Active sessions
curl http://localhost:8080/api/sessions

# Tool availability
curl -X POST http://localhost:8080/api/tools
```

## Performance Optimization

### For Large Codebases

```yaml
# config.yaml optimizations
search:
  max_results: 50
  snippet_length: 150
  
server:
  multi_ide:
    max_connections: 20
    connection_timeout_seconds: 120
    
    resource_management:
      max_concurrent_operations: 5
      operation_timeout_minutes: 3
```

### Memory Usage

```bash
# Monitor memory usage
ps aux | grep code-indexer

# Check index size
du -sh ./index/
```

## Security Notes

### Local Development
- Use `localhost` binding for security
- Default port 8080 is usually safe for local use

### Network Access
```bash
# Secure local binding
./bin/code-indexer daemon --host 127.0.0.1 --port 8080

# For team access (use with caution)
./bin/code-indexer daemon --host 0.0.0.0 --port 8080
```

### Firewall Configuration
```bash
# Allow local access only
sudo ufw allow from 127.0.0.1 to any port 8080

# Allow team network access
sudo ufw allow from 192.168.1.0/24 to any port 8080
```
