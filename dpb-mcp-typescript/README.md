# PHP Dependency MCP Server

A comprehensive Model Context Protocol (MCP) server for analyzing PHP projects, validating PSR-4 compliance, auditing security, and managing dependencies across multiple repositories.

## Features

### Single Repository Analysis
- **Dependency Analysis**: Full dependency tree with production and dev packages
- **PSR-4 Validation**: Validate autoloading configuration and namespace compliance
- **Namespace Detection**: Discover all namespaces and their usage patterns
- **Security Auditing**: Check for outdated packages and potential vulnerabilities
- **License Analysis**: Analyze license distribution and compatibility
- **Dependency Graphs**: Generate Mermaid diagrams of dependency relationships
- **Circular Dependencies**: Detect circular dependency chains

### Multi-Repository Analysis
- **Cross-Repo Analysis**: Analyze dependencies across multiple repositories
- **Shared Dependencies**: Identify packages used by multiple projects
- **Version Conflicts**: Find version mismatches across repositories
- **Consolidated Reports**: Generate comprehensive multi-repo reports

## Installation

### Prerequisites
- Node.js 18+ 
- npm or yarn
- PHP projects with `composer.json`

### Quick Install

```bash
# Clone or extract the package
cd dpb-mcp-complete

# Install dependencies
npm install

# Build the project
npm run build

# Install globally (optional)
npm install -g .

# Or link for development
npm link
```

## Configuration

### For Claude Code

Add to `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "php-dependency-analyzer": {
      "type": "stdio",
      "command": "dpb-mcp"
    }
  }
}
```

Or use the CLI:

```bash
claude mcp add php-analyzer --scope user -- dpb-mcp
```

### For Cursor

Create `.cursor/mcp.json` in your project root:

```json
{
  "mcpServers": {
    "php-analyzer": {
      "command": "dpb-mcp"
    }
  }
}
```

Or add to global Cursor config:
- macOS: `~/Library/Application Support/Cursor/mcp.json`
- Windows: `%APPDATA%\Cursor\mcp.json`
- Linux: `~/.config/Cursor/mcp.json`

### For VSCode with Cline

Add to VSCode `settings.json`:

```json
{
  "cline.mcpServers": {
    "php-dependency-analyzer": {
      "command": "dpb-mcp"
    }
  }
}
```

## Usage

### Single Repository Analysis

#### In Claude Code

```bash
cd /path/to/your/php/project
claude .
```

Then ask:

```
Analyze this PHP repository:
1. Check dependencies
2. Validate PSR-4 compliance
3. Run security audit
4. Generate dependency graph
5. Create comprehensive documentation
```

#### In Cursor

Open Composer (Cmd/Ctrl + K), switch to Agent mode, and ask:

```
Use php-analyzer to:
- Analyze dependencies in this repo
- Check PSR-4 compliance
- Generate dependency graph focusing on Symfony packages
```

### Multi-Repository Analysis

1. Create a config file (see `config/example-repos.example.json`):

```json
[
  {
    "name": "my-service",
    "path": "/absolute/path/to/service",
    "type": "service",
    "team": "Backend Team",
    "description": "Main API service"
  }
]
```

2. Ask Claude/Cursor:

```
Analyze all repositories in config/my-repos.json:
- Find shared dependencies
- Identify version conflicts
- Generate consolidated report
```

## Available Tools

### `analyze_dependencies`
Analyzes composer.json and composer.lock for production and dev dependencies.

**Input:**
- `repo_path`: Absolute path to PHP repository

**Output:**
- Production dependencies
- Development dependencies
- Dependency tree with relationships
- Statistics

### `analyze_psr4`
Validates PSR-4 autoloading configuration and namespace compliance.

**Input:**
- `repo_path`: Absolute path to PHP repository

**Output:**
- PSR-4 mappings
- Namespace violations
- Compliance statistics

### `detect_namespaces`
Detects all namespaces used in the codebase.

**Input:**
- `repo_path`: Absolute path to PHP repository

**Output:**
- All namespaces with their files, classes, interfaces, and traits
- Files without namespaces

### `analyze_namespace_usage`
Analyzes usage of a specific namespace across the codebase.

**Input:**
- `repo_path`: Absolute path to PHP repository
- `namespace`: Target namespace to analyze

**Output:**
- Files defining the namespace
- Files importing from the namespace
- Total usage count

### `generate_dependency_graph`
Generates a Mermaid diagram of dependency relationships.

**Input:**
- `repo_path`: Absolute path to PHP repository
- `max_depth` (optional): Maximum depth for tree (default: 2)
- `include_dev` (optional): Include dev dependencies
- `focus_package` (optional): Focus on specific package

**Output:**
- Mermaid diagram syntax

### `audit_security`
Audits dependencies for security issues and outdated packages.

**Input:**
- `repo_path`: Absolute path to PHP repository

**Output:**
- List of vulnerabilities
- Risk level (low/medium/high/critical)
- Summary by severity

### `analyze_licenses`
Analyzes license distribution and compatibility.

**Input:**
- `repo_path`: Absolute path to PHP repository

**Output:**
- License distribution with risk levels
- Compatibility issues
- Summary statistics

### `find_circular_dependencies`
Finds circular dependency chains.

**Input:**
- `repo_path`: Absolute path to PHP repository

**Output:**
- Array of circular dependency chains
- Count of cycles

### `analyze_multi_repo`
Analyzes dependencies across multiple repositories.

**Input:**
- `config_path`: Path to repository configuration JSON

**Output:**
- Consolidated multi-repo analysis report

### `generate_comprehensive_docs`
Generates comprehensive markdown documentation.

**Input:**
- `repo_path`: Absolute path to PHP repository
- `output_path` (optional): Where to save the file

**Output:**
- Markdown documentation or file path

## Testing with AzuraCast

Test the MCP server with the feature-rich AzuraCast project:

```bash
# Clone AzuraCast
git clone https://github.com/AzuraCast/AzuraCast.git ~/test/azuracast

# Navigate to it
cd ~/test/azuracast

# Start Claude Code
claude .

# Or open in Cursor
cursor .
```

Then ask:

```
Analyze the AzuraCast repository comprehensively:
1. What are the main dependencies?
2. Is the PSR-4 autoloading properly configured?
3. Are there any security concerns?
4. What licenses are used?
5. Generate a dependency graph
6. Save documentation to DEPENDENCIES.md
```

## Development

### Project Structure

```
dpb-mcp-complete/
├── src/
│   ├── server.ts              # Main MCP server
│   ├── tools/                 # Analysis tools
│   │   ├── dependency-analyzer.ts
│   │   ├── psr4-analyzer.ts
│   │   ├── namespace-detector.ts
│   │   ├── security-auditor.ts
│   │   ├── license-analyzer.ts
│   │   ├── multi-repo-analyzer.ts
│   │   └── graph-generator.ts
│   ├── utils/                 # Utilities
│   │   ├── composer-utils.ts
│   │   ├── php-parser.ts
│   │   └── file-scanner.ts
│   └── types/                 # TypeScript types
│       └── index.ts
├── config/                    # Configuration examples
├── package.json
├── tsconfig.json
└── README.md
```

### Building

```bash
npm run build        # Build TypeScript
npm run dev          # Watch mode
npm run clean        # Clean build directory
```

## Troubleshooting

### MCP Server Not Connecting

1. Check that it's properly installed:
```bash
which dpb-mcp
```

2. Test manually:
```bash
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | dpb-mcp
```

3. Check Claude Code status:
```bash
claude mcp list
```

### Cursor Not Finding Tools

1. Restart Cursor completely
2. Verify config in Cursor Settings → MCP
3. Check that you're in Agent mode (not Ask mode)

### Path Issues

- Always use absolute paths for `repo_path`
- On Windows, use forward slashes or escaped backslashes
- Ensure the path contains a valid `composer.json`

## Dependency Buster Platform Integration

For multi-repository analysis across the Dependency Buster platform:

1. Copy `config/example-repos.example.json` to `config/example-repos.json`
2. Update paths to match your local setup
3. Run multi-repo analysis:

```
Analyze the Dependency Buster platform using config/example-repos.json:
- Find all shared dependencies
- Identify version conflicts that need resolution
- Recommend a dependency consolidation strategy
```

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.

## Author

Dependency Buster Contributors Team

## Version

1.0.0
