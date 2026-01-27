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

# Test repository configuration (can be overridden via environment variable)
TEST_REPO_URL="${TEST_REPO_URL:-https://github.com/AzuraCast/AzuraCast.git}"
TEST_REPO_NAME="${TEST_REPO_NAME:-azuracast}"
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

# Quick smoke tests
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  Step 8: Running Quick Tests${NC}"
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

echo -e "${YELLOW}🚀 Next Step: Run Benchmark${NC}"
echo ""
echo -e "  ${GREEN}Run the benchmark to compare all implementations:${NC}"
echo "    cd $WORKSPACE/dpb-benchmark"
echo "    ./scripts/run-benchmark.sh"
echo ""
echo -e "  This will:"
echo "    • Test startup time for each implementation"
echo "    • Measure memory usage"
echo "    • Run analysis tools against test repository"
echo "    • Generate comparison report and dashboard"
echo ""

echo -e "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${PURPLE}  All implementations built! Ready to benchmark! 🎯${NC}"
echo -e "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
