# dependency-buster (dpb) IDE Integration Guide

> Universal dependency analyzer MCP server for any programming language

## Table of Contents
- [Overview](#overview)
- [Quick Start](#quick-start)
- [Cursor IDE](#cursor-ide)
- [VSCode](#vscode)
- [Cline Extension](#cline-extension)
- [Claude Code CLI](#claude-code-cli)
- [Available Tools](#available-tools)
- [Configuration](#configuration)

---

## Overview

dependency-buster is an MCP (Model Context Protocol) server that analyzes dependencies in any codebase. It provides:

- **Dependency Analysis**: Production & development dependencies
- **Security Auditing**: Vulnerability detection across package ecosystems
- **License Compliance**: Track and verify dependency licenses
- **Namespace Detection**: Code structure analysis
- **PSR-4 Autoloading**: PHP autoloader validation

### Supported Languages

| Language | Package Manager | Detection File |
|----------|----------------|----------------|
| PHP | Composer | `composer.json` |
| JavaScript | NPM/Yarn/PNPM | `package.json` |
| TypeScript | NPM/Yarn/PNPM | `package.json` + `tsconfig.json` |
| Python | Pip/Poetry/Pipenv | `requirements.txt`, `pyproject.toml` |
| Go | Go Modules | `go.mod` |
| Rust | Cargo | `Cargo.toml` |
| Java | Maven/Gradle | `pom.xml`, `build.gradle` |
| Ruby | Bundler | `Gemfile` |
| C# | NuGet | `*.csproj` |

---

## Quick Start

### 1. Build the MCP Server

Choose your preferred implementation:

```bash
# TypeScript (recommended for quick setup)
cd dpb-mcp-typescript
npm install && npm run build

# Go (fastest execution)
cd dpb-mcp-go
make build

# Rust (best performance)
cd dpb-mcp-rust
cargo build --release
```

### 2. Test the Server

```bash
# TypeScript
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | node build/server.js

# Go
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./build/dpb-mcp

# Rust
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./target/release/dpb-mcp
```

---

## Cursor IDE

### Configuration (mcpServers.json)

Add to your Cursor MCP configuration file:

**macOS**: `~/.cursor/mcp/mcpServers.json`
**Windows**: `%APPDATA%\Cursor\mcp\mcpServers.json`
**Linux**: `~/.config/cursor/mcp/mcpServers.json`

```json
{
  "mcpServers": {
    "dependency-buster": {
      "command": "node",
      "args": ["/path/to/dpb-mcp-workspace/dpb-mcp-typescript/build/server.js"],
      "env": {
        "MCP_AUTH_ENABLED": "false"
      }
    }
  }
}
```

### Alternative: Go Implementation

```json
{
  "mcpServers": {
    "dependency-buster": {
      "command": "/path/to/dpb-mcp-workspace/dpb-mcp-go/build/dpb-mcp",
      "env": {}
    }
  }
}
```

### Alternative: Rust Implementation

```json
{
  "mcpServers": {
    "dependency-buster": {
      "command": "/path/to/dpb-mcp-workspace/dpb-mcp-rust/target/release/dpb-mcp",
      "env": {}
    }
  }
}
```

### Using in Cursor

Once configured, the AI assistant can use dependency-buster tools:

```
@dependency-buster analyze_dependencies for this project
@dependency-buster audit_security to check for vulnerabilities
@dependency-buster analyze_licenses to verify compliance
```

---

## VSCode

### Using MCP Extension

1. Install the **MCP Extension** from VSCode Marketplace
2. Open Settings (JSON) and add:

```json
{
  "mcp.servers": {
    "dependency-buster": {
      "command": "node",
      "args": ["/path/to/dpb-mcp-workspace/dpb-mcp-typescript/build/server.js"]
    }
  }
}
```

### Using Tasks (Alternative)

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Analyze Dependencies",
      "type": "shell",
      "command": "echo '{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"analyze_dependencies\",\"arguments\":{\"repo_path\":\"${workspaceFolder}\"}},\"id\":1}' | node /path/to/dpb-mcp-workspace/dpb-mcp-typescript/build/server.js",
      "problemMatcher": []
    },
    {
      "label": "Security Audit",
      "type": "shell",
      "command": "echo '{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"audit_security\",\"arguments\":{\"repo_path\":\"${workspaceFolder}\"}},\"id\":1}' | node /path/to/dpb-mcp-workspace/dpb-mcp-typescript/build/server.js",
      "problemMatcher": []
    }
  ]
}
```

---

## Cline Extension

### Configuration

Add to your Cline MCP settings (`.cline/mcp_settings.json`):

```json
{
  "mcpServers": {
    "dependency-buster": {
      "command": "node",
      "args": [
        "/path/to/dpb-mcp-workspace/dpb-mcp-typescript/build/server.js"
      ],
      "disabled": false,
      "autoApprove": ["analyze_dependencies", "detect_namespaces", "analyze_licenses"]
    }
  }
}
```

### Auto-Approve Tools

For security auditing, you may want manual approval:

```json
{
  "autoApprove": ["analyze_dependencies", "detect_namespaces"],
  "alwaysAllow": false
}
```

### Using in Cline

```
Use the dependency-buster MCP to analyze this project's dependencies
Check for security vulnerabilities using dependency-buster
Generate a license report with dependency-buster
```

---

## Claude Code CLI

### Installation

```bash
# Install Claude Code CLI (if not already installed)
npm install -g @anthropic-ai/claude-code

# Or use npx
npx @anthropic-ai/claude-code
```

### Configuration

Add to `~/.claude/mcp_servers.json`:

```json
{
  "dependency-buster": {
    "command": "node",
    "args": ["/path/to/dpb-mcp-workspace/dpb-mcp-typescript/build/server.js"],
    "env": {
      "MCP_TRANSPORT": "stdio"
    }
  }
}
```

### Using in Claude Code

```bash
# Start Claude Code in your project directory
claude

# Then ask Claude to analyze dependencies
> Use dependency-buster to analyze this project

# Or run specific tools
> @dependency-buster audit_security
> @dependency-buster analyze_dependencies
```

### Programmatic Usage

```typescript
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

const transport = new StdioClientTransport({
  command: "node",
  args: ["/path/to/dpb-mcp-workspace/dpb-mcp-typescript/build/server.js"]
});

const client = new Client({ name: "my-app", version: "1.0.0" }, {});
await client.connect(transport);

// List available tools
const tools = await client.listTools();
console.log(tools);

// Analyze dependencies
const result = await client.callTool({
  name: "analyze_dependencies",
  arguments: { repo_path: "/path/to/your/project" }
});
console.log(result);
```

---

## Available Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `analyze_dependencies` | Comprehensive dependency analysis | `repo_path` (string) |
| `analyze_psr4` | PSR-4 autoloading analysis (PHP) | `repo_path` (string) |
| `detect_namespaces` | Namespace detection and mapping | `repo_path` (string) |
| `audit_security` | Security vulnerability audit | `repo_path` (string) |
| `analyze_licenses` | License compliance analysis | `repo_path` (string) |

### Tool Annotations (Enterprise)

All tools include MCP annotations:

```json
{
  "readOnlyHint": true,
  "idempotentHint": true,
  "cacheTtlSeconds": 300,
  "tags": ["analysis", "dependencies"]
}
```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MCP_TRANSPORT` | Transport mode (`stdio` or `http`) | `stdio` |
| `MCP_AUTH_ENABLED` | Enable authentication | `false` |
| `MCP_TOKEN` | Static auth token (if auth enabled) | - |
| `MCP_HTTP_PORT` | HTTP server port (if http transport) | `3000` |

### HTTP Transport (for remote access)

```bash
# Start server in HTTP mode
MCP_TRANSPORT=http MCP_HTTP_PORT=3000 node build/server.js

# Or with Go
MCP_TRANSPORT=http MCP_HTTP_PORT=3000 ./build/dpb-mcp
```

### Authentication

```bash
# Enable authentication
MCP_AUTH_ENABLED=true MCP_TOKEN=your-secret-token node build/server.js

# Client must include Authorization header
curl -H "Authorization: Bearer your-secret-token" http://localhost:3000/
```

---

## Troubleshooting

### Server Not Starting

```bash
# Check if the server binary exists
ls -la /path/to/dpb-mcp-workspace/dpb-mcp-typescript/build/server.js

# Check for Node.js
node --version

# Check for permissions
chmod +x /path/to/binary
```

### Tools Not Appearing in IDE

1. Verify the MCP server is running
2. Check IDE MCP configuration syntax
3. Restart the IDE
4. Check IDE logs for MCP errors

### Permission Denied

```bash
# Make binary executable
chmod +x ./build/dpb-mcp
chmod +x ./target/release/dpb-mcp
```

### JSON Parse Errors

Ensure the server is outputting valid JSON-RPC:

```bash
echo '{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}' | node build/server.js 2>/dev/null | head -1 | jq .
```

---

## Examples

### Analyze a Project

```bash
# Using the benchmark script
./scripts/run-benchmark.sh /path/to/your/project

# Using the MCP server directly
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"analyze_dependencies","arguments":{"repo_path":"/path/to/project"}},"id":1}' | node build/server.js
```

### Generate Dashboard

```bash
# Run the full benchmark suite
cd dpb-mcp-workspace/dpb-benchmark
./scripts/run-benchmark.sh /path/to/your/project

# Open the dashboard
open dashboard/index.html
```

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup and contribution guidelines.

## License

MIT License - See [LICENSE](./LICENSE) for details.
