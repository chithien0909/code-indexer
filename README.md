# MCP Code Indexer

A Model Context Protocol (MCP) server written in Go that indexes source code from multiple repositories and provides powerful search capabilities for LLM applications.

## Features

- **Multi-Repository Support**: Index code from multiple Git repositories (local paths or URLs)
- **Language Agnostic**: Parse and index common source code file types (.go, .py, .js, .java, .cpp, etc.)
- **Rich Metadata Extraction**: Extract functions, classes, variables, comments, and documentation
- **Powerful Search**: Search by function names, variable names, code content, file paths, and comments
- **MCP Protocol**: Full compliance with Model Context Protocol for seamless LLM integration
- **High Performance**: Efficient indexing and search using Bleve search engine
- **Configurable**: Customizable file type filters and indexing options

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

## Development

### Building from Source

```bash
git clone https://github.com/my-mcp/code-indexer.git
cd code-indexer
go build -o code-indexer ./cmd/server
```

### Running Tests

```bash
go test ./...
```

## License

MIT License - see LICENSE file for details.
# code-indexer
