# Quick Start Guide

Get up and running in 5 minutes!

## 1. Extract and Setup

```bash
# Extract the zip file
unzip php-dependency-mcp-complete.zip
cd php-dependency-mcp-complete

# Run setup script (recommended)
./setup.sh

# OR do it manually:
npm install
npm run build
npm link
```

## 2. Configure Your IDE

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

## 3. Test with AzuraCast

```bash
# Clone test repository
git clone https://github.com/AzuraCast/AzuraCast.git ~/test/azuracast
cd ~/test/azuracast

# Start Claude Code
claude .
```

Then ask:
```
Analyze this PHP repository:
- What are the main dependencies?
- Check PSR-4 compliance
- Generate dependency graph
- Run security audit
```

## 4. Use with Your Project

```bash
cd /path/to/your/php/project
claude .  # or open in Cursor
```

Then ask:
```
Use php-dependency-analyzer to:
1. Analyze all dependencies
2. Validate PSR-4 autoloading
3. Check for security issues
4. Generate documentation in DEPENDENCIES.md
```

## 5. Multi-Repo Analysis (Faith FM)

```bash
# Copy config template
cp config/faith-fm-repos.example.json config/faith-fm-repos.json

# Edit paths
nano config/faith-fm-repos.json
```

Then ask:
```
Analyze all repositories in config/faith-fm-repos.json:
- Find shared dependencies
- Identify version conflicts
- Generate consolidated report
```

## Available Commands

### Dependency Analysis
```
Analyze dependencies in this repository
```

### PSR-4 Validation
```
Check PSR-4 autoloading compliance and show violations
```

### Security Audit
```
Run security audit and show vulnerabilities
```

### Namespace Detection
```
Detect all namespaces and their usage
```

### Dependency Graph
```
Generate dependency graph focusing on Symfony packages
```

### Comprehensive Documentation
```
Generate comprehensive docs and save to DEPENDENCIES.md
```

## Troubleshooting

### Server not found
```bash
# Check installation
which php-dependency-mcp

# If not found, try:
npm link
```

### Cursor not detecting tools
1. Restart Cursor completely
2. Check Settings → MCP
3. Switch to Agent mode (not Ask mode)

### Claude Code connection issues
```bash
# Check status
claude mcp list

# Re-add server
claude mcp remove php-analyzer
claude mcp add php-analyzer --scope user -- php-dependency-mcp
```

## Next Steps

- Read full [README.md](README.md) for all features
- Check `examples/` folder for more usage patterns
- Configure Faith FM multi-repo analysis
- Integrate with CI/CD for automated audits

## Support

For issues or questions, check the README.md or create an issue in the repository.
