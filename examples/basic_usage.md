# MCP Code Indexer - Basic Usage Examples

This document provides examples of how to use the MCP Code Indexer server.

## Starting the Server

### Basic Start
```bash
./bin/code-indexer serve
```

### With Custom Configuration
```bash
./bin/code-indexer serve --config /path/to/config.yaml
```

### With Debug Logging
```bash
./bin/code-indexer serve --log-level debug
```

## MCP Tools Usage

The server provides several MCP tools that can be called by LLM applications:

### 1. Index Repository

Index a local repository:
```json
{
  "tool": "index_repository",
  "arguments": {
    "path": "/path/to/local/repository",
    "name": "my-project"
  }
}
```

Index a remote repository:
```json
{
  "tool": "index_repository",
  "arguments": {
    "path": "https://github.com/user/repo.git",
    "name": "external-project"
  }
}
```

### 2. Search Code

Basic search:
```json
{
  "tool": "search_code",
  "arguments": {
    "query": "function main"
  }
}
```

Search for functions only:
```json
{
  "tool": "search_code",
  "arguments": {
    "query": "handleRequest",
    "type": "function"
  }
}
```

Search in specific language:
```json
{
  "tool": "search_code",
  "arguments": {
    "query": "class User",
    "type": "class",
    "language": "python"
  }
}
```

Search in specific repository:
```json
{
  "tool": "search_code",
  "arguments": {
    "query": "error handling",
    "repository": "my-project",
    "max_results": 50
  }
}
```

### 3. Get File Metadata

Get metadata for a specific file:
```json
{
  "tool": "get_metadata",
  "arguments": {
    "file_path": "src/main.go"
  }
}
```

Get metadata with repository filter:
```json
{
  "tool": "get_metadata",
  "arguments": {
    "file_path": "handlers/user.py",
    "repository": "my-project"
  }
}
```

### 4. List Repositories

List all indexed repositories:
```json
{
  "tool": "list_repositories",
  "arguments": {}
}
```

### 5. Get Index Statistics

Get indexing statistics:
```json
{
  "tool": "get_index_stats",
  "arguments": {}
}
```

## Example Responses

### Index Repository Response
```json
{
  "success": true,
  "repository": {
    "id": "a1b2c3d4e5f6g7h8",
    "name": "my-project",
    "path": "/path/to/repository",
    "indexed_at": "2025-01-15T10:30:00Z",
    "file_count": 150,
    "total_lines": 12500,
    "languages": ["go", "python", "javascript"]
  },
  "message": "Successfully indexed repository 'my-project'"
}
```

### Search Code Response
```json
{
  "success": true,
  "query": {
    "query": "handleRequest",
    "type": "function",
    "max_results": 100
  },
  "results": [
    {
      "id": "func:a1b2c3d4:src/handlers.go:handleRequest:25",
      "repository_id": "a1b2c3d4e5f6g7h8",
      "repository": "my-project",
      "file_path": "src/handlers.go",
      "language": "go",
      "type": "function",
      "name": "handleRequest",
      "content": "func handleRequest(w http.ResponseWriter, r *http.Request) {",
      "snippet": "func handleRequest(w http.ResponseWriter, r *http.Request) {\n    // Handle HTTP request\n    ...",
      "start_line": 25,
      "end_line": 45,
      "score": 0.95,
      "highlights": {
        "name": "<mark>handleRequest</mark>"
      }
    }
  ],
  "total_found": 3
}
```

### Get Metadata Response
```json
{
  "success": true,
  "metadata": {
    "id": "a1b2c3d4:src/main.go",
    "repository_id": "a1b2c3d4e5f6g7h8",
    "path": "/path/to/repository/src/main.go",
    "relative_path": "src/main.go",
    "language": "go",
    "extension": ".go",
    "size": 2048,
    "lines": 85,
    "functions": [
      {
        "name": "main",
        "start_line": 15,
        "end_line": 25,
        "signature": "func main() {",
        "parameters": [],
        "return_type": ""
      }
    ],
    "classes": [],
    "variables": [
      {
        "name": "port",
        "type": "string",
        "start_line": 10,
        "is_constant": true
      }
    ],
    "imports": [
      {
        "module": "fmt",
        "start_line": 3
      },
      {
        "module": "net/http",
        "start_line": 4
      }
    ],
    "comments": [
      {
        "text": "Main entry point for the application",
        "start_line": 14,
        "end_line": 14,
        "type": "line"
      }
    ]
  }
}
```

## Configuration Examples

### Custom File Types
```yaml
indexer:
  supported_extensions:
    - .go
    - .py
    - .js
    - .ts
    - .java
    - .cpp
    - .rs
    - .rb
    - .php
    - .custom  # Add custom extensions
```

### Exclude Patterns
```yaml
indexer:
  exclude_patterns:
    - "*/node_modules/*"
    - "*/vendor/*"
    - "*/.git/*"
    - "*/build/*"
    - "*/dist/*"
    - "*/target/*"
    - "*/__pycache__/*"
    - "*/my_custom_exclude/*"  # Custom exclusions
```

### Search Configuration
```yaml
search:
  max_results: 200
  highlight_snippets: true
  snippet_length: 300
  fuzzy_tolerance: 0.3
```

### Logging Configuration
```yaml
logging:
  level: debug
  file: "/var/log/code-indexer.log"
  json_format: true
  max_size: 100
  max_backups: 5
  max_age: 30
```

## Tips and Best Practices

1. **Repository Naming**: Use descriptive names when indexing repositories to make searching easier.

2. **Search Types**: Use specific search types (`function`, `class`, `variable`, `comment`) for more precise results.

3. **Language Filtering**: Filter by language when working with polyglot repositories.

4. **File Size Limits**: Adjust `max_file_size` in configuration for very large files.

5. **Exclude Patterns**: Use exclude patterns to avoid indexing generated files, dependencies, and build artifacts.

6. **Regular Re-indexing**: Re-index repositories periodically to capture changes.

7. **Monitoring**: Use `get_index_stats` to monitor the health and size of your index.

## Troubleshooting

### Common Issues

1. **Repository Not Found**: Ensure the path exists and is accessible.
2. **Permission Denied**: Check file and directory permissions.
3. **Large Index**: Monitor disk space and consider excluding more file types.
4. **Slow Search**: Reduce `max_results` or add more specific filters.
5. **Memory Usage**: Adjust configuration for large repositories.

### Debug Mode

Run with debug logging to troubleshoot issues:
```bash
./bin/code-indexer serve --log-level debug
```

Check the log file for detailed information about indexing and search operations.
