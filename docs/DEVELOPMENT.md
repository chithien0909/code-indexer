# Development Guide

This document provides information for developers who want to contribute to or extend the MCP Code Indexer.

## Project Structure

```
├── cmd/server/          # Main server executable
├── internal/            # Internal packages
│   ├── config/         # Configuration management
│   ├── indexer/        # Repository indexing logic
│   ├── parser/         # Language-specific parsers
│   ├── repository/     # Git repository management
│   ├── search/         # Search engine (Bleve)
│   └── server/         # MCP server implementation
├── pkg/                # Public packages
│   ├── types/          # Type definitions
│   └── utils/          # Utility functions
├── examples/           # Usage examples and tests
├── config.yaml         # Default configuration
├── Makefile           # Build automation
└── README.md          # Project documentation
```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, but recommended)

### Initial Setup

1. **Clone the repository:**
```bash
git clone https://github.com/my-mcp/code-indexer.git
cd code-indexer
```

2. **Set up development environment:**
```bash
make dev-setup
```

3. **Build the project:**
```bash
make build
```

4. **Run tests:**
```bash
make test
```

5. **Run the test example:**
```bash
make test-example
```

## Development Workflow

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/parser
```

### Code Quality

```bash
# Format code
make fmt

# Lint code
make lint

# Run all quality checks
make check
```

### Running the Server

```bash
# Run with default settings
make run

# Run with debug logging
make run-debug

# Run with custom config
./bin/code-indexer serve --config custom-config.yaml
```

## Architecture Overview

### Core Components

1. **MCP Server** (`internal/server`): Implements the Model Context Protocol and handles tool requests.

2. **Repository Manager** (`internal/repository`): Manages Git repositories, file discovery, and cloning operations.

3. **Parser Registry** (`internal/parser`): Language-specific parsers that extract metadata from source code.

4. **Search Engine** (`internal/search`): Bleve-based search functionality with indexing and querying capabilities.

5. **Indexer** (`internal/indexer`): Orchestrates the indexing process, combining repository management, parsing, and search indexing.

6. **Configuration** (`internal/config`): Manages application configuration with validation and defaults.

### Data Flow

1. **Indexing Process:**
   ```
   Repository URL/Path → Repository Manager → File Discovery → 
   Parser Registry → Metadata Extraction → Search Engine → Index Storage
   ```

2. **Search Process:**
   ```
   Search Query → MCP Server → Search Engine → 
   Index Lookup → Result Formatting → MCP Response
   ```

## Adding New Features

### Adding a New Language Parser

1. **Create the parser:**
```go
// internal/parser/mylang.go
type MyLangParser struct {
    BaseParser
}

func NewMyLangParser() *MyLangParser {
    return &MyLangParser{
        BaseParser: BaseParser{language: "mylang"},
    }
}

func (p *MyLangParser) Parse(content string, filePath string) (*types.CodeFile, error) {
    // Implementation here
}
```

2. **Register the parser:**
```go
// internal/parser/parser.go
func NewRegistry() *Registry {
    registry := &Registry{
        parsers: make(map[string]Parser),
    }
    
    // Register existing parsers...
    registry.Register(NewMyLangParser()) // Add this line
    
    return registry
}
```

3. **Add language mapping:**
```go
// internal/repository/manager.go
func (m *Manager) GetFileLanguage(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    
    languageMap := map[string]string{
        // Existing mappings...
        ".mylang": "mylang", // Add this line
    }
    
    // Rest of function...
}
```

4. **Update configuration:**
```yaml
# config.yaml
indexer:
  supported_extensions:
    - .mylang  # Add this line
```

### Adding a New MCP Tool

1. **Define the tool:**
```go
// internal/server/server.go
func (s *MCPServer) registerTools() error {
    // Existing tools...
    
    myTool := mcp.NewTool("my_tool",
        mcp.WithDescription("Description of my tool"),
        mcp.WithString("param1",
            mcp.Required(),
            mcp.Description("Parameter description"),
        ),
    )
    s.server.AddTool(myTool, s.handleMyTool)
    
    return nil
}
```

2. **Implement the handler:**
```go
func (s *MCPServer) handleMyTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    param1, err := request.RequireString("param1")
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Invalid param1: %v", err)), nil
    }
    
    // Tool logic here...
    
    result := map[string]any{
        "success": true,
        "data":    "result data",
    }
    
    resultJSON, _ := json.Marshal(result)
    return mcp.NewToolResultText(string(resultJSON)), nil
}
```

### Extending Search Functionality

1. **Add new search types:**
```go
// pkg/types/types.go
// Update SearchQuery.Type to include new types

// internal/search/engine.go
// Update buildSearchQuery to handle new types
```

2. **Add new metadata fields:**
```go
// pkg/types/types.go
// Add fields to relevant structs (Function, Class, etc.)

// internal/search/engine.go
// Update Document struct and indexing logic
```

## Testing

### Unit Tests

- Place test files next to the code they test with `_test.go` suffix
- Use table-driven tests for multiple test cases
- Mock external dependencies when necessary

### Integration Tests

- Use the `examples/test_server.go` as a reference
- Create temporary directories for file system operations
- Clean up resources in test teardown

### Test Coverage

```bash
make test-coverage
open coverage.html
```

## Performance Considerations

### Indexing Performance

- **File Size Limits**: Configure `max_file_size` to avoid indexing very large files
- **Exclude Patterns**: Use exclude patterns to skip unnecessary files
- **Batch Operations**: The search engine uses batch operations for better performance

### Search Performance

- **Index Size**: Monitor index size and consider cleanup strategies
- **Query Optimization**: Use specific search types and filters
- **Result Limits**: Set appropriate `max_results` values

### Memory Usage

- **Large Repositories**: Consider memory usage when indexing large repositories
- **Concurrent Operations**: The current implementation is single-threaded for simplicity

## Debugging

### Enable Debug Logging

```bash
./bin/code-indexer serve --log-level debug
```

### Common Issues

1. **Index Corruption**: Delete the index directory and re-index
2. **Permission Errors**: Check file and directory permissions
3. **Memory Issues**: Reduce file size limits or exclude more patterns
4. **Search Issues**: Check index statistics and verify data was indexed

### Profiling

```go
// Add to main.go for CPU profiling
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

## Contributing

### Code Style

- Follow Go conventions and use `gofmt`
- Add comments for exported functions and types
- Use meaningful variable and function names
- Keep functions focused and small

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `make check` to ensure quality
5. Submit a pull request with a clear description

### Commit Messages

Use conventional commit format:
```
feat: add support for Rust language parsing
fix: resolve index corruption on large files
docs: update installation instructions
test: add unit tests for parser registry
```

## Release Process

### Creating a Release

1. **Update version:**
```bash
make release VERSION=1.1.0
```

2. **Test the release:**
```bash
# Test binaries in release/ directory
```

3. **Create GitHub release:**
- Tag the commit with version number
- Upload release archives
- Write release notes

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Future Enhancements

### Planned Features

- [ ] Async indexing with progress tracking
- [ ] Index persistence and incremental updates
- [ ] Support for more programming languages
- [ ] Advanced search features (regex, proximity)
- [ ] Web UI for repository management
- [ ] Distributed indexing for large codebases
- [ ] Integration with popular IDEs

### Architecture Improvements

- [ ] Plugin system for custom parsers
- [ ] Configurable storage backends
- [ ] Horizontal scaling support
- [ ] Caching layer for frequent queries
- [ ] Metrics and monitoring integration

## Resources

- [Model Context Protocol Specification](https://github.com/modelcontextprotocol/specification)
- [Bleve Search Documentation](https://blevesearch.com/)
- [Go Documentation](https://golang.org/doc/)
- [Git Documentation](https://git-scm.com/doc)
