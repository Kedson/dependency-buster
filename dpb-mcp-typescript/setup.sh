#!/bin/bash

# PHP Dependency MCP Server Setup Script
# This script will install dependencies, build the project, and configure MCP clients

set -e

echo "🚀 PHP Dependency MCP Server Setup"
echo "===================================="
echo ""

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "❌ Error: Node.js 18+ required (you have $(node -v))"
    exit 1
fi
echo "✓ Node.js version: $(node -v)"

# Install dependencies
echo ""
echo "📦 Installing dependencies..."
npm install

# Build project
echo ""
echo "🔨 Building TypeScript..."
npm run build

# Check if build was successful
if [ ! -f "build/server.js" ]; then
    echo "❌ Build failed - server.js not found"
    exit 1
fi
echo "✓ Build successful"

# Make executable
chmod +x build/server.js

# Test the server
echo ""
echo "🧪 Testing MCP server..."
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | node build/server.js > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ Server test passed"
else
    echo "❌ Server test failed"
    exit 1
fi

# Offer to install globally
echo ""
read -p "Install globally? This allows you to run 'dpb-mcp' from anywhere (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    npm install -g .
    echo "✓ Installed globally as 'dpb-mcp'"
else
    npm link
    echo "✓ Linked locally (use 'npm unlink' to remove)"
fi

# Detect available MCP clients
echo ""
echo "🔍 Detecting MCP clients..."

CLAUDE_CODE_INSTALLED=false
CURSOR_INSTALLED=false

if command -v claude &> /dev/null; then
    CLAUDE_CODE_INSTALLED=true
    echo "✓ Claude Code detected"
fi

if command -v cursor &> /dev/null || [ -d "/Applications/Cursor.app" ] || [ -d "$HOME/Applications/Cursor.app" ]; then
    CURSOR_INSTALLED=true
    echo "✓ Cursor detected"
fi

# Configure Claude Code
if [ "$CLAUDE_CODE_INSTALLED" = true ]; then
    echo ""
    read -p "Configure Claude Code? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        claude mcp add php-analyzer --scope user -- dpb-mcp
        echo "✓ Claude Code configured"
    fi
fi

# Show Cursor instructions
if [ "$CURSOR_INSTALLED" = true ]; then
    echo ""
    echo "📝 To configure Cursor:"
    echo "   1. Open Cursor Settings"
    echo "   2. Go to MCP section"
    echo "   3. Add this configuration:"
    echo ""
    echo '   {'
    echo '     "mcpServers": {'
    echo '       "php-analyzer": {'
    echo '         "command": "dpb-mcp"'
    echo '       }'
    echo '     }'
    echo '   }'
    echo ""
fi

# Show next steps
echo ""
echo "✨ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Test with test repository:"
echo "   git clone https://github.com/test repository/test repository.git ~/test/test repository"
echo "   cd ~/test/test repository"
echo "   claude .  # or open in Cursor"
echo ""
echo "2. For Faith FM multi-repo analysis:"
echo "   cp config/faith-fm-repos.example.json config/faith-fm-repos.json"
echo "   # Edit the paths in the config file"
echo ""
echo "3. Read the full documentation:"
echo "   cat README.md"
echo ""
