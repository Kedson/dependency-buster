# dependency-buster

> Universal dependency analyzer for any codebase • MCP Server implementations in TypeScript, Go, and Rust

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![MCP Compatible](https://img.shields.io/badge/MCP-Compatible-blue.svg)](https://modelcontextprotocol.io)

---

## Features

- **Universal Language Support** - Analyze PHP, JavaScript, TypeScript, Python, Go, Rust, Java, Ruby, C# and more
- **Dependency Analysis** - Production & development dependency tracking with version resolution
- **Security Auditing** - Vulnerability detection across package ecosystems
- **License Compliance** - Track and verify dependency licenses for legal compliance
- **Namespace Detection** - Code structure and module analysis
- **Triple Implementation** - Choose TypeScript (easy), Go (fast), or Rust (fastest)
- **Beautiful Dashboard** - Bauhaus-inspired visualization with charts and trees
- **CLI Reports** - ASCII-styled terminal output for CI/CD pipelines

---

## Quick Start

### 1. Clone and Build

```bash
git clone https://github.com/your-username/dependency-buster.git
cd dependency-buster

# Build all implementations
./build-all.sh

# Or build individually:
# TypeScript
cd dpb-mcp-typescript && npm install && npm run build

# Go
cd dpb-mcp-go && make build

# Rust  
cd dpb-mcp-rust && cargo build --release
```

### 2. Run Analysis

```bash
# Analyze current directory
cd your-project
/path/to/dependency-buster/dpb-benchmark/scripts/run-benchmark.sh .

# Analyze specific project
/path/to/dependency-buster/dpb-benchmark/scripts/run-benchmark.sh /path/to/project
```

### 3. View Results

```bash
# Open the generated dashboard
open dpb-benchmark/dashboard/index.html

# Or view terminal report (generated automatically)
```

---

## Implementations

| Implementation | Speed | Size | Best For |
|----------------|-------|------|----------|
| **TypeScript** | ★★★ | 0.01 MB | Quick setup, familiar ecosystem |
| **Go** | ★★★★★ | 6 MB | Production workloads, Go shops |
| **Rust** | ★★★★★ | 2.6 MB | Maximum performance, memory safety |

### Performance Benchmarks

| Metric | TypeScript | Go | Rust |
|--------|-----------|-----|------|
| Startup Time | ~430ms | ~177ms | ~176ms |
| Dependencies Analysis | ~570ms | ~164ms | ~140ms |
| Namespace Detection | ~1160ms | ~389ms | ~381ms |
| Security Audit | ~709ms | ~206ms | ~191ms |

---

## IDE Integration

dependency-buster works with any MCP-compatible IDE or tool:

### Cursor IDE

```json
{
  "mcpServers": {
    "dependency-buster": {
      "command": "node",
      "args": ["/path/to/dpb-mcp-typescript/build/server.js"]
    }
  }
}
```

### Claude Code CLI

```json
{
  "dependency-buster": {
    "command": "/path/to/dpb-mcp-go/build/dpb-mcp"
  }
}
```

### Cline Extension

```json
{
  "mcpServers": {
    "dependency-buster": {
      "command": "/path/to/dpb-mcp-rust/target/release/dpb-mcp"
    }
  }
}
```

**See [IDE_INTEGRATION.md](./IDE_INTEGRATION.md) for complete setup instructions**

---

## Available Tools (15 Total)

### Core Analysis Tools

| # | Tool | Description |
|---|------|-------------|
| 1 | `analyze_dependencies` | Comprehensive dependency analysis with production/dev breakdown and tree visualization |
| 2 | `analyze_psr4` | PSR-4 autoloading analysis and namespace compliance validation |
| 3 | `detect_namespaces` | Detect all namespaces and module structure in the codebase |
| 4 | `analyze_namespace_usage` | Analyze usage of a specific namespace across the codebase |
| 5 | `generate_dependency_graph` | Generate Mermaid diagram of dependency relationships |
| 6 | `audit_security` | Audit dependencies for security vulnerabilities and outdated packages |
| 7 | `analyze_licenses` | Analyze license distribution and compatibility across dependencies |
| 8 | `find_circular_dependencies` | Find circular dependency chains in the package graph |
| 9 | `analyze_multi_repo` | Analyze dependencies across multiple repositories |
| 10 | `generate_comprehensive_docs` | Generate comprehensive markdown documentation for a repository |
| 11 | `generate_mkdocs_docs` | Generate MkDocs-compatible documentation site with multi-file structure, navigation, and changelog |

### Tracking & AI Agent Tools

| # | Tool | Description |
|---|------|-------------|
| 11 | `track_dependencies` | Create timestamped snapshot of dependencies for tracking changes over time |
| 12 | `get_dependency_history` | Get dependency history with timestamps, recently added/updated, and stale packages |
| 13 | `check_compliance` | Check dependencies for compliance issues (licenses, outdated, deprecated) |
| 14 | `get_agent_suggestions` | Get structured suggestions for AI agents (Cursor, Cline, Claude Code) about dependency issues |

### MCP Tool Annotations

All tools include enterprise annotations:

```typescript
{
  readOnlyHint: true,      // Doesn't modify files
  idempotentHint: true,    // Same result on repeated calls
  cacheTtlSeconds: 300,    // Cache results for 5 minutes
  tags: ["analysis", "dependencies", "security"]
}
```

---

## Architecture

```
dependency-buster/
├── dpb-mcp-typescript/    # TypeScript implementation
│   ├── src/
│   │   ├── server.ts               # MCP server entry
│   │   ├── tools/                  # Analysis tools
│   │   ├── errors.ts               # Typed errors
│   │   ├── annotations.ts          # Tool annotations
│   │   └── auth.ts                 # Authentication
│   └── build/
│
├── dpb-mcp-go/          # Go implementation
│   ├── cmd/server/main.go
│   └── pkg/
│       ├── mcp/                    # MCP protocol
│       └── analyzer/               # Analysis tools
│
├── dpb-mcp-rust/        # Rust implementation
│   └── src/
│       ├── main.rs
│       ├── mcp/                    # MCP protocol
│       └── analyzer/               # Analysis tools
│
└── dpb-benchmark/              # Benchmark suite
    ├── scripts/run-benchmark.sh    # Main runner
    ├── dashboard/index.html        # Visualization
    └── results/                    # Generated reports
```

---

## Enterprise Features

All implementations support:

| Feature | Description |
|---------|-------------|
| **Typed Errors** | NotFoundError, ValidationError, AuthenticationError |
| **Authentication** | Static tokens via `MCP_TOKEN` environment variable |
| **HTTP Transport** | SSE streaming for remote access |
| **Request Context** | Credentials and request tracking |
| **Tool Annotations** | Caching, read-only hints, tags |

### Configuration

```bash
# Enable authentication
export MCP_AUTH_ENABLED=true
export MCP_TOKEN=your-secret-token

# Use HTTP transport
export MCP_TRANSPORT=http
export MCP_HTTP_PORT=3000

# Run server
./build/dpb-mcp
```

---

## Dashboard

The benchmark suite generates a beautiful Bauhaus-inspired dashboard:

- **Performance Metrics** - Startup time, binary size, tool execution times
- **Dependency Tree** - Interactive D3.js visualization
- **Security Summary** - Vulnerability breakdown by severity
- **License Matrix** - License distribution across packages
- **Namespace Map** - Code structure visualization
- **Smoke Tests** - Run tests directly from the dashboard

### Local Dashboard Server

A lightweight Go server with hot-reload for local development:

```bash
# Build and serve (opens browser automatically)
cd dpb-benchmark
make serve

# Development mode with Air hot-reload
make dev

# Install Air (first time only)
make install-air
```

The server runs on `http://localhost:8080` by default.

**Check if server is running:**
```bash
# Check port 8080
lsof -ti:8080 && echo "Server is running" || echo "Server is not running"

# Or check process
ps aux | grep dashboard-server
```

---

## Documentation Generation

The `generate_mkdocs_docs` tool creates comprehensive documentation from your dependency analysis data.

### Quick Start

```bash
# Generate docs (via Cursor IDE or CLI)
@dependency-buster generate_mkdocs_docs repo_path=/path/to/repo

# Or via command line
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"/path/to/repo"}},"id":1}' | node dpb-mcp-typescript/build/server.js
```

### Viewing Generated Documentation

**Option 1: MkDocs (Recommended)**
```bash
cd docs
pip install mkdocs mkdocs-material
mkdocs serve
# Opens at http://127.0.0.1:8000
```

**Option 2: View Markdown Directly**
```bash
# Open in your editor
code docs/index.md
```

**Option 3: HTML Output**
```bash
# Generate with format=html
# Then open docs/index.html in browser
```

### When Documentation is Generated

**Documentation does NOT regenerate automatically** on code changes. You need to:

1. **Manual Generation**: Call `generate_mkdocs_docs` tool when needed
2. **CI/CD Integration**: Add to your GitHub Actions/GitLab CI workflow:
   ```yaml
   - name: Generate Documentation
     run: |
       # Call generate_mkdocs_docs tool
       # Commit docs/ directory to repository
   ```
3. **Pre-commit Hook**: Add a git hook to regenerate docs before commits
4. **Release Workflow**: Generate docs as part of your release process

**Note**: Documentation generation is **not** included in smoke tests or benchmarks by default. To add it:
- Add `generate_mkdocs_docs` to benchmark tool list
- Include it in smoke test suite

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Dependency Analysis
on: [push]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          
      - name: Run dependency-buster
        run: |
          git clone https://github.com/your-username/dependency-buster.git /tmp/dpb
          cd /tmp/dpb/dpb-mcp-typescript
          npm install && npm run build
          cd $GITHUB_WORKSPACE
          /tmp/dpb/dpb-benchmark/scripts/run-benchmark.sh .
          
      - name: Upload Dashboard
        uses: actions/upload-artifact@v4
        with:
          name: dependency-report
          path: /tmp/dpb/dpb-benchmark/dashboard/index.html
```

---

## Development

### Prerequisites

- Node.js 18+ (TypeScript)
- Go 1.21+ (Go)
- Rust 1.70+ (Rust)
- jq (for JSON processing)

### Running Tests

```bash
# TypeScript
cd dpb-mcp-typescript
npm test

# Go
cd dpb-mcp-go
go test ./...

# Rust
cd dpb-mcp-rust
cargo test
```

### Building All

```bash
./build-all.sh
```

---

## Contributing

Contributions are welcome! Please read our [Contributing Guide](./CONTRIBUTING.md) first.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT License - see [LICENSE](./LICENSE) for details.

---

## Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io) - The protocol that powers this
- [Backstage](https://backstage.io) - Inspiration for enterprise features
- [AzuraCast](https://azuracast.com) - Test repository for benchmarking
