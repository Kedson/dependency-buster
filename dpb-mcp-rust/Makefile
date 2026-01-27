.PHONY: build release install clean test run bench size

# Development build
build:
	@echo "Building debug version..."
	cargo build

# Release build (optimized)
release:
	@echo "Building release version with optimizations..."
	cargo build --release
	@echo "Binary size:"
	@ls -lh target/release/php-dependency-mcp

# Super optimized build
release-small:
	@echo "Building super-optimized version..."
	cargo build --profile release-small
	strip target/release-small/php-dependency-mcp 2>/dev/null || true
	@echo "Binary size:"
	@ls -lh target/release-small/php-dependency-mcp

# Install to system
install: release
	@echo "Installing to /usr/local/bin..."
	@sudo cp target/release/php-dependency-mcp /usr/local/bin/
	@echo "Installed successfully!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	cargo clean

# Run tests
test:
	cargo test --verbose

# Run the server (debug mode)
run:
	cargo run

# Run benchmarks
bench:
	cargo bench

# Check code without building
check:
	cargo check

# Format code
fmt:
	cargo fmt

# Lint code
lint:
	cargo clippy

# Show binary sizes
size: release
	@echo "\nRelease binary analysis:"
	@size target/release/php-dependency-mcp 2>/dev/null || ls -lh target/release/php-dependency-mcp

# Cross-compile for different platforms
cross-linux:
	cargo build --release --target x86_64-unknown-linux-gnu

cross-windows:
	cargo build --release --target x86_64-pc-windows-gnu

cross-macos-arm:
	cargo build --release --target aarch64-apple-darwin

# Update dependencies
update:
	cargo update

# Generate documentation
doc:
	cargo doc --open
