# MCP Code Indexer Makefile

# Variables
BINARY_NAME=code-indexer
BUILD_DIR=bin
MAIN_PATH=./cmd/server
VERSION?=1.0.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Default target
.PHONY: all
all: clean build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	@echo "✅ Multi-platform build completed"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "✅ Tests completed"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run the test example
.PHONY: test-example
test-example:
	@echo "Running test example..."
	$(GOCMD) run examples/test_server.go
	@echo "✅ Test example completed"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ Clean completed"

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies updated"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "✅ Code formatted"

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi
	@echo "✅ Linting completed"

# Run the server
.PHONY: run
run: build
	@echo "Starting MCP Code Indexer server..."
	./$(BUILD_DIR)/$(BINARY_NAME) serve

# Run with debug logging
.PHONY: run-debug
run-debug: build
	@echo "Starting MCP Code Indexer server with debug logging..."
	./$(BUILD_DIR)/$(BINARY_NAME) serve --log-level debug

# Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Create a release package
.PHONY: release
release: clean build-all
	@echo "Creating release package..."
	@mkdir -p release
	
	# Copy binaries
	cp $(BUILD_DIR)/* release/
	
	# Copy documentation
	cp README.md release/
	cp LICENSE release/
	cp config.yaml release/config.example.yaml
	cp -r examples release/
	
	# Create archives
	cd release && tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 README.md LICENSE config.example.yaml examples/
	cd release && tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 README.md LICENSE config.example.yaml examples/
	cd release && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 README.md LICENSE config.example.yaml examples/
	cd release && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 README.md LICENSE config.example.yaml examples/
	cd release && zip -r $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe README.md LICENSE config.example.yaml examples/
	
	@echo "✅ Release packages created in release/"

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	
	# Install development tools
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	
	@echo "✅ Development environment setup completed"

# Check code quality
.PHONY: check
check: fmt lint test
	@echo "✅ All checks passed"

# Show help
.PHONY: help
help:
	@echo "MCP Code Indexer - Available commands:"
	@echo ""
	@echo "  build         Build the binary"
	@echo "  build-all     Build for multiple platforms"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  test-example  Run the test example"
	@echo "  clean         Clean build artifacts"
	@echo "  deps          Download and update dependencies"
	@echo "  fmt           Format code"
	@echo "  lint          Lint code"
	@echo "  run           Build and run the server"
	@echo "  run-debug     Build and run with debug logging"
	@echo "  install       Install binary to GOPATH/bin"
	@echo "  release       Create release packages"
	@echo "  dev-setup     Setup development environment"
	@echo "  check         Run all quality checks (fmt, lint, test)"
	@echo "  help          Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make run"
	@echo "  make release VERSION=1.0.1"

# Default help target
.DEFAULT_GOAL := help
