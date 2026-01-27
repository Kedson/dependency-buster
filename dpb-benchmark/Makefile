# Dashboard Server Makefile

.PHONY: serve dev build clean

# Default: run server
serve: build
	@./server/dashboard-server

# Development with hot-reload (requires: go install github.com/cosmtrek/air@latest)
dev:
	@cd server && air

# Build the server binary
build:
	@echo "Building dashboard server..."
	@cd server && go build -o dashboard-server .
	@echo "✓ Built: server/dashboard-server"

# Clean build artifacts
clean:
	@rm -f server/dashboard-server
	@rm -rf server/tmp
	@echo "✓ Cleaned"

# Install Air for hot-reload
install-air:
	@go install github.com/cosmtrek/air@latest
	@echo "✓ Air installed"
