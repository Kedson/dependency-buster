# Quick Start Guide - Go Edition

Get the Go version running in 3 steps!

## Prerequisites

- Go 1.22 or later
- Make (optional, but recommended)

Check Go version:
```bash
go version
```

## 1. Extract and Build

```bash
# Extract
unzip dpb-mcp-go.zip
cd dpb-mcp-go

# Download dependencies
go mod download

# Build
make build
# OR without Make:
go build -o build/dpb-mcp ./cmd/server
```

## 2. Install (Optional)

```bash
# Install to system PATH
make install
# OR manually:
sudo cp build/dpb-mcp /usr/local/bin/
```

## 3. Configure MCP Client

### Claude Code
```bash
claude mcp add php-analyzer --scope user -- dpb-mcp
```

### Cursor
Add to `.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "php-analyzer": {
      "command": "dpb-mcp"
    }
  }
}
```

## 4. Test

```bash
# Clone AzuraCast
git clone https://github.com/AzuraCast/AzuraCast.git ~/test/azuracast
cd ~/test/azuracast

# Start Claude Code or Cursor
claude .
```

Ask:
```
Analyze this PHP repository comprehensively
```

## Build for All Platforms

```bash
make build-all
```

This creates binaries for:
- Linux (amd64, arm64)
- macOS (Intel, Apple Silicon)
- Windows (amd64)

Find them in `build/` directory.

## Performance Check

```bash
# Check binary size
ls -lh build/dpb-mcp

# Test startup time
time echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | ./build/dpb-mcp
```

Expected:
- Binary: ~8MB
- Startup: <10ms
- Memory: ~15MB during analysis

## Troubleshooting

### "go: command not found"
Install Go from https://go.dev/dl/

### Build errors
```bash
go mod tidy
go mod download
```

### Permission denied
```bash
chmod +x build/dpb-mcp
```

## Next Steps

- Compare with TypeScript version
- Try multi-repo analysis
- Test with your Dependency Buster repos
- Wait for Rust version! 🦀

Enjoy the speed! 🚀
