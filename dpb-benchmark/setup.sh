#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║   PHP MCP Server Benchmark - Quick Setup                  ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Check if we're in the right directory
if [ ! -f "README.md" ]; then
    echo "❌ Error: Please run this script from the dpb-benchmark directory"
    exit 1
fi

# Make scripts executable
echo "📝 Making scripts executable..."
chmod +x scripts/*.sh
chmod +x scripts/*.py

# Check for test repository
test repository_PATH="${HOME}/test/test repository"
if [ ! -d "$test repository_PATH" ]; then
    echo "📥 Cloning test repository test repository..."
    git clone --depth 1 https://github.com/test repository/test repository.git "$test repository_PATH"
    echo "✅ test repository cloned to $test repository_PATH"
else
    echo "✅ test repository already exists at $test repository_PATH"
fi

# Check for Python
if ! command -v python3 &> /dev/null; then
    echo "⚠️  Warning: python3 not found - report generation will be limited"
fi

# Create results directory
mkdir -p results

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Build all three implementations:"
echo "     - TypeScript: cd dpb-mcp-complete && npm install && npm run build"
echo "     - Go:         cd dpb-mcp-go && make build"
echo "     - Rust:       cd dpb-mcp-rust && cargo build --release"
echo ""
echo "  2. Run benchmark:"
echo "     ./scripts/run-benchmark.sh"
echo ""
echo "  3. View results:"
echo "     open dashboard/index.html"
echo ""
