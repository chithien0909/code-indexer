# Augment Code Integration Guide

This guide explains how to integrate the MCP Code Indexer with Augment Code IDE.

## Prerequisites

- Augment Code IDE installed
- MCP Code Indexer built and configured
- Network connectivity between Augment Code and the indexer

## Setup Methods

### Method 1: Direct uvx Execution (Recommended)

This method uses uvx to run the MCP server directly without requiring a separate daemon process.

#### 1. Install via uvx

```bash
uvx install git+https://github.com/my-mcp/code-indexer.git
```

#### 2. Configure Augment Code

Add the following to your Augment Code MCP configuration:

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

#### 1. Start the MCP Code Indexer Daemon

```bash
# Navigate to your code-indexer directory
cd /path/to/code-indexer

# Start the daemon server
./bin/code-indexer daemon --port 9991 --host localhost

# Or with custom config
./bin/code-indexer daemon --port 9991 --config config.yaml
```

#### 2. Configure Augment Code

Add the following to your Augment Code MCP configuration:

**Option A: Using uvx (Recommended)**
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
      ]
    }
  }
}
```

**Option B: Using curl (Alternative)**
```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:9991/api/call",
        "-H", "Content-Type: application/json",
        "-H", "X-Session-ID: augment-session",
        "-d", "@-"
      ]
    }
  }
}
```

**Option B: WebSocket Transport**
```json
{
  "mcpServers": {
    "code-indexer": {
      "transport": {
        "type": "websocket",
        "url": "ws://localhost:8080/ws"
      },
      "capabilities": {
        "tools": true,
        "resources": false,
        "prompts": false
      },
      "headers": {
        "X-Session-ID": "augment-${workspaceFolder}"
      },
      "timeout": 30000
    }
  }
}
```

### Method 2: Direct Process Mode

#### 1. Configure Augment Code for Direct Process

**Option A: Using uvx with direct process**
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
        "CONFIG_PATH": "/path/to/code-indexer/config.yaml"
      }
    }
  }
}
```

**Option B: Direct stdio (Traditional)**
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

### Augment Code Configuration

The MCP configuration is typically stored in one of these locations:

**macOS:**
```
~/Library/Application Support/Augment/mcp_settings.json
```

**Windows:**
```
%APPDATA%\Augment\mcp_settings.json
```

**Linux:**
```
~/.config/augment/mcp_settings.json
```

### Alternative: Workspace Configuration

You can also add the configuration to your workspace settings:

**`.augment/mcp_settings.json`** (in your project root):
```json
{
  "mcpServers": {
    "code-indexer": {
      "transport": {
        "type": "http",
        "url": "http://localhost:8080/api/call"
      },
      "capabilities": {
        "tools": true
      },
      "headers": {
        "Content-Type": "application/json",
        "X-Session-ID": "augment-${workspaceFolder}"
      }
    }
  }
}
```

## Verification

### 1. Check Server Status

```bash
# Test if the daemon is running
curl http://localhost:8080/api/health

# Expected response:
# {"status":"healthy","version":"1.1.0","multi_ide_enabled":true}
```

### 2. Test Tool Access in Augment Code

1. Open Augment Code
2. Open a project/workspace
3. Try using MCP tools through the AI assistant:

```
"Index this repository for searching"
"Search for authentication functions"
"List all indexed repositories"
```

### 3. Verify Connection

Check the server logs for connection messages:

```bash
tail -f indexer.log | grep -E "(connection|augment)"
```

## Available Tools

Once configured, you'll have access to these tools in Augment Code:

### Core Tools
- `index_repository` - Index Git repositories
- `search_code` - Search across indexed code
- `get_metadata` - Get file metadata
- `list_repositories` - List indexed repositories
- `get_index_stats` - Get indexing statistics

### Utility Tools
- `find_files` - Find files by pattern
- `find_symbols` - Find code symbols
- `get_file_content` - Get file contents
- `list_directory` - List directory contents

### Advanced Tools
- `get_file_snippet` - Extract code snippets
- `find_references` - Find symbol references
- `refresh_index` - Refresh search index
- `git_blame` - Get Git blame information

## Usage Examples

### Index Your Current Project

```
"Please index this repository so I can search through the code"
```

This will trigger the `index_repository` tool with your current workspace.

### Search for Code

```
"Find all functions related to authentication"
"Search for error handling patterns"
"Show me all database connection code"
```

### Get File Information

```
"Show me the structure of the main.go file"
"Get the content of the config file"
"List all files in the src directory"
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure the daemon is running: `./bin/code-indexer daemon --port 8080`
   - Check if port 8080 is available: `lsof -i :8080`
   - Verify firewall settings

2. **Tools Not Available**
   - Restart Augment Code after configuration changes
   - Check MCP server logs for errors
   - Verify JSON configuration syntax

3. **Slow Performance**
   - Increase timeout values in configuration
   - Check server resource usage
   - Consider using workspace isolation mode

### Debug Mode

Enable debug logging for troubleshooting:

```bash
./bin/code-indexer daemon --port 8080 --log-level debug
```

### Health Check

```bash
# Check server health
curl http://localhost:8080/api/health

# Check active sessions
curl http://localhost:8080/api/sessions

# Test tool call
curl -X POST http://localhost:8080/api/call \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: test-session" \
  -d '{"tool": "list_repositories", "arguments": {}}'
```

## Advanced Configuration

### Custom Session Management

```json
{
  "mcpServers": {
    "code-indexer": {
      "transport": {
        "type": "http",
        "url": "http://localhost:8080/api/call"
      },
      "headers": {
        "Content-Type": "application/json",
        "X-Session-ID": "augment-${workspaceFolder}-${timestamp}",
        "X-User-ID": "${username}",
        "X-Workspace": "${workspaceFolder}"
      },
      "capabilities": {
        "tools": true
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
      "transport": {
        "type": "http",
        "url": "http://localhost:8080/api/call"
      },
      "headers": {
        "X-Session-ID": "augment-main-${workspaceFolder}"
      }
    },
    "code-indexer-secondary": {
      "transport": {
        "type": "http", 
        "url": "http://localhost:8081/api/call"
      },
      "headers": {
        "X-Session-ID": "augment-secondary-${workspaceFolder}"
      }
    }
  }
}
```

## Security Considerations

### Local Development
- Use `localhost` binding for security
- Consider firewall rules for network access

### Team/Remote Setup
- Use HTTPS in production environments
- Implement authentication if needed
- Consider VPN for remote access

### Network Configuration
```bash
# Bind to specific interface
./bin/code-indexer daemon --host 127.0.0.1 --port 8080

# Allow network access (use with caution)
./bin/code-indexer daemon --host 0.0.0.0 --port 8080
```
