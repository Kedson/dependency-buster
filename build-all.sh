#!/bin/bash
# Note: Not using 'set -e' so script continues even if some builds fail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

clear
echo -e "${PURPLE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║                                                            ║${NC}"
echo -e "${PURPLE}║     dependency-buster // Universal Dependency Analyzer    ║${NC}"
echo -e "${PURPLE}║     Build All → Test All → Generate Report → View         ║${NC}"
echo -e "${PURPLE}║                                                            ║${NC}"
echo -e "${PURPLE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Get the directory where this script is located (the workspace)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
WORKSPACE="$SCRIPT_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo -e "${BLUE}📁 Workspace: $WORKSPACE${NC}"
echo -e "${BLUE}🕐 Started: $(date)${NC}"
echo ""

# Check prerequisites
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 1: Checking Prerequisites${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

check_command() {
    if command -v $1 &> /dev/null; then
        VERSION=$($1 --version 2>&1 | head -n1)
        echo -e "${GREEN}✓${NC} $1: $VERSION"
        return 0
    else
        echo -e "${RED}✗${NC} $1: Not found"
        return 1
    fi
}

MISSING=0
check_command node || MISSING=$((MISSING + 1))
check_command npm || MISSING=$((MISSING + 1))
check_command go || MISSING=$((MISSING + 1))
check_command cargo || MISSING=$((MISSING + 1))
check_command git || MISSING=$((MISSING + 1))

echo ""

if [ $MISSING -gt 0 ]; then
    echo -e "${RED}⚠️  Missing $MISSING required tools!${NC}"
    echo ""
    echo "Please install:"
    command -v node &> /dev/null || echo "  - Node.js 18+ (https://nodejs.org)"
    command -v go &> /dev/null || echo "  - Go 1.21+ (https://go.dev)"
    command -v cargo &> /dev/null || echo "  - Rust (https://rustup.rs)"
    echo ""
    exit 1
fi

echo -e "${GREEN}✓ All prerequisites installed!${NC}"
echo ""

# Create workspace
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 2: Setting Up Workspace${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd "$WORKSPACE"

echo -e "${GREEN}✓${NC} Using workspace: $WORKSPACE"
echo ""

# Verify required folders exist
verify_folder() {
    local folder="$1"
    if [ -d "$folder" ]; then
        echo -e "${GREEN}✓${NC} $folder"
    else
        echo -e "${RED}✗${NC} $folder not found!"
        return 1
    fi
}

echo -e "${BLUE}Verifying source folders...${NC}"
MISSING_FOLDERS=0
verify_folder "dpb-mcp-typescript" || MISSING_FOLDERS=$((MISSING_FOLDERS + 1))
verify_folder "dpb-mcp-go" || MISSING_FOLDERS=$((MISSING_FOLDERS + 1))
verify_folder "dpb-mcp-rust" || MISSING_FOLDERS=$((MISSING_FOLDERS + 1))
verify_folder "dpb-benchmark" || MISSING_FOLDERS=$((MISSING_FOLDERS + 1))

if [ $MISSING_FOLDERS -gt 0 ]; then
    echo ""
    echo -e "${RED}⚠️  Missing $MISSING_FOLDERS required folders!${NC}"
    echo -e "Please ensure you have cloned the complete repository."
    exit 1
fi

echo ""

# Setup Test Repository
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 3: Setting Up Test Repository${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Default test repository
DEFAULT_REPO_URL="https://github.com/AzuraCast/AzuraCast.git"
DEFAULT_REPO_NAME="azuracast"

# Prompt for repository URL (if not set via environment variable)
if [ -z "$TEST_REPO_URL" ]; then
    echo -e "${CYAN}Test Repository Configuration${NC}"
    echo -e "${DIM}Default: $DEFAULT_REPO_URL (AzuraCast)${NC}"
    echo ""
    echo -e "${YELLOW}Enter repository URL (or press Enter to use default):${NC}"
    read -r USER_REPO_URL
    
    if [ -z "$USER_REPO_URL" ]; then
        # User pressed Enter, use default
        TEST_REPO_URL="$DEFAULT_REPO_URL"
        TEST_REPO_NAME="$DEFAULT_REPO_NAME"
        echo -e "${GREEN}✓${NC} Using default repository: ${BOLD}$DEFAULT_REPO_NAME${NC}"
    else
        # User provided custom URL
        TEST_REPO_URL="$USER_REPO_URL"
        # Extract repository name from URL
        if [[ "$USER_REPO_URL" =~ github.com[:/]([^/]+)/([^/]+) ]]; then
            TEST_REPO_NAME="${BASH_REMATCH[2]%.git}"
        else
            # Fallback: use last part of URL
            TEST_REPO_NAME=$(basename "$USER_REPO_URL" .git)
        fi
        echo -e "${GREEN}✓${NC} Using custom repository: ${BOLD}$TEST_REPO_NAME${NC}"
    fi
    echo ""
else
    # Environment variable already set
    TEST_REPO_NAME="${TEST_REPO_NAME:-$(basename "$TEST_REPO_URL" .git)}"
    echo -e "${GREEN}✓${NC} Using repository from environment: ${BOLD}$TEST_REPO_NAME${NC}"
    echo ""
fi

TEST_REPO_PATH="$WORKSPACE/test-repos/$TEST_REPO_NAME"

mkdir -p "$WORKSPACE/test-repos"
if [ ! -d "$TEST_REPO_PATH" ]; then
    echo -e "${BLUE}📥 Cloning test repository...${NC}"
    git clone --depth 1 "$TEST_REPO_URL" "$TEST_REPO_PATH"
    echo -e "${GREEN}✓${NC} Test repository cloned: $TEST_REPO_NAME"
else
    echo -e "${GREEN}✓${NC} Test repository already exists: $TEST_REPO_NAME"
fi

# Count source files
PHP_FILES=$(find "$TEST_REPO_PATH" -name "*.php" 2>/dev/null | wc -l)
JS_FILES=$(find "$TEST_REPO_PATH" -name "*.js" -o -name "*.ts" 2>/dev/null | wc -l)
echo -e "${BLUE}📊 Source files in test repo: PHP=$PHP_FILES, JS/TS=$JS_FILES${NC}"
echo ""

# Build TypeScript
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 4: Building TypeScript Implementation${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

if [ -d "dpb-mcp-typescript" ]; then
    cd dpb-mcp-typescript
    echo -e "${BLUE}🔧 Installing dependencies...${NC}"
    if npm install 2>&1; then
        echo -e "${BLUE}🏗️  Building TypeScript...${NC}"
        if npm run build 2>&1; then
            if [ -f "build/server.js" ]; then
                SIZE=$(du -sh build/server.js | cut -f1)
                echo -e "${GREEN}✓${NC} TypeScript build complete (server.js: $SIZE)"
            else
                echo -e "${RED}✗${NC} TypeScript build failed - server.js not created"
            fi
        else
            echo -e "${RED}✗${NC} TypeScript build failed"
        fi
    else
        echo -e "${RED}✗${NC} TypeScript npm install failed"
    fi
    cd ..
else
    echo -e "${RED}✗${NC} TypeScript source folder not found"
fi
echo ""

# Build Go
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 5: Building Go Implementation${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

if [ -d "dpb-mcp-go" ]; then
    cd dpb-mcp-go
    echo -e "${BLUE}🔧 Downloading Go dependencies...${NC}"
    go mod download 2>&1 || true
    go mod tidy 2>&1 || true
    echo -e "${BLUE}🏗️  Building Go binary...${NC}"
    if make build 2>&1; then
        if [ -f "build/dpb-mcp" ]; then
            SIZE=$(du -sh build/dpb-mcp | cut -f1)
            echo -e "${GREEN}✓${NC} Go build complete (binary: $SIZE)"
        else
            echo -e "${RED}✗${NC} Go build failed - binary not created"
        fi
    else
        echo -e "${RED}✗${NC} Go build failed"
    fi
    cd ..
else
    echo -e "${RED}✗${NC} Go source folder not found"
fi
echo ""

# Build Rust
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 6: Building Rust Implementation (this may take a while...)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

if [ -d "dpb-mcp-rust" ]; then
    cd dpb-mcp-rust
    echo -e "${BLUE}🏗️  Building Rust binary (optimized - this may take a while)...${NC}"
    if cargo build --release 2>&1; then
        if [ -f "target/release/dpb-mcp" ]; then
            SIZE=$(du -sh target/release/dpb-mcp | cut -f1)
            echo -e "${GREEN}✓${NC} Rust build complete (binary: $SIZE)"
        else
            echo -e "${RED}✗${NC} Rust build failed - binary not created"
        fi
    else
        echo -e "${RED}✗${NC} Rust build failed"
    fi
    cd ..
else
    echo -e "${RED}✗${NC} Rust source folder not found"
fi
echo ""

# Build report generator
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 7: Building Report Generator (Go)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

if [ -d "dpb-benchmark" ]; then
    cd dpb-benchmark/scripts
    echo -e "${BLUE}🏗️  Building Go report generator...${NC}"
    go build -o generate-report generate-report.go > /dev/null 2>&1
    
    if [ -f "generate-report" ]; then
        echo -e "${GREEN}✓${NC} Report generator built"
    else
        echo -e "${YELLOW}⚠️${NC}  Using Python fallback"
    fi
    cd ../..
fi
echo ""

# Build Dashboard Server
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 8: Building Dashboard Server${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

if [ -d "dpb-benchmark" ]; then
    cd dpb-benchmark
    echo -e "${BLUE}🏗️  Building dashboard server...${NC}"
    if make build > /dev/null 2>&1; then
        if [ -f "server/dashboard-server" ]; then
            SIZE=$(du -sh server/dashboard-server | cut -f1)
            echo -e "${GREEN}✓${NC} Dashboard server built (binary: $SIZE)"
        else
            echo -e "${YELLOW}⚠️${NC}  Dashboard server build completed but binary not found"
        fi
    else
        echo -e "${YELLOW}⚠️${NC}  Dashboard server build failed (optional)"
    fi
    cd ..
else
    echo -e "${YELLOW}⚠️${NC}  Benchmark directory not found, skipping dashboard server"
fi
echo ""

# Quick smoke tests
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 9: Running Quick Tests${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

TEST_JSON='{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}'

# Test TypeScript
if [ -f "dpb-mcp-typescript/build/server.js" ]; then
    echo -e "${BLUE}Testing TypeScript...${NC}"
    if echo "$TEST_JSON" | node dpb-mcp-typescript/build/server.js > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} TypeScript: Working"
    else
        echo -e "${RED}✗${NC} TypeScript: Failed"
    fi
fi

# Test Go
if [ -f "dpb-mcp-go/build/dpb-mcp" ]; then
    echo -e "${BLUE}Testing Go...${NC}"
    if echo "$TEST_JSON" | dpb-mcp-go/build/dpb-mcp > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Go: Working"
    else
        echo -e "${RED}✗${NC} Go: Failed"
    fi
fi

# Test Rust
if [ -f "dpb-mcp-rust/target/release/dpb-mcp" ]; then
    echo -e "${BLUE}Testing Rust...${NC}"
    if echo "$TEST_JSON" | dpb-mcp-rust/target/release/dpb-mcp > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Rust: Working"
    else
        echo -e "${RED}✗${NC} Rust: Failed"
    fi
fi

echo ""

# Summary
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  ✅ Setup Complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

echo -e "${BLUE}📊 What's Ready:${NC}"
echo ""

TS_STATUS="${RED}✗ Not built${NC}"
GO_STATUS="${RED}✗ Not built${NC}"
RUST_STATUS="${RED}✗ Not built${NC}"

[ -f "dpb-mcp-typescript/build/server.js" ] && TS_STATUS="${GREEN}✓ Ready${NC}"
[ -f "dpb-mcp-go/build/dpb-mcp" ] && GO_STATUS="${GREEN}✓ Ready${NC}"
[ -f "dpb-mcp-rust/target/release/dpb-mcp" ] && RUST_STATUS="${GREEN}✓ Ready${NC}"

echo -e "  TypeScript: $TS_STATUS"
echo -e "  Go:         $GO_STATUS"
echo -e "  Rust:       $RUST_STATUS"
echo ""

echo -e "${BLUE}📁 Locations:${NC}"
echo "  Workspace:  $WORKSPACE"
echo "  Test Repo:  $TEST_REPO_PATH"
echo "  Benchmark:  $WORKSPACE/dpb-benchmark"
echo ""

# Auto-run benchmark
echo -e "${YELLOW}🚀 Running Benchmark...${NC}"
echo ""

if [ -f "$WORKSPACE/dpb-benchmark/scripts/run-benchmark.sh" ]; then
    chmod +x "$WORKSPACE/dpb-benchmark/scripts/run-benchmark.sh"
    cd "$WORKSPACE/dpb-benchmark"
    ./scripts/run-benchmark.sh "$TEST_REPO_PATH"
    
    echo ""
    echo -e "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${PURPLE}  ✅ Build & Benchmark Complete!${NC}"
    echo -e "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    # Dashboard and Documentation Setup
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  Dashboard & Documentation Setup${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${YELLOW}The dashboard server allows you to:${NC}"
    echo "  • View interactive analysis dashboard"
    echo "  • Serve generated documentation"
    echo "  • Access at http://localhost:8080"
    echo ""
    echo -e "${CYAN}Would you like to set up dashboard and documentation visualization? [Y/n]${NC}"
    read -r -t 15 response
    response=${response:-Y}
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        # Check if dashboard server is built
        DASHBOARD_SERVER="$WORKSPACE/dpb-benchmark/server/dashboard-server"
        if [ ! -f "$DASHBOARD_SERVER" ]; then
            echo -e "${YELLOW}⚠️  Dashboard server not found. Building now...${NC}"
            cd "$WORKSPACE/dpb-benchmark"
            make build > /dev/null 2>&1
            cd "$WORKSPACE"
        fi
        
        # Generate documentation with all three implementations
        echo -e "${BLUE}📚 Generating documentation with all implementations (HTML format)...${NC}"
        echo -e "${DIM}  No Python required - using native HTML generation${NC}"
        echo ""
        
        DOCS_GENERATED=0
        
        # TypeScript implementation
        DOCS_TS_DIR="$WORKSPACE/docs-typescript"
        if [ -f "dpb-mcp-typescript/build/server.js" ]; then
            echo -e "${BLUE}  Generating with TypeScript...${NC}"
            DOCS_REQUEST='{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"'$TEST_REPO_PATH'","output_dir":"'$DOCS_TS_DIR'","include_changelog":true,"format":"html"}},"id":1}'
            ERROR_OUTPUT=$(echo "$DOCS_REQUEST" | node dpb-mcp-typescript/build/server.js 2>&1)
            EXIT_CODE=$?
            if [ $EXIT_CODE -eq 0 ] && [ -f "$DOCS_TS_DIR/index.html" ]; then
                echo -e "${GREEN}    ✓${NC} TypeScript docs: $DOCS_TS_DIR/index.html"
                DOCS_GENERATED=$((DOCS_GENERATED + 1))
            else
                echo -e "${RED}    ✗${NC} TypeScript failed"
                if [ -n "$ERROR_OUTPUT" ]; then
                    echo -e "${DIM}      Error: $(echo "$ERROR_OUTPUT" | head -3 | tr '\n' ' ')${NC}"
                fi
            fi
        else
            echo -e "${YELLOW}    ⚠${NC}  TypeScript server not found: dpb-mcp-typescript/build/server.js"
        fi
        
        # Go implementation
        DOCS_GO_DIR="$WORKSPACE/docs-go"
        if [ -f "dpb-mcp-go/build/dpb-mcp" ]; then
            echo -e "${BLUE}  Generating with Go...${NC}"
            DOCS_REQUEST='{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"'$TEST_REPO_PATH'","output_dir":"'$DOCS_GO_DIR'","include_changelog":true,"format":"html"}},"id":1}'
            ERROR_OUTPUT=$(echo "$DOCS_REQUEST" | dpb-mcp-go/build/dpb-mcp 2>&1)
            EXIT_CODE=$?
            if [ $EXIT_CODE -eq 0 ] && [ -f "$DOCS_GO_DIR/index.html" ]; then
                echo -e "${GREEN}    ✓${NC} Go docs: $DOCS_GO_DIR/index.html"
                DOCS_GENERATED=$((DOCS_GENERATED + 1))
            else
                echo -e "${RED}    ✗${NC} Go failed"
                if [ -n "$ERROR_OUTPUT" ]; then
                    echo -e "${DIM}      Error: $(echo "$ERROR_OUTPUT" | head -3 | tr '\n' ' ')${NC}"
                fi
            fi
        else
            echo -e "${YELLOW}    ⚠${NC}  Go binary not found: dpb-mcp-go/build/dpb-mcp"
        fi
        
        # Rust implementation
        DOCS_RUST_DIR="$WORKSPACE/docs-rust"
        if [ -f "dpb-mcp-rust/target/release/dpb-mcp" ]; then
            echo -e "${BLUE}  Generating with Rust...${NC}"
            DOCS_REQUEST='{"jsonrpc":"2.0","method":"tools/call","params":{"name":"generate_mkdocs_docs","arguments":{"repo_path":"'$TEST_REPO_PATH'","output_dir":"'$DOCS_RUST_DIR'","include_changelog":true,"format":"html"}},"id":1}'
            ERROR_OUTPUT=$(echo "$DOCS_REQUEST" | dpb-mcp-rust/target/release/dpb-mcp 2>&1)
            EXIT_CODE=$?
            if [ $EXIT_CODE -eq 0 ] && [ -f "$DOCS_RUST_DIR/index.html" ]; then
                echo -e "${GREEN}    ✓${NC} Rust docs: $DOCS_RUST_DIR/index.html"
                DOCS_GENERATED=$((DOCS_GENERATED + 1))
            else
                echo -e "${RED}    ✗${NC} Rust failed"
                if [ -n "$ERROR_OUTPUT" ]; then
                    echo -e "${DIM}      Error: $(echo "$ERROR_OUTPUT" | head -3 | tr '\n' ' ')${NC}"
                fi
            fi
        else
            echo -e "${YELLOW}    ⚠${NC}  Rust binary not found: dpb-mcp-rust/target/release/dpb-mcp"
        fi
        
        echo ""
        if [ $DOCS_GENERATED -gt 0 ]; then
            echo -e "${GREEN}✓${NC} Documentation generated with $DOCS_GENERATED implementation(s)"
            echo -e "${DIM}  • TypeScript: docs-typescript/index.html${NC}"
            echo -e "${DIM}  • Go:         docs-go/index.html${NC}"
            echo -e "${DIM}  • Rust:       docs-rust/index.html${NC}"
            echo ""
            echo -e "${CYAN}💡 To view docs:${NC}"
            echo -e "${DIM}  • Open HTML files directly in browser${NC}"
            echo -e "${DIM}  • Or use dashboard server to serve them${NC}"
        else
            echo -e "${RED}✗${NC}  No documentation generated"
            echo -e "${DIM}  Check errors above or ensure implementations are built${NC}"
        fi
        echo ""
        
        # Ask to start dashboard server or open docs
        echo ""
        echo -e "${CYAN}Would you like to start the dashboard server now? [Y/n]${NC}"
        read -r -t 10 server_response
        server_response=${server_response:-Y}
        
        if [[ "$server_response" =~ ^[Yy]$ ]]; then
            # Check if port 8080 is already in use
            if lsof -ti:8080 > /dev/null 2>&1; then
                EXISTING_PID=$(lsof -ti:8080 | head -1)
                echo -e "${YELLOW}⚠️  Port 8080 is already in use (PID: $EXISTING_PID)${NC}"
                echo -e "${CYAN}Would you like to stop the existing server and start a new one? [Y/n]${NC}"
                read -r -t 10 kill_response
                kill_response=${kill_response:-Y}
                if [[ "$kill_response" =~ ^[Yy]$ ]]; then
                    echo -e "${BLUE}Stopping existing server...${NC}"
                    kill $EXISTING_PID 2>/dev/null || lsof -ti:8080 | xargs kill 2>/dev/null
                    sleep 1
                    echo -e "${GREEN}✓${NC} Existing server stopped"
                else
                    echo -e "${CYAN}Would you like to open the dashboard in your browser anyway? [Y/n]${NC}"
                    read -r -t 10 open_response
                    open_response=${open_response:-Y}
                    if [[ "$open_response" =~ ^[Yy]$ ]]; then
                        if command -v open &> /dev/null; then
                            open "http://localhost:8080"
                        elif command -v xdg-open &> /dev/null; then
                            xdg-open "http://localhost:8080"
                        fi
                    fi
                    cd "$WORKSPACE"
                    exit 0
                fi
            fi
            
            # Try to find an available port if 8080 is still in use
            PORT=8080
            if lsof -ti:$PORT > /dev/null 2>&1; then
                echo -e "${YELLOW}⚠️  Port $PORT still in use, trying alternative port...${NC}"
                for alt_port in 8081 8082 8083 8084 8085; do
                    if ! lsof -ti:$alt_port > /dev/null 2>&1; then
                        PORT=$alt_port
                        echo -e "${GREEN}✓${NC} Using port $PORT instead"
                        break
                    fi
                done
            fi
            
            echo -e "${GREEN}🚀 Starting dashboard server on port $PORT...${NC}"
            echo -e "${DIM}Server will run in the background. Press Ctrl+C to stop.${NC}"
            echo ""
            cd "$WORKSPACE/dpb-benchmark"
            ./server/dashboard-server -port=$PORT -open=true &
            SERVER_PID=$!
            sleep 2
            echo -e "${GREEN}✓${NC} Dashboard server started (PID: $SERVER_PID)"
            echo -e "${GREEN}✓${NC} Dashboard available at: ${BOLD}http://localhost:$PORT${NC}"
            echo ""
            echo -e "${DIM}To stop the server later: kill $SERVER_PID${NC}"
            echo -e "${DIM}Or use: lsof -ti:$PORT | xargs kill${NC}"
            cd "$WORKSPACE"
        else
            # Offer to open docs HTML files directly
            echo ""
            if [ $DOCS_GENERATED -gt 0 ]; then
                echo -e "${CYAN}Would you like to open generated documentation in your browser? [Y/n]${NC}"
                read -r -t 10 docs_response
                docs_response=${docs_response:-Y}
                if [[ "$docs_response" =~ ^[Yy]$ ]]; then
                    # Open first available docs
                    if [ -f "$DOCS_TS_DIR/index.html" ]; then
                        if command -v open &> /dev/null; then
                            open "$DOCS_TS_DIR/index.html"
                        elif command -v xdg-open &> /dev/null; then
                            xdg-open "$DOCS_TS_DIR/index.html"
                        elif command -v start &> /dev/null; then
                            start "$DOCS_TS_DIR/index.html"
                        fi
                        echo -e "${GREEN}✓${NC} Opened TypeScript docs"
                    elif [ -f "$DOCS_GO_DIR/index.html" ]; then
                        if command -v open &> /dev/null; then
                            open "$DOCS_GO_DIR/index.html"
                        elif command -v xdg-open &> /dev/null; then
                            xdg-open "$DOCS_GO_DIR/index.html"
                        elif command -v start &> /dev/null; then
                            start "$DOCS_GO_DIR/index.html"
                        fi
                        echo -e "${GREEN}✓${NC} Opened Go docs"
                    elif [ -f "$DOCS_RUST_DIR/index.html" ]; then
                        if command -v open &> /dev/null; then
                            open "$DOCS_RUST_DIR/index.html"
                        elif command -v xdg-open &> /dev/null; then
                            xdg-open "$DOCS_RUST_DIR/index.html"
                        elif command -v start &> /dev/null; then
                            start "$DOCS_RUST_DIR/index.html"
                        fi
                        echo -e "${GREEN}✓${NC} Opened Rust docs"
                    fi
                fi
            fi
            
            echo ""
            echo -e "${BLUE}To start the dashboard server later:${NC}"
            echo "  cd dpb-benchmark"
            echo "  make serve"
            echo ""
            echo -e "${BLUE}Or open dashboard/docs directly:${NC}"
            DASHBOARD_PATH="$WORKSPACE/dpb-benchmark/dashboard/index.html"
            if [ -f "$DASHBOARD_PATH" ]; then
                echo "  Dashboard: $DASHBOARD_PATH"
            fi
            if [ $DOCS_GENERATED -gt 0 ]; then
                echo "  Docs: docs-typescript/index.html, docs-go/index.html, docs-rust/index.html"
            fi
        fi
    else
        echo ""
        echo -e "${DIM}Dashboard setup skipped.${NC}"
        echo -e "${BLUE}To set up later, see:${NC}"
        echo "  • README.md - Dashboard section"
        echo "  • dpb-benchmark/README.md"
        echo ""
        echo -e "${BLUE}Quick start:${NC}"
        echo "  cd dpb-benchmark && make serve"
    fi
else
    echo -e "${RED}✗ Benchmark script not found${NC}"
fi

echo ""
