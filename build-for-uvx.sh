#!/bin/bash

set -e

echo "🔧 Building MCP Code Indexer for uvx installation..."

# Step 1: Build the Go binary
echo "📦 Building Go binary..."
mkdir -p python/mcp_code_indexer/bin
go build -o python/mcp_code_indexer/bin/code-indexer ./cmd/server

# Make it executable
chmod +x python/mcp_code_indexer/bin/code-indexer

# Test the binary
echo "🧪 Testing binary..."
./python/mcp_code_indexer/bin/code-indexer --version

echo "✅ Binary built successfully!"

# Step 2: Test local installation
echo "🧪 Testing local Python package installation..."
python -m pip install -e . --force-reinstall

# Test the installed package
echo "✅ Testing installed package..."
python -c "import mcp_code_indexer; print(f'Version: {mcp_code_indexer.__version__}')"

echo "🎉 Build completed successfully!"
echo ""
echo "Next steps:"
echo "1. Commit and push: git add . && git commit -m 'Add uvx support' && git push"
echo "2. Test uvx installation: uvx install git+https://github.com/chithien0909/code-indexer.git --force"
echo "3. Use in IDE with the configuration provided in the documentation"
