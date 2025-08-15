# MCP Code Indexer Integration with Cursor/Augment IDE

This guide provides step-by-step instructions for integrating the MCP Code Indexer server with Cursor or Augment IDE to enable powerful code search and analysis capabilities.

## Prerequisites

- Cursor IDE or Augment IDE installed
- MCP Code Indexer built and ready (`make build`)
- Basic familiarity with JSON configuration files

## 1. MCP Configuration Setup

### For Cursor IDE

1. **Locate the MCP configuration file:**
   ```bash
   # On macOS
   ~/Library/Application Support/Cursor/User/globalStorage/rooveterinaryinc.roo-cline/settings/cline_mcp_settings.json
   
   # On Linux
   ~/.config/Cursor/User/globalStorage/rooveterinaryinc.roo-cline/settings/cline_mcp_settings.json
   
   # On Windows
   %APPDATA%\Cursor\User\globalStorage\rooveterinaryinc.roo-cline\settings\cline_mcp_settings.json
   ```

2. **Create or update the MCP configuration:**
   ```json
   {
     "mcpServers": {
       "code-indexer": {
         "command": "/path/to/your/my-mcp/bin/code-indexer",
         "args": ["serve"],
         "env": {
           "INDEXER_LOG_LEVEL": "info"
         }
       }
     }
   }
   ```

### For Augment IDE

1. **Open Augment settings:**
   - Go to `Settings` → `Extensions` → `MCP Servers`
   - Or edit the configuration file directly

2. **Add the MCP Code Indexer configuration:**
   ```json
   {
     "mcp": {
       "servers": {
         "code-indexer": {
           "command": "/path/to/your/my-mcp/bin/code-indexer",
           "args": ["serve"],
           "cwd": "/path/to/your/my-mcp",
           "env": {
             "INDEXER_LOG_LEVEL": "info"
           }
         }
       }
     }
   }
   ```

### Configuration Options

You can customize the server behavior with additional arguments:

```json
{
  "command": "/path/to/your/my-mcp/bin/code-indexer",
  "args": [
    "serve",
    "--config", "/path/to/custom/config.yaml",
    "--log-level", "debug"
  ],
  "env": {
    "INDEXER_INDEX_DIR": "/custom/index/path",
    "INDEXER_REPO_DIR": "/custom/repos/path"
  }
}
```

## 2. Server Integration

### Step 1: Verify Server Installation

```bash
# Test the server can start
cd /path/to/your/my-mcp
./bin/code-indexer --help

# Test with example data
make test-example
```

### Step 2: Start the IDE with MCP Support

1. **Restart Cursor/Augment** after updating the configuration
2. **Verify connection** in the IDE:
   - Look for MCP server status in the status bar
   - Check the output/logs panel for connection messages
   - You should see "Connected to MCP server: code-indexer"

### Step 3: Verify Tool Registration

Open the command palette and look for MCP tools:
- `MCP: List Available Tools`
- You should see all 5 Code Indexer tools listed

## 3. Tool Registration and Availability

Once connected, the following tools will be available:

### Available MCP Tools

| Tool Name | Description | Parameters |
|-----------|-------------|------------|
| `index_repository` | Index a Git repository | `path` (required), `name` (optional) |
| `search_code` | Search across indexed code | `query` (required), `type`, `language`, `repository`, `max_results` |
| `get_metadata` | Get file metadata | `file_path` (required), `repository` (optional) |
| `list_repositories` | List indexed repositories | None |
| `get_index_stats` | Get indexing statistics | None |

### Verifying Tool Registration

1. **Open Command Palette** (`Cmd/Ctrl + Shift + P`)
2. **Type "MCP"** to see available MCP commands
3. **Select "MCP: List Available Tools"**
4. **Confirm all 5 tools are listed**

## 4. Usage Examples

### Example 1: Index a Repository

**Prompt to AI Assistant:**
```
Please index my current project repository using the MCP Code Indexer.
```

**Expected AI Action:**
The AI will use the `index_repository` tool:
```json
{
  "tool": "index_repository",
  "arguments": {
    "path": "/path/to/current/project",
    "name": "my-project"
  }
}
```

**Expected Response:**
```json
{
  "success": true,
  "repository": {
    "id": "a1b2c3d4e5f6g7h8",
    "name": "my-project",
    "file_count": 150,
    "total_lines": 12500,
    "languages": ["go", "python", "javascript"]
  },
  "message": "Successfully indexed repository 'my-project'"
}
```

### Example 2: Search for Functions

**Prompt to AI Assistant:**
```
Find all functions related to "authentication" in my codebase.
```

**Expected AI Action:**
```json
{
  "tool": "search_code",
  "arguments": {
    "query": "authentication",
    "type": "function",
    "max_results": 20
  }
}
```

### Example 3: Get File Metadata

**Prompt to AI Assistant:**
```
Show me the structure and metadata for the file src/auth/login.go
```

**Expected AI Action:**
```json
{
  "tool": "get_metadata",
  "arguments": {
    "file_path": "src/auth/login.go"
  }
}
```

### Example 4: Repository Overview

**Prompt to AI Assistant:**
```
Give me an overview of all indexed repositories and their statistics.
```

**Expected AI Actions:**
1. First, list repositories:
```json
{
  "tool": "list_repositories",
  "arguments": {}
}
```

2. Then, get detailed statistics:
```json
{
  "tool": "get_index_stats",
  "arguments": {}
}
```

### Example 5: Language-Specific Search

**Prompt to AI Assistant:**
```
Find all Python classes that contain "database" in their name or methods.
```

**Expected AI Action:**
```json
{
  "tool": "search_code",
  "arguments": {
    "query": "database",
    "type": "class",
    "language": "python",
    "max_results": 15
  }
}
```

## 5. Troubleshooting

### Common Issues and Solutions

#### Issue 1: Server Not Starting
**Symptoms:** MCP server shows as "disconnected" or "failed to start"

**Solutions:**
1. **Check binary path:**
   ```bash
   ls -la /path/to/your/my-mcp/bin/code-indexer
   chmod +x /path/to/your/my-mcp/bin/code-indexer
   ```

2. **Test manual start:**
   ```bash
   cd /path/to/your/my-mcp
   ./bin/code-indexer serve --log-level debug
   ```

3. **Check configuration syntax:**
   ```bash
   # Validate JSON configuration
   cat ~/.config/Cursor/User/globalStorage/rooveterinaryinc.roo-cline/settings/cline_mcp_settings.json | jq .
   ```

#### Issue 2: Tools Not Available
**Symptoms:** MCP tools don't appear in command palette

**Solutions:**
1. **Restart the IDE** completely
2. **Check MCP server logs** in the output panel
3. **Verify server connection:**
   - Look for "Connected to MCP server: code-indexer" message
   - Check for any error messages in logs

#### Issue 3: Indexing Fails
**Symptoms:** `index_repository` tool returns errors

**Solutions:**
1. **Check repository permissions:**
   ```bash
   ls -la /path/to/repository
   ```

2. **Verify disk space:**
   ```bash
   df -h
   ```

3. **Check configuration:**
   ```bash
   # Ensure index directory is writable
   ls -la ./index/
   ls -la ./repositories/
   ```

#### Issue 4: Search Returns No Results
**Symptoms:** `search_code` returns empty results for known code

**Solutions:**
1. **Verify repository was indexed:**
   ```json
   {
     "tool": "list_repositories",
     "arguments": {}
   }
   ```

2. **Check index statistics:**
   ```json
   {
     "tool": "get_index_stats",
     "arguments": {}
   }
   ```

3. **Try broader search:**
   ```json
   {
     "tool": "search_code",
     "arguments": {
       "query": "function",
       "max_results": 50
     }
   }
   ```

#### Issue 5: Performance Issues
**Symptoms:** Slow indexing or search responses

**Solutions:**
1. **Adjust configuration:**
   ```yaml
   # config.yaml
   indexer:
     max_file_size: 524288  # Reduce to 512KB
     exclude_patterns:
       - "*/node_modules/*"
       - "*/vendor/*"
       - "*/.git/*"
       - "*/build/*"
       - "*/dist/*"
   
   search:
     max_results: 50  # Reduce default results
   ```

2. **Monitor resource usage:**
   ```bash
   # Check disk usage
   du -sh ./index/
   du -sh ./repositories/
   
   # Monitor during indexing
   top -p $(pgrep code-indexer)
   ```

### Debug Mode

Enable debug logging for detailed troubleshooting:

1. **Update MCP configuration:**
   ```json
   {
     "command": "/path/to/your/my-mcp/bin/code-indexer",
     "args": ["serve", "--log-level", "debug"],
     "env": {
       "INDEXER_LOG_LEVEL": "debug"
     }
   }
   ```

2. **Check log output:**
   ```bash
   tail -f indexer.log
   ```

## 6. Best Practices

### Configuration Optimization

1. **Custom Configuration File:**
   ```yaml
   # ~/.config/code-indexer/config.yaml
   indexer:
     supported_extensions:
       - .go
       - .py
       - .js
       - .ts
       - .java
       - .cpp
       - .rs
     max_file_size: 1048576  # 1MB
     exclude_patterns:
       - "*/node_modules/*"
       - "*/vendor/*"
       - "*/.git/*"
       - "*/build/*"
       - "*/dist/*"
       - "*/target/*"
       - "*/__pycache__/*"
       - "*.min.js"
       - "*.min.css"
   
   search:
     max_results: 100
     highlight_snippets: true
     snippet_length: 200
   
   logging:
     level: info
     file: "~/.config/code-indexer/indexer.log"
   ```

2. **Use the custom configuration:**
   ```json
   {
     "command": "/path/to/your/my-mcp/bin/code-indexer",
     "args": ["serve", "--config", "~/.config/code-indexer/config.yaml"]
   }
   ```

### Usage Patterns

1. **Index Management:**
   - Index your main projects at the start of each session
   - Re-index repositories after major changes
   - Use descriptive names for repositories

2. **Effective Searching:**
   - Use specific search types (`function`, `class`, `variable`)
   - Filter by language for polyglot projects
   - Use repository filters for focused searches

3. **Performance Optimization:**
   - Exclude unnecessary file types and directories
   - Set appropriate file size limits
   - Monitor index size and clean up periodically

### Workflow Integration

1. **Project Setup:**
   ```
   AI: "Please index the current project and give me an overview of its structure."
   ```

2. **Code Exploration:**
   ```
   AI: "Find all authentication-related functions and show me their implementations."
   ```

3. **Refactoring Support:**
   ```
   AI: "Search for all usages of the 'UserService' class across the codebase."
   ```

4. **Documentation:**
   ```
   AI: "List all functions that lack documentation comments in the utils package."
   ```

### Security Considerations

1. **Repository Access:**
   - Ensure the MCP server has appropriate read permissions
   - Be cautious with private repositories and sensitive code

2. **Index Storage:**
   - Store indexes in secure locations
   - Consider encryption for sensitive codebases

3. **Network Access:**
   - The server runs locally and doesn't require network access
   - Git operations may require network for remote repositories

## Conclusion

With this setup, you'll have powerful code search and analysis capabilities directly integrated into your Cursor/Augment IDE. The MCP Code Indexer provides comprehensive code understanding that enhances AI-assisted development workflows.

For additional help, refer to:
- [MCP Code Indexer README](../README.md)
- [Development Guide](../DEVELOPMENT.md)
- [Basic Usage Examples](../examples/basic_usage.md)
