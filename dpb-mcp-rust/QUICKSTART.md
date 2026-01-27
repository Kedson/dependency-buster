# Quick Start Guide - Rust Edition 🦀

Get the blazing-fast Rust version running in 2 steps!

## Prerequisites

- Rust 1.70 or later

Check Rust version:
```bash
rustc --version
```

If not installed:
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

## 1. Extract and Build

```bash
# Extract
unzip php-dependency-mcp-rust.zip
cd php-dependency-mcp-rust

# Build release version (optimized)
cargo build --release
# OR use Makefile:
make release
```

Build time: ~2 minutes (first time), ~10 seconds (subsequent builds)
Binary location: `target/release/php-dependency-mcp`

## 2. Install (Optional)

```bash
# Install to system PATH
make install
# OR manually:
cargo install --path .
# OR copy:
sudo cp target/release/php-dependency-mcp /usr/local/bin/
```

## 3. Configure MCP Client

### Claude Code
```bash
claude mcp add php-analyzer --scope user -- php-dependency-mcp
```

### Cursor
Add to `.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "php-analyzer": {
      "command": "php-dependency-mcp"
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

## Super Optimized Build

For absolute minimum binary size and maximum speed:

```bash
make release-small
# Creates ~2MB binary!
```

## Performance Check

```bash
# Check binary size
ls -lh target/release/php-dependency-mcp

# Test startup time
time echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | \
  ./target/release/php-dependency-mcp
```

Expected:
- Binary: 2-5MB (with strip)
- Startup: <5ms
- Memory: ~8MB during analysis

## Troubleshooting

### "cargo: command not found"
Install Rust from https://rustup.rs/

### Slow compile
First compile is slow (downloads deps). Subsequent builds are fast.
Use `cargo build` for faster dev builds.

### Permission denied
```bash
chmod +x target/release/php-dependency-mcp
```

### Link errors
```bash
# Update dependencies
cargo update
# Clean and rebuild
cargo clean
cargo build --release
```

## Development Mode

```bash
# Fast compile, slower runtime
cargo build
cargo run
```

## Next Steps

- Compare with TypeScript and Go versions
- Run benchmarks on large projects
- Try multi-repo analysis
- Test with your Faith FM repos

Enjoy the speed! 🦀⚡
