# Changelog

All notable changes to dependency-buster will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2026-02-01

### Fixed
- TypeScript HTML documentation double markdown processing issue
- Go HTML documentation empty sections (JavaScript escaping fixes)
- Duplicate route registration panic in dashboard server
- Browser opening before server is ready
- ANALYSIS_DATA hardcoding in dashboard (now uses empty default)
- Template literal escaping in TypeScript and Go implementations
- DOM ready state handling for reliable content rendering

### Changed
- Improved JavaScript string escaping for template literals
- Enhanced DOM ready state checks in both TypeScript and Go
- Better error handling for documentation generation

## [1.1.0] - 2026-01-26

### Added
- Native HTML documentation generation (no Python required)
- Multi-implementation documentation generation (TypeScript, Go, Rust)
- Documentation selection modal in dashboard
- Auto-generation of documentation on first server startup
- Documentation generation benchmarks in dashboard
- Enhanced error reporting with actual error messages
- Recursive path detection for documentation directories (up to 4 levels)
- "View Documentation" menu item in dashboard sidebar
- Benchmark injection into dashboard HTML

### Changed
- Documentation generation now uses native implementations instead of Python/MkDocs
- Build scripts use absolute paths with `$WORKSPACE` variable
- Improved path resolution for binaries and documentation
- Enhanced dashboard documentation detection logic
- Updated "View Docs Guide" modal with native HTML instructions

### Fixed
- Path resolution issues causing "binary not found" errors
- Modal visibility issues in documentation selection
- Async race conditions in documentation availability checking
- Dashboard not finding documentation in nested directories
- Error messages now show actual errors instead of generic warnings
- Documentation benchmarks not appearing in dashboard on first run
- **TypeScript documentation rendering** - Fixed double markdown processing causing broken HTML output
- **Go documentation empty sections** - Fixed JavaScript string escaping and template literal handling
- **Duplicate route registration panic** - Prevented server crash when registering multiple documentation routes
- **Browser opening timing** - Server now verifies it's listening before opening browser
- **ANALYSIS_DATA hardcoding** - Dashboard uses empty default instead of stale 438KB JSON data
- **Template literal escaping** - Proper escaping for TypeScript and Go prevents rendering errors
- **DOM ready state handling** - Both TypeScript and Go now wait for DOM before rendering content

### Performance
- Documentation generation is 2-5x faster with native implementations
- HTML generation: TypeScript ~450ms, Go ~180ms, Rust ~160ms
- Markdown generation: TypeScript ~380ms, Go ~150ms, Rust ~140ms

## [1.0.0] - 2025-12-XX

### Added
- Initial release with TypeScript, Go, and Rust implementations
- Multi-language dependency analysis
- Security vulnerability scanning
- License compliance checking
- Dependency timestamp tracking
- Agent suggestion hooks (Cursor, Cline, Claude Code)
- Beautiful dashboard visualization
- ASCII terminal reports
- MkDocs documentation generation (Python-based)

---

## Version History

- **1.1.0** - Native HTML documentation generation, multi-implementation docs, dashboard improvements
- **1.0.0** - Initial release with core dependency analysis features
