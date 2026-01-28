# Viewing Generated Documentation

The `generate_mkdocs_docs` tool creates comprehensive documentation from your dependency analysis data. Here's how to view it:

## Generated Files

When you run `generate_mkdocs_docs`, it creates the following structure in the `docs/` directory (or your specified `output_dir`):

```
docs/
├── index.md              # Main documentation page
├── dependencies.md        # Dependency analysis and tree
├── security.md           # Security audit results
├── licenses.md           # License compliance report
├── architecture.md       # PSR-4 and namespace structure
├── changelog.md          # Dependency change history (if enabled)
└── mkdocs.yml            # MkDocs configuration (if format=mkdocs)
```

## Viewing Options

### Option 1: MkDocs (Recommended)

If you generated with `format=mkdocs` (default), you can view a beautiful web interface:

```bash
# Install MkDocs (if not already installed)
pip install mkdocs mkdocs-material

# Navigate to docs directory
cd docs

# Start local server
mkdocs serve

# Open in browser
# http://127.0.0.1:8000
```

**Features:**
- Beautiful Material Design theme
- Search functionality
- Responsive navigation
- Syntax highlighting
- Mermaid diagram rendering

### Option 2: View Markdown Directly

You can view the markdown files directly in:
- Your IDE/editor (VS Code, Cursor, etc.)
- GitHub/GitLab (if committed)
- Any markdown viewer

```bash
# View in terminal
cat docs/index.md

# Or open in your editor
code docs/index.md
```

### Option 3: HTML Output

If you generated with `format=html`, a standalone HTML file is created:

```bash
# Open in browser
open docs/index.html
```

## Smoke Tests & Benchmarking

### Smoke Tests

**Smoke tests do NOT regenerate documentation automatically.** They test the MCP server functionality, not documentation generation.

To generate docs during smoke tests, you would need to explicitly call `generate_mkdocs_docs` as one of the test tools.

### Benchmarking

**Documentation generation is NOT currently benchmarked** in the performance tests. The benchmark focuses on:
- Analysis tools (dependency analysis, security audit, etc.)
- Performance metrics (startup time, memory usage, execution time)

To add documentation generation to benchmarks, you would need to:
1. Add `generate_mkdocs_docs` to the benchmark tool list in `run-benchmark.sh`
2. Add it to the dashboard's tool list

## Example Usage

### Via Cursor IDE

```typescript
// In Cursor chat
@dependency-buster generate_mkdocs_docs

// With options
@dependency-buster generate_mkdocs_docs repo_path=/path/to/repo output_dir=./docs include_changelog=true format=mkdocs
```

### Via Command Line (TypeScript)

```bash
# Using the test script
cd dpb-mcp-workspace
node dpb-mcp-typescript/build/server.js <<EOF
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"/path/to/repo","output_dir":"./docs"}},"id":1}
EOF
```

### Via Command Line (Go)

```bash
./dpb-mcp-go/build/dpb-mcp <<EOF
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"/path/to/repo","output_dir":"./docs"}},"id":1}
EOF
```

### Via Command Line (Rust)

```bash
./dpb-mcp-rust/target/release/dpb-mcp <<EOF
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"/path/to/repo","output_dir":"./docs"}},"id":1}
EOF
```

## Documentation Content

All three implementations (TypeScript, Go, Rust) generate documentation from **real analysis data**:

- **Dependencies**: Production/dev counts, package lists, dependency graph
- **Security**: Risk level, vulnerability counts, detailed vulnerability list
- **Licenses**: Distribution, compatibility issues, summary statistics
- **Architecture**: PSR-4 mappings, namespace detection, file counts
- **Changelog**: Added/updated/removed dependencies (if snapshots exist)

## Troubleshooting

### MkDocs not rendering diagrams

Install the Mermaid plugin:
```bash
pip install mkdocs-mermaid2-plugin
```

Then add to `mkdocs.yml`:
```yaml
plugins:
  - mermaid2
```

### Documentation seems empty

Make sure you're running the tool on a repository that has:
- `composer.json` (for PHP projects)
- Dependency files for your language
- Analysis data available

### Permission errors

Ensure the output directory is writable:
```bash
chmod -R 755 docs/
```

## Next Steps

- Customize `mkdocs.yml` for your project branding
- Add custom pages to the documentation
- Integrate with CI/CD to auto-generate docs on commits
- Host on GitHub Pages using `mkdocs gh-deploy`
