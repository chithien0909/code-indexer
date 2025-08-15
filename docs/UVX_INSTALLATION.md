# UVX Installation Guide

This guide explains how to install and use the MCP Code Indexer directly via uvx, eliminating the need for a separate HTTP daemon.

## What is uvx?

`uvx` is a modern Python package runner that allows you to install and run Python applications in isolated environments. With uvx, you can install the MCP Code Indexer directly from the Git repository and use it seamlessly with your IDE.

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

**Windows (using pip):**
```bash
pip install uvx
```

**Alternative (using pipx):**
```bash
pipx install uvx
```

### 2. Verify uvx Installation

```bash
uvx --version
```

## Installation Methods

### Method 1: Install from Git Repository (Recommended)

Install directly from the GitHub repository:

```bash
uvx install git+https://github.com/my-mcp/code-indexer.git
```

### Method 2: Install from Local Directory

If you have the source code locally:

```bash
# Clone the repository
git clone https://github.com/my-mcp/code-indexer.git
cd code-indexer

# Install from local directory
uvx install .
```

### Method 3: Install Specific Version/Branch

Install a specific version or branch:

```bash
# Install specific version
uvx install git+https://github.com/my-mcp/code-indexer.git@v1.1.0

# Install from specific branch
uvx install git+https://github.com/my-mcp/code-indexer.git@main
```

## IDE Configuration

### Augment Code Configuration

Add this to your Augment Code MCP settings:

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

### Cursor IDE Configuration

Add this to your Cursor IDE MCP settings:

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

### VS Code Configuration

Add this to your VS Code MCP settings:

```json
{
  "mcp.servers": {
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

## Configuration

### Workspace Configuration

Create a `.code-indexer/config.yaml` file in your workspace root:

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

### Global Configuration

You can also create a global configuration file:

**macOS/Linux:**
```bash
mkdir -p ~/.config/code-indexer
cp config.yaml ~/.config/code-indexer/
```

**Windows:**
```cmd
mkdir %APPDATA%\code-indexer
copy config.yaml %APPDATA%\code-indexer\
```

## Usage Examples

### Basic Usage

Once configured, you can use these commands in your IDE:

```
"Index this repository for code search"
"Search for authentication functions"
"Find all error handling patterns"
"Show me the main configuration files"
"List all Go files in this project"
```

### Manual Testing

You can also test the installation manually:

```bash
# Test the installation
uvx --from git+https://github.com/my-mcp/code-indexer.git code-indexer --version

# Run the MCP server directly
uvx --from git+https://github.com/my-mcp/code-indexer.git code-indexer mcp-server

# Run with custom config
uvx --from git+https://github.com/my-mcp/code-indexer.git code-indexer mcp-server --config ./my-config.yaml
```

## Advanced Configuration

### Custom Installation Name

Install with a custom name:

```bash
uvx install --name my-code-indexer git+https://github.com/my-mcp/code-indexer.git
```

Then use in IDE configuration:

```json
{
  "command": "my-code-indexer",
  "args": ["mcp-server"]
}
```

### Development Installation

For development, install in editable mode:

```bash
git clone https://github.com/my-mcp/code-indexer.git
cd code-indexer
uvx install --editable .
```

### Multiple Versions

Install multiple versions side by side:

```bash
uvx install --name code-indexer-v1 git+https://github.com/my-mcp/code-indexer.git@v1.0.0
uvx install --name code-indexer-v2 git+https://github.com/my-mcp/code-indexer.git@v1.1.0
```

## Troubleshooting

### Installation Issues

1. **Go compiler not found:**
   ```bash
   # Install Go
   # macOS
   brew install go
   
   # Linux (Ubuntu/Debian)
   sudo apt install golang-go
   
   # Windows
   # Download from https://golang.org/dl/
   ```

2. **Permission errors:**
   ```bash
   # Use --user flag
   uvx install --user git+https://github.com/my-mcp/code-indexer.git
   ```

3. **Network issues:**
   ```bash
   # Use SSH instead of HTTPS
   uvx install git+ssh://git@github.com/my-mcp/code-indexer.git
   ```

### Runtime Issues

1. **Binary not found:**
   - Check that Go is installed and in PATH
   - Verify the installation completed successfully
   - Try reinstalling: `uvx uninstall mcp-code-indexer && uvx install git+https://github.com/my-mcp/code-indexer.git`

2. **Configuration not found:**
   - Ensure config file exists in the expected location
   - Check file permissions
   - Use absolute paths in configuration

3. **IDE connection issues:**
   - Verify the MCP server starts manually
   - Check IDE logs for error messages
   - Ensure uvx is in the IDE's PATH

### Debug Mode

Enable debug logging:

```json
{
  "command": "uvx",
  "args": [
    "--from",
    "git+https://github.com/my-mcp/code-indexer.git",
    "code-indexer",
    "mcp-server",
    "--log-level", "debug"
  ]
}
```

## Benefits of uvx Installation

1. **No Daemon Required**: Each IDE spawns its own process
2. **Automatic Updates**: Easy to update with `uvx upgrade`
3. **Isolation**: Each installation is isolated
4. **Cross-Platform**: Works on macOS, Linux, and Windows
5. **Version Management**: Easy to install multiple versions
6. **No System Dependencies**: uvx handles all dependencies

## Migration from HTTP Daemon

If you're currently using the HTTP daemon mode, here's how to migrate:

1. **Stop the daemon:**
   ```bash
   # Stop any running daemon
   pkill code-indexer
   ```

2. **Install via uvx:**
   ```bash
   uvx install git+https://github.com/my-mcp/code-indexer.git
   ```

3. **Update IDE configuration:**
   Replace HTTP-based configuration with uvx-based configuration (see examples above)

4. **Test the new setup:**
   Restart your IDE and test the MCP tools

The uvx installation provides the same functionality as the daemon mode but with better process isolation and easier management.
