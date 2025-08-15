# MCP Code Indexer

A Model Context Protocol (MCP) server written in Go that indexes source code from multiple repositories and provides powerful search capabilities for LLM applications.

## Features

- **Multi-Repository Support**: Index code from multiple Git repositories (local paths or URLs)
- **Multi-IDE Support**: Concurrent connections from multiple IDE instances (Cursor, VS Code, etc.)
- **Language Agnostic**: Parse and index common source code file types (.go, .py, .js, .java, .cpp, etc.)
- **Rich Metadata Extraction**: Extract functions, classes, variables, comments, and documentation
- **Powerful Search**: Search by function names, variable names, code content, file paths, and comments
- **MCP Protocol**: Full compliance with Model Context Protocol for seamless LLM integration
- **High Performance**: Efficient indexing and search using Bleve search engine with concurrent access
- **Resource Management**: Advanced locking and session isolation for conflict-free operation
- **Configurable**: Customizable file type filters, connection limits, and isolation modes

## Quick Start

### Installation

#### Option 1: Build from Source
```bash
git clone https://github.com/my-mcp/code-indexer.git
cd code-indexer
make build
```

#### Option 2: Using Go Install
```bash
go install github.com/my-mcp/code-indexer@latest
```

#### Option 3: Download Pre-built Binaries
Download the latest release from the [releases page](https://github.com/my-mcp/code-indexer/releases).

### Quick Start

#### Method 1: Direct uvx Installation (Recommended)

Install and use directly with uvx - no daemon required:

```bash
# Install via uvx
uvx install git+https://github.com/my-mcp/code-indexer.git

# Test the installation
uvx --from git+https://github.com/my-mcp/code-indexer.git code-indexer --version
```

Configure your IDE (example for Cursor):
```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "uvx",
      "args": [
        "--from", "git+https://github.com/my-mcp/code-indexer.git",
        "code-indexer", "mcp-server"
      ]
    }
  }
}
```

#### Method 2: Build from Source

1. **Build the server:**
```bash
make build
```

2. **Test the installation:**
```bash
make test-example
```

3. **Start the MCP server:**
```bash
./bin/code-indexer serve
```

4. **Or run with debug logging:**
```bash
./bin/code-indexer serve --log-level debug
```

5. **For multiple IDE instances:**
```bash
./bin/code-indexer daemon --port 9991
```

### MCP Tools

The server provides these MCP tools for LLM applications:

- **`index_repository`**: Index a Git repository (local path or URL)
- **`search_code`**: Search across indexed code with filters
- **`get_metadata`**: Retrieve detailed metadata for specific files
- **`list_repositories`**: List all indexed repositories with statistics
- **`get_index_stats`**: Get comprehensive indexing statistics

### Configuration

Create a `config.yaml` file:

```yaml
indexer:
  supported_extensions:
    - .go
    - .py
    - .js
    - .java
    - .cpp
    - .h
    - .rs
    - .rb
    - .php
  max_file_size: 1048576  # 1MB
  exclude_patterns:
    - "*/node_modules/*"
    - "*/vendor/*"
    - "*/.git/*"

search:
  max_results: 100
  highlight_snippets: true

logging:
  level: info
  file: "indexer.log"
```

## Architecture

The MCP Code Indexer consists of several key components:

- **MCP Server**: Handles protocol communication and tool registration
- **Repository Manager**: Manages Git repository operations and file discovery
- **Parser Engine**: Language-specific parsers for metadata extraction
- **Search Engine**: Bleve-based indexing and search functionality
- **Configuration Manager**: Handles settings and file type configurations

## MCP Tools

### index_repository
Index a Git repository for searching.

**Parameters:**
- `path` (string): Local path or Git URL to repository
- `name` (string, optional): Custom name for the repository

### search_code
Search across all indexed repositories.

**Parameters:**
- `query` (string): Search query
- `type` (string, optional): Search type ("function", "class", "variable", "content", "file", "comment")
- `language` (string, optional): Filter by programming language
- `repository` (string, optional): Filter by repository name

### get_metadata
Get detailed metadata for a specific file.

**Parameters:**
- `file_path` (string): Path to the file
- `repository` (string, optional): Repository name

### list_repositories
List all indexed repositories with statistics.

## IDE Integration

### Cursor/Augment IDE Setup

For detailed integration with Cursor or Augment IDE:

1. **Quick Setup:**
   ```bash
   ./scripts/setup-cursor-integration.sh
   ```

2. **Manual Setup:** See [Cursor/Augment Integration Guide](docs/CURSOR_AUGMENT_INTEGRATION.md)

3. **Tool Reference:** See [MCP Tools Quick Reference](docs/MCP_TOOLS_REFERENCE.md)

### Other IDEs

The MCP Code Indexer can work with any IDE that supports the Model Context Protocol. Configuration will vary by IDE.

## Development

### Building from Source

```bash
git clone https://github.com/my-mcp/code-indexer.git
cd code-indexer
make build
```

### Running Tests

```bash
make test
```

### Development Setup

```bash
make dev-setup
```

For detailed development information, see [Development Guide](docs/DEVELOPMENT.md).

## Documentation

### Getting Started
- [Quick Start Guide](docs/QUICK_START.md) - Fast setup and basic usage
- [Basic Usage Examples](examples/basic_usage.md) - Usage examples and patterns
- [API Usage Guide](docs/API_USAGE.md) - Detailed API usage documentation

### Integration & Setup
- [**🚀 Direct uvx Installation**](docs/UVX_INSTALLATION.md) - Install and run directly via uvx (recommended)
- [**Quick IDE Setup Script**](scripts/setup-ide-integration.sh) - Automated setup for Cursor and Augment Code
- [uvx Integration Guide](docs/UVX_MCP_SETUP.md) - Modern Python-based MCP client setup
- [Cursor IDE Setup](docs/CURSOR_MCP_SETUP.md) - Step-by-step Cursor IDE configuration
- [Augment Code Setup](docs/AUGMENT_CODE_SETUP.md) - Step-by-step Augment Code configuration
- [Multi-IDE Setup Guide](docs/MULTI_IDE_SETUP.md) - Configure multiple IDE instances concurrently
- [Multi-IDE Architecture](docs/MULTI_IDE_ARCHITECTURE.md) - Technical architecture for concurrent IDE support
- [Cursor/Augment Integration Guide](docs/CURSOR_AUGMENT_INTEGRATION.md) - IDE setup and configuration
- [MCP Integration Summary](docs/MCP_INTEGRATION_SUMMARY.md) - MCP protocol integration details

### Tools & Reference
- [MCP Tools Reference](docs/MCP_TOOLS_REFERENCE.md) - Quick reference for all MCP tools
- [Complete Tools Documentation](docs/TOOLS.md) - Comprehensive tools documentation
- [24 Tools Summary](docs/FINAL_24_TOOLS_SUMMARY.md) - Overview of all available tools

### Development & Implementation
- [Development Guide](docs/DEVELOPMENT.md) - Contributing and extending the codebase
- [Implementation Status](docs/IMPLEMENTATION_STATUS.md) - Current implementation status
- [Modular Refactor Summary](docs/MODULAR_REFACTOR_SUMMARY.md) - Architecture and refactoring details

### Technical Documentation
- [Multi-Session Implementation](docs/MULTI_SESSION_IMPLEMENTATION.md) - Multi-session support details
- [Multi-Session Success](docs/MULTI_SESSION_SUCCESS.md) - Multi-session implementation results
- [Final Multi-Session Summary](docs/FINAL_MULTI_SESSION_SUMMARY.md) - Complete multi-session overview
- [Expanded Implementation Summary](docs/EXPANDED_IMPLEMENTATION_SUMMARY.md) - Detailed implementation summary
- [Integration Summary](docs/INTEGRATION_SUMMARY.md) - Integration overview
- [Enhancement Summary](docs/ENHANCEMENT_SUMMARY.md) - Recent enhancements and improvements

## License

MIT License - see LICENSE file for details.
