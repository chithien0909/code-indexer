# Migration Guide: From HTTP Daemon to Direct uvx Execution

This guide helps you migrate from the HTTP daemon-based setup to the new direct uvx execution method for the MCP Code Indexer.

## Why Migrate?

The new uvx direct execution method offers several advantages:

- **No Daemon Required**: Each IDE spawns its own isolated process
- **Simplified Setup**: Single command installation via uvx
- **Better Resource Management**: Automatic process cleanup when IDE closes
- **Easier Updates**: Simple `uvx upgrade` command
- **Cross-Platform**: Works consistently across macOS, Linux, and Windows
- **Version Management**: Easy to install and switch between versions

## Migration Overview

| Current Setup | New Setup |
|---------------|-----------|
| HTTP Daemon + curl/mcp-client-http | Direct uvx execution |
| Manual server management | Automatic process management |
| Shared daemon process | Isolated per-IDE processes |
| Port configuration required | No network configuration |
| Manual updates | Simple uvx upgrade |

## Step-by-Step Migration

### Step 1: Stop Current Daemon

If you're currently running the HTTP daemon, stop it:

```bash
# Find and stop any running daemon
pkill code-indexer

# Or if you know the specific process
ps aux | grep code-indexer
kill <PID>
```

### Step 2: Install uvx (if not already installed)

```bash
# macOS (using Homebrew)
brew install uvx

# Linux/macOS (using pip)
pip install uvx

# Windows (using pip)
pip install uvx

# Verify installation
uvx --version
```

### Step 3: Install MCP Code Indexer via uvx

```bash
# Install from GitHub repository
uvx install git+https://github.com/my-mcp/code-indexer.git

# Verify installation
uvx --from git+https://github.com/my-mcp/code-indexer.git code-indexer --version
```

### Step 4: Update IDE Configuration

#### Before (HTTP Daemon Method)

**Augment Code:**
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

**Cursor IDE:**
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
      ]
    }
  }
}
```

#### After (Direct uvx Method)

**Augment Code:**
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

**Cursor IDE:**
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

### Step 5: Update Configuration Files

#### Create Workspace Configuration

Create `.code-indexer/config.yaml` in your workspace root:

```yaml
indexer:
  supported_extensions:
    - .go
    - .py
    - .js
    - .ts
    - .java
    - .cpp
    - .c
    - .h
    - .rs
    - .rb
    - .php
    - .cs
    - .kt
    - .swift
  
  max_file_size: 10485760  # 10MB
  
  exclude_patterns:
    - "*/node_modules/*"
    - "*/vendor/*"
    - "*/.git/*"
    - "*/build/*"
    - "*/dist/*"
    - "*/target/*"
    - "*/__pycache__/*"
  
  index_dir: ".code-indexer/index"
  repo_dir: ".code-indexer/repositories"

search:
  max_results: 100
  highlight_snippets: true
  snippet_length: 200
  fuzzy_tolerance: 0.2

logging:
  level: info
  file: ".code-indexer/indexer.log"
  json_format: false

models:
  enabled: true
  default_model: "code-assistant-v1"
  models_dir: ".code-indexer/models"
  max_tokens: 2048
  temperature: 0.7
```

#### Migrate Existing Configuration

If you have an existing global configuration, you can:

1. **Copy to workspace:**
   ```bash
   mkdir -p .code-indexer
   cp /path/to/old/config.yaml .code-indexer/config.yaml
   ```

2. **Update paths to be relative:**
   ```yaml
   # Change absolute paths like:
   index_dir: "/home/user/indexer/index"
   repo_dir: "/home/user/indexer/repositories"
   
   # To relative paths:
   index_dir: ".code-indexer/index"
   repo_dir: ".code-indexer/repositories"
   ```

### Step 6: Test the New Setup

1. **Restart your IDE** to load the new configuration

2. **Test MCP tools:**
   - "Index this repository"
   - "Search for authentication functions"
   - "List all indexed repositories"

3. **Verify process isolation:**
   ```bash
   # Open multiple IDE instances and check processes
   ps aux | grep code-indexer
   # You should see separate processes for each IDE
   ```

### Step 7: Clean Up Old Setup

Once you've verified the new setup works:

1. **Remove old daemon startup scripts:**
   ```bash
   rm -f start-mcp-daemon.sh
   rm -f /path/to/old/startup/scripts
   ```

2. **Remove old configuration files:**
   ```bash
   # Remove global config if no longer needed
   rm -f ~/.config/code-indexer/config.yaml
   ```

3. **Uninstall old HTTP client tools (if installed):**
   ```bash
   uvx uninstall mcp-client-http
   ```

## Troubleshooting Migration

### Common Issues

1. **uvx command not found:**
   ```bash
   # Ensure uvx is in PATH
   which uvx
   
   # Reinstall if needed
   pip install --upgrade uvx
   ```

2. **Installation fails:**
   ```bash
   # Check Go installation
   go version
   
   # Install Go if missing
   # macOS: brew install go
   # Linux: sudo apt install golang-go
   ```

3. **IDE can't find uvx:**
   ```bash
   # Check IDE's PATH
   echo $PATH
   
   # Add uvx location to IDE environment
   export PATH="$HOME/.local/bin:$PATH"
   ```

4. **Configuration not found:**
   ```bash
   # Verify config file location
   ls -la .code-indexer/config.yaml
   
   # Check file permissions
   chmod 644 .code-indexer/config.yaml
   ```

### Rollback Plan

If you need to rollback to the HTTP daemon method:

1. **Start the daemon:**
   ```bash
   ./bin/code-indexer daemon --port 9991
   ```

2. **Restore old IDE configuration** (see "Before" examples above)

3. **Restart your IDE**

### Performance Comparison

| Aspect | HTTP Daemon | Direct uvx |
|--------|-------------|------------|
| Startup Time | Fast (daemon already running) | Slightly slower (process spawn) |
| Memory Usage | Shared (one process) | Higher (multiple processes) |
| Resource Isolation | Limited | Complete |
| Update Process | Manual restart | Automatic on next spawn |
| Debugging | Centralized logs | Per-process logs |
| Scalability | Limited by daemon | Scales with IDE instances |

## Best Practices for uvx Setup

1. **Use workspace-specific configuration:**
   - Place config in `.code-indexer/config.yaml`
   - Use relative paths for better portability

2. **Version pinning for teams:**
   ```json
   {
     "command": "uvx",
     "args": [
       "--from",
       "git+https://github.com/my-mcp/code-indexer.git@v1.1.0",
       "code-indexer",
       "mcp-server"
     ]
   }
   ```

3. **Environment variables for customization:**
   ```json
   {
     "env": {
       "CONFIG_PATH": "${workspaceFolder}/.code-indexer/config.yaml",
       "LOG_LEVEL": "info",
       "INDEX_DIR": "${workspaceFolder}/.code-indexer/index"
     }
   }
   ```

4. **Regular updates:**
   ```bash
   # Update to latest version
   uvx upgrade mcp-code-indexer
   
   # Or reinstall from git
   uvx uninstall mcp-code-indexer
   uvx install git+https://github.com/my-mcp/code-indexer.git
   ```

## Support

If you encounter issues during migration:

1. Check the [UVX Installation Guide](UVX_INSTALLATION.md)
2. Review the [troubleshooting section](UVX_INSTALLATION.md#troubleshooting)
3. Open an issue on GitHub with migration details

The uvx direct execution method provides a more robust and maintainable setup for the MCP Code Indexer while eliminating the complexity of daemon management.
