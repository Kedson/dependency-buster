# dependency-buster Roadmap

## v1.0.0 (Current)
- ✅ Multi-language dependency analysis
- ✅ Security vulnerability scanning
- ✅ License compliance checking
- ✅ Dependency timestamp tracking
- ✅ Agent suggestion hooks (Cursor, Cline, Claude Code)
- ✅ Beautiful dashboard visualization
- ✅ ASCII terminal reports

## v2.0.0 (Planned)

### Runtime Implementation Selection
- [ ] Choose TypeScript, Go, or Rust at runtime
- [ ] Auto-detect best implementation for system
- [ ] Fallback chain if preferred not available

### Build Cleanup CLI
- [ ] `dpb clean` - Remove build artifacts
- [ ] `dpb clean --deps` - Remove dependency caches
- [ ] `dpb clean --all` - Full cleanup

### Enhanced Cleanup Features
- [ ] Parallel cleanup for faster execution
- [ ] `--dry-run` mode to preview deletions
- [ ] Interactive delete confirmation
- [ ] Summary of freed disk space
- [ ] Selective cleanup by age/size

### Additional Features
- [ ] npm/yarn/pnpm support for JS projects
- [ ] pip/poetry support for Python projects
- [ ] cargo support for Rust projects
- [ ] go.mod support for Go projects
- [ ] Vulnerability database integration
- [ ] Dependency update suggestions

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for how to help with these features.
