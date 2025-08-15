# MCP Code Indexer - Project Overview

## Project Purpose
The MCP Code Indexer is a Model Context Protocol (MCP) server written in Go that indexes source code from multiple repositories and provides powerful search capabilities for LLM applications. It enables AI assistants to understand and search through codebases efficiently.

## Key Features
- **Multi-Repository Support**: Index code from multiple Git repositories (local paths or URLs)
- **Language Agnostic**: Parse and index common source code file types (.go, .py, .js, .java, .cpp, etc.)
- **Rich Metadata Extraction**: Extract functions, classes, variables, comments, and documentation
- **Powerful Search**: Search by function names, variable names, code content, file paths, and comments
- **MCP Protocol**: Full compliance with Model Context Protocol for seamless LLM integration
- **High Performance**: Efficient indexing and search using Bleve search engine
- **Configurable**: Customizable file type filters and indexing options

## Architecture
The project follows a clean, modular architecture with clear separation of concerns:

### Core Components
1. **MCP Server** (`internal/server`): Implements MCP protocol and handles tool requests
2. **Repository Manager** (`internal/repository`): Manages Git repositories and file discovery
3. **Parser Registry** (`internal/parser`): Language-specific parsers for metadata extraction
4. **Search Engine** (`internal/search`): Bleve-based indexing and search functionality
5. **Indexer** (`internal/indexer`): Orchestrates the indexing process
6. **Configuration** (`internal/config`): Manages application configuration

### Technology Stack
- **Language**: Go 1.23
- **MCP Library**: github.com/mark3labs/mcp-go v0.37.0
- **Search Engine**: github.com/blevesearch/bleve/v2 v2.3.10
- **Git Operations**: github.com/go-git/go-git/v5 v5.11.0
- **CLI Framework**: github.com/spf13/cobra v1.8.0
- **Configuration**: github.com/spf13/viper v1.18.2
- **Logging**: go.uber.org/zap v1.26.0

## MCP Tools Provided
The server exposes 5 MCP tools for LLM applications:

1. **index_repository**: Index a Git repository for searching
   - Parameters: path (required), name (optional)

2. **search_code**: Search across all indexed repositories
   - Parameters: query (required), type, language, repository, max_results

3. **get_metadata**: Get detailed metadata for a specific file
   - Parameters: file_path (required), repository (optional)

4. **list_repositories**: List all indexed repositories with statistics
   - Parameters: None

5. **get_index_stats**: Get indexing statistics and information
   - Parameters: None

## Supported Languages
The parser registry includes specialized parsers for:
- **Go**: Functions, structs, variables, constants, imports, comments
- **Python**: Functions, classes, variables, imports, comments
- **JavaScript**: Functions, classes, variables, imports, comments
- **Java**: Methods, classes, fields, imports, comments
- **Generic**: Basic parsing for any text file with common comment styles

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
├── docs/               # Documentation
├── scripts/            # Setup and utility scripts
├── config.yaml         # Default configuration
├── Makefile           # Build automation
└── README.md          # Project documentation
```

## Build and Development
- **Build**: `make build` creates binary in `bin/code-indexer`
- **Test**: `make test` runs unit tests
- **Development Setup**: `make dev-setup` installs development tools
- **Integration Test**: `make test-example` runs comprehensive test

## IDE Integration
The project includes comprehensive integration guides and tools for:
- **Cursor IDE**: Complete setup automation with `scripts/setup-cursor-integration.sh`
- **Augment IDE**: Configuration examples and setup instructions
- **Documentation**: Detailed guides in `docs/CURSOR_AUGMENT_INTEGRATION.md`

## Configuration
Highly configurable via YAML with support for:
- File type filters and exclude patterns
- Search engine parameters
- Logging configuration
- Index and repository storage locations
- Performance tuning options

## Key Files
- **Main Entry**: `cmd/server/main.go`
- **MCP Server**: `internal/server/server.go`
- **Core Types**: `pkg/types/types.go`
- **Configuration**: `internal/config/config.go`
- **Parser Registry**: `internal/parser/parser.go`
- **Search Engine**: `internal/search/engine.go`
- **Repository Manager**: `internal/repository/manager.go`
- **Indexer**: `internal/indexer/indexer.go`

This project represents a complete, production-ready MCP server that bridges the gap between AI assistants and code understanding, enabling powerful code search and analysis capabilities.