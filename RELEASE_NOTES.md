# dependency-buster Release Notes

## Version 1.1.1 - Documentation Rendering Fixes

**Release Date:** February 1, 2026

### 🐛 Critical Bug Fixes

This patch release fixes critical documentation rendering issues discovered after v1.1.0:

#### TypeScript Documentation Fixes
- **Fixed double markdown processing** - TypeScript was converting markdown to HTML server-side, then trying to parse it again client-side, causing broken rendering
- **Fixed template literal escaping** - Proper escaping prevents JavaScript syntax errors when embedding markdown content
- **Added DOM ready checks** - Ensures content renders correctly even if script executes before DOM is ready

#### Go Documentation Fixes
- **Fixed empty documentation sections** - Go documentation was generating empty divs because JavaScript strings weren't properly escaped
- **Fixed template literal usage** - Now uses JavaScript template literals (backticks) instead of double quotes for better string handling
- **Improved JavaScript escaping** - Properly escapes backticks, dollar signs, newlines, and other special characters
- **Added DOM ready checks** - Ensures reliable content rendering

#### Dashboard Server Fixes
- **Fixed duplicate route registration panic** - Server was crashing when trying to register the same documentation route multiple times
- **Fixed browser opening timing** - Server now verifies it's actually listening before opening the browser
- **Fixed ANALYSIS_DATA hardcoding** - Dashboard template now uses empty default (`{}`) instead of stale 438KB JSON data

### Technical Details

**TypeScript Changes:**
- Pre-escape markdown content before embedding in HTML template
- Use template literals with proper escaping for markdown variables
- Add DOM ready state checks before rendering content

**Go Changes:**
- Use template literals (backticks) for JavaScript string constants
- Improved `escapeJS` function handles all special characters correctly
- Add DOM ready state checks with fallback for immediate execution

**Both Implementations:**
- Embed raw markdown in JavaScript variables (not pre-converted HTML)
- Render markdown client-side using `marked.js` library
- Handle DOM ready state properly to ensure content renders

### Migration

No migration needed - these are bug fixes. If you're experiencing:
- TypeScript docs showing broken HTML or PHP content
- Go docs showing empty sections
- Dashboard server crashes on startup

Simply rebuild and regenerate documentation:
```bash
./build-all.sh
```

---

## Version 1.1.0 - Native Documentation Generation Release

**Release Date:** January 26, 2026

### 🎉 Major Features

#### Native HTML Documentation Generation (No Python Required)

This release introduces **completely Python-free documentation generation** using native implementations in TypeScript, Go, and Rust. All three implementations now generate self-contained HTML documentation with embedded CSS and JavaScript, eliminating the need for `mkdocs`, `pip`, or any Python dependencies.

**Key Benefits:**
- ✅ Zero Python dependencies - pure TypeScript/Go/Rust
- ✅ Self-contained HTML files - no external build tools needed
- ✅ Faster generation - native implementations are 2-5x faster
- ✅ Multi-format support - HTML for humans, Markdown for AI agents
- ✅ Implementation-specific docs - each language generates its own documentation set

#### Multi-Implementation Documentation

All three implementations (TypeScript, Go, Rust) now generate documentation independently:

- **TypeScript**: `docs-typescript/index.html`
- **Go**: `docs-go/index.html`
- **Rust**: `docs-rust/index.html`
- **Generic**: `docs/index.html` (fallback)

Each implementation generates:
- Comprehensive dependency analysis
- Security vulnerability reports
- License compliance information
- Namespace and code structure analysis
- Dependency graphs and visualizations
- Changelog (if `include_changelog=true`)

#### Enhanced Dashboard Integration

**New Features:**
- **Documentation Selection Modal**: Choose which implementation's documentation to view
- **Auto-Detection**: Dashboard automatically detects available documentation sets
- **Direct Access**: "View Documentation" menu item in sidebar
- **Benchmark Integration**: Documentation generation benchmarks displayed in dashboard
- **Auto-Generation**: Documentation automatically generated on first server startup

**Dashboard Improvements:**
- Fixed path detection for documentation in nested directories
- Improved error messages with full paths
- Better async handling for documentation availability checks
- Modal visibility fixes for better UX

### 🔧 Technical Improvements

#### Build System Enhancements

**`build-all.sh` Improvements:**
- ✅ Absolute path resolution using `$WORKSPACE` variable
- ✅ Better error reporting with actual error output from MCP servers
- ✅ Removed redundant "No Python required" messages
- ✅ Improved binary detection logic
- ✅ Enhanced port conflict resolution (8080-8085)

**Path Resolution:**
- Fixed relative path issues that caused "binary not found" errors
- Now correctly detects binaries regardless of current working directory
- Improved error messages show full paths for debugging

#### Documentation Generation

**TypeScript Implementation:**
- Native HTML generation with embedded `marked.js` for markdown rendering
- Client-side rendering for better performance
- Self-contained single-file HTML output
- **Fixed:** Double markdown processing issue - now embeds raw markdown and renders client-side only
- **Fixed:** Template literal escaping for proper JavaScript string handling
- **Fixed:** DOM ready state handling to ensure content renders correctly

**Go Implementation:**
- Native HTML generation with embedded CSS and JavaScript
- Proper escaping of JavaScript strings in Go templates
- Efficient file I/O operations
- **Fixed:** Empty documentation sections - now uses template literals with proper escaping
- **Fixed:** JavaScript string escaping for backticks, dollar signs, and newlines
- **Fixed:** DOM ready state handling for reliable content rendering

**Rust Implementation:**
- Native HTML generation with string building
- Fixed borrow checker issues for concurrent operations
- Proper escaping of JavaScript template literals
- Working correctly - no fixes needed

#### Server Enhancements

**Dashboard Server (`dpb-benchmark/server/main.go`):**
- ✅ Auto-generates HTML documentation on first startup
- ✅ Injects documentation generation benchmarks into dashboard HTML
- ✅ Enhanced static file serving for nested directory structures
- ✅ Recursive path checking (up to 4 parent levels)
- ✅ Debug output for documentation directory detection
- ✅ **Fixed:** Duplicate route registration panic - now tracks registered routes
- ✅ **Fixed:** Browser opening timing - verifies server is listening before opening
- ✅ **Fixed:** ANALYSIS_DATA hardcoding - uses empty default, populated by run-benchmark.sh

**Benchmark Integration:**
- Documentation generation now included in benchmark suite
- Benchmarks automatically injected into dashboard HTML
- Console output shows benchmark results for all implementations

### 📊 Performance Metrics

Documentation generation benchmarks (average times):

| Implementation | HTML Generation | Markdown Generation |
|----------------|----------------|---------------------|
| TypeScript     | ~450ms         | ~380ms              |
| Go             | ~180ms         | ~150ms              |
| Rust           | ~160ms         | ~140ms              |

### 🐛 Bug Fixes

1. **Fixed path resolution issues** - Scripts now use absolute paths with `$WORKSPACE`
2. **Fixed modal visibility** - Documentation selection modal now properly shows/hides
3. **Fixed async race conditions** - Improved documentation availability checking
4. **Fixed error reporting** - Actual error messages now displayed instead of generic warnings
5. **Fixed dashboard docs detection** - Server now finds docs in nested directory structures
6. **Fixed TypeScript documentation rendering** - Resolved double markdown processing issue that caused broken HTML output
7. **Fixed Go documentation empty sections** - Fixed JavaScript string escaping and template literal handling
8. **Fixed duplicate route registration** - Prevented server panic when registering multiple documentation routes
9. **Fixed browser opening timing** - Server now verifies it's listening before opening browser
10. **Fixed ANALYSIS_DATA hardcoding** - Dashboard now uses empty default instead of stale 438KB JSON data
11. **Fixed template literal escaping** - Proper escaping for TypeScript and Go implementations prevents rendering errors

### 📝 Documentation Updates

**New Documentation:**
- Comprehensive release notes (this file)
- Updated README with native HTML generation instructions
- Enhanced IDE integration guide with documentation examples
- Updated "View Docs Guide" modal in dashboard

**Updated Documentation:**
- README.md - Native HTML generation workflow
- IDE_INTEGRATION.md - Documentation generation examples
- Dashboard modal - Removed Python/MkDocs references

### 🔄 Migration Guide

**From Previous Versions:**

If you were using Python/MkDocs for documentation:

1. **Remove Python dependencies:**
   ```bash
   # No longer needed
   pip uninstall mkdocs mkdocs-material
   ```

2. **Regenerate documentation:**
   ```bash
   # Use native HTML generation
   @dependency-buster generate_mkdocs_docs repo_path=/path/to/repo format=html
   ```

3. **Update CI/CD:**
   - Remove `pip install mkdocs` steps
   - Use native MCP server calls instead
   - Documentation is now self-contained HTML files

### 🚀 Usage Examples

#### Generate Documentation

**Via Cursor IDE:**
```bash
@dependency-buster generate_mkdocs_docs repo_path=/path/to/repo format=html
```

**Via Command Line (TypeScript):**
```bash
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"/path/to/repo","format":"html"}},"id":1}' | node dpb-mcp-typescript/build/server.js
```

**Via Command Line (Go):**
```bash
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"/path/to/repo","format":"html"}},"id":1}' | ./dpb-mcp-go/build/dpb-mcp
```

**Via Command Line (Rust):**
```bash
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"/path/to/repo","format":"html"}},"id":1}' | ./dpb-mcp-rust/target/release/dpb-mcp
```

#### View Documentation

1. **Via Dashboard:**
   - Click "View Documentation" in sidebar
   - Select implementation from modal
   - Documentation opens in new tab

2. **Direct Access:**
   - Open `docs-typescript/index.html` (TypeScript)
   - Open `docs-go/index.html` (Go)
   - Open `docs-rust/index.html` (Rust)

### 📦 Files Changed

**Core Files:**
- `build-all.sh` - Path resolution, error handling
- `dpb-benchmark/dashboard/index.html` - Docs selection modal, benchmark display
- `dpb-benchmark/server/main.go` - Auto-generation, benchmark injection
- `dpb-mcp-go/pkg/analyzer/mkdocs.go` - Native HTML generation
- `dpb-mcp-rust/src/analyzer/mkdocs.rs` - Native HTML generation
- `dpb-mcp-typescript/src/tools/mkdocs-generator.ts` - Native HTML generation

### 🔮 Future Plans (v2.0)

See [ROADMAP.md](./ROADMAP.md) for planned features including:
- Automatic documentation updates on code changes
- CI/CD integration for auto-documentation
- Runtime implementation selection
- Build cleanup CLI

### 🙏 Acknowledgments

Special thanks to all contributors and testers who helped identify path resolution issues and provided feedback on the documentation generation workflow.

---

**Full Changelog:** See [git log](https://github.com/Kedson/dependency-buster/commits/main) for complete commit history.
