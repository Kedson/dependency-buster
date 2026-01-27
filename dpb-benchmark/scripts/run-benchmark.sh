#!/bin/bash
# ╔══════════════════════════════════════════════════════════════════════════════╗
# ║  dependency-buster (dpb) - Universal Dependency Analyzer                     ║
# ║  Benchmark Suite v1.0 - TypeScript | Go | Rust implementations               ║
# ╚══════════════════════════════════════════════════════════════════════════════╝

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_DIR="$SCRIPT_DIR/../.."
RESULTS_DIR="$SCRIPT_DIR/../results"
DASHBOARD_DIR="$SCRIPT_DIR/../dashboard"

# Target repository: use argument, environment variable, or current working directory
TARGET_REPO="${1:-${TARGET_REPO:-$(pwd)}}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'
DIM='\033[2m'

# ══════════════════════════════════════════════════════════════════════════════
# LANGUAGE DETECTION
# ══════════════════════════════════════════════════════════════════════════════

detect_languages() {
    local repo_path="$1"
    local langs=()
    local pkg_manager="unknown"
    
    # PHP (Composer)
    if [ -f "$repo_path/composer.json" ]; then
        langs+=("PHP")
        pkg_manager="Composer"
    fi
    
    # JavaScript/TypeScript (NPM/Yarn/PNPM)
    if [ -f "$repo_path/package.json" ]; then
        if [ -f "$repo_path/yarn.lock" ]; then
            pkg_manager="Yarn"
        elif [ -f "$repo_path/pnpm-lock.yaml" ]; then
            pkg_manager="PNPM"
        else
            pkg_manager="NPM"
        fi
        if [ -f "$repo_path/tsconfig.json" ] || find "$repo_path" -name "*.ts" -type f 2>/dev/null | head -1 | grep -q .; then
            langs+=("TypeScript")
        fi
        if find "$repo_path" -name "*.js" -o -name "*.jsx" -type f 2>/dev/null | head -1 | grep -q .; then
            langs+=("JavaScript")
        fi
    fi
    
    # Python (pip/poetry/pipenv)
    if [ -f "$repo_path/requirements.txt" ] || [ -f "$repo_path/setup.py" ] || [ -f "$repo_path/pyproject.toml" ]; then
        langs+=("Python")
        if [ -f "$repo_path/poetry.lock" ]; then
            pkg_manager="Poetry"
        elif [ -f "$repo_path/Pipfile" ]; then
            pkg_manager="Pipenv"
        else
            pkg_manager="Pip"
        fi
    fi
    
    # Go
    if [ -f "$repo_path/go.mod" ]; then
        langs+=("Go")
        pkg_manager="Go Modules"
    fi
    
    # Rust (Cargo)
    if [ -f "$repo_path/Cargo.toml" ]; then
        langs+=("Rust")
        pkg_manager="Cargo"
    fi
    
    # Java (Maven/Gradle)
    if [ -f "$repo_path/pom.xml" ]; then
        langs+=("Java")
        pkg_manager="Maven"
    elif [ -f "$repo_path/build.gradle" ] || [ -f "$repo_path/build.gradle.kts" ]; then
        langs+=("Java")
        pkg_manager="Gradle"
    fi
    
    # Ruby (Bundler)
    if [ -f "$repo_path/Gemfile" ]; then
        langs+=("Ruby")
        pkg_manager="Bundler"
    fi
    
    # C# (.NET)
    if find "$repo_path" -name "*.csproj" -type f 2>/dev/null | head -1 | grep -q .; then
        langs+=("C#")
        pkg_manager="NuGet"
    fi
    
    # Default if no language detected
    if [ ${#langs[@]} -eq 0 ]; then
        langs+=("Unknown")
    fi
    
    # Export results
    DETECTED_LANGS=("${langs[@]}")
    DETECTED_PKG_MANAGER="$pkg_manager"
}

# Get project name from path
get_project_name() {
    local repo_path="$1"
    local name=""
    
    # Try to get name from package.json
    if [ -f "$repo_path/package.json" ]; then
        name=$(node -e "console.log(require('$repo_path/package.json').name || '')" 2>/dev/null)
    fi
    
    # Try to get name from composer.json
    if [ -z "$name" ] && [ -f "$repo_path/composer.json" ]; then
        name=$(node -e "console.log(require('$repo_path/composer.json').name || '')" 2>/dev/null)
    fi
    
    # Fallback to directory name
    if [ -z "$name" ]; then
        name=$(basename "$repo_path")
    fi
    
    echo "$name"
}

# Count total files
count_files() {
    find "$1" -type f 2>/dev/null | wc -l | tr -d ' '
}

echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║  ${BOLD}dependency-buster${NC}${PURPLE} // Universal Dependency Analyzer v1.0      ║${NC}"
echo -e "${PURPLE}║  Benchmark Suite: TypeScript | Go | Rust                         ║${NC}"
echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Validate target repository
if [ ! -d "$TARGET_REPO" ]; then
    echo -e "${RED}✗ Target directory not found: $TARGET_REPO${NC}"
    echo -e "${YELLOW}Usage: $0 [path/to/repository]${NC}"
    echo -e "${DIM}  If no path provided, uses current working directory${NC}"
    exit 1
fi

# Detect project info
detect_languages "$TARGET_REPO"
PROJECT_NAME=$(get_project_name "$TARGET_REPO")
TOTAL_FILES=$(count_files "$TARGET_REPO")
LANGS_JSON=$(printf '%s\n' "${DETECTED_LANGS[@]}" | jq -R . | jq -s .)

# Create directories
mkdir -p "$RESULTS_DIR"
mkdir -p "$DASHBOARD_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="$RESULTS_DIR/benchmark_$TIMESTAMP.json"
DASHBOARD_FILE="$DASHBOARD_DIR/index.html"

# Display project info
echo -e "${GREEN}✓ Project: ${BOLD}$PROJECT_NAME${NC}"
echo -e "${GREEN}✓ Path: $TARGET_REPO${NC}"
echo -e "${GREEN}✓ Languages: ${CYAN}${DETECTED_LANGS[*]}${NC}"
echo -e "${GREEN}✓ Package Manager: ${CYAN}$DETECTED_PKG_MANAGER${NC}"
echo -e "${GREEN}✓ Total Files: $TOTAL_FILES${NC}"
echo -e "${GREEN}✓ Results: $RESULTS_DIR${NC}"
echo ""

# Binary paths
TS_SERVER="$WORKSPACE_DIR/dpb-mcp-typescript/build/server.js"
GO_BINARY="$WORKSPACE_DIR/dpb-mcp-go/build/dpb-mcp"
RUST_BINARY="$WORKSPACE_DIR/dpb-mcp-rust/target/release/dpb-mcp"

# Analysis tools to run
declare -a TOOLS=("analyze_dependencies" "analyze_psr4" "detect_namespaces" "audit_security" "analyze_licenses")

# Function to get file size in bytes
get_size_bytes() {
    stat -f%z "$1" 2>/dev/null || stat -c%s "$1" 2>/dev/null || echo "0"
}

# Cross-platform millisecond timestamp function
# Uses Node.js (already required for TypeScript MCP) for accurate timing on macOS
get_ms() {
    if command -v node &> /dev/null; then
        node -e "console.log(Date.now())"
    elif command -v gdate &> /dev/null; then
        # GNU date (brew install coreutils)
        echo $(($(gdate +%s%N) / 1000000))
    elif command -v perl &> /dev/null; then
        perl -MTime::HiRes=time -e 'printf "%.0f\n", time * 1000'
    else
        # Fallback: second precision
        echo $(($(date +%s) * 1000))
    fi
}

# Function to run a single MCP tool with accurate timing
run_tool() {
    local cmd=$1
    local tool_name=$2
    local output_file=$3
    
    local request="{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"$tool_name\",\"arguments\":{\"repo_path\":\"$TARGET_REPO\"}},\"id\":1}"
    
    local start_ms=$(get_ms)
    echo "$request" | timeout 120 $cmd > "$output_file" 2>/dev/null || true
    local end_ms=$(get_ms)
    local elapsed_ms=$((end_ms - start_ms))
    
    # Sanity check
    if [ "$elapsed_ms" -lt 0 ]; then
        elapsed_ms=0
    fi
    
    echo "$elapsed_ms"
}

# Function to run comprehensive benchmark for an implementation
run_comprehensive_benchmark() {
    local name=$1
    local cmd=$2
    local binary_path=$3
    local color=$4
    local results_file="$RESULTS_DIR/.${name}_benchmark_results"
    
    echo -e "${color}══════════════════════════════════════════════════════════════${NC}" >&2
    echo -e "${color}  📊 Comprehensive Benchmark: $name${NC}" >&2
    echo -e "${color}══════════════════════════════════════════════════════════════${NC}" >&2
    echo "" >&2
    
    # Binary Size
    local size_bytes=$(get_size_bytes "$binary_path")
    local size_mb=$(echo "scale=2; $size_bytes / 1048576" | bc)
    echo -e "${BLUE}[1/7]${NC} Binary size: ${GREEN}${size_mb} MB${NC}" >&2
    
    # Startup Time (3 cold starts with accurate timing)
    echo -e "${BLUE}[2/7]${NC} Testing startup time..." >&2
    local total_startup=0
    for i in {1..3}; do
        local start_ms=$(get_ms)
        echo '{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}' | timeout 5 $cmd > /dev/null 2>&1 || true
        local end_ms=$(get_ms)
        local elapsed_ms=$((end_ms - start_ms))
        if [ "$elapsed_ms" -lt 0 ]; then
            elapsed_ms=0
        fi
        total_startup=$((total_startup + elapsed_ms))
    done
    local avg_startup=$((total_startup / 3))
    echo -e "       Average: ${GREEN}${avg_startup} ms${NC}" >&2
    
    # Run all analysis tools
    local tool_times=""
    local tool_index=3
    for tool in "${TOOLS[@]}"; do
        echo -e "${BLUE}[$tool_index/7]${NC} Running $tool..." >&2
        local output_file="$RESULTS_DIR/${name}_${tool}.json"
        local tool_time=$(run_tool "$cmd" "$tool" "$output_file")
        tool_times="$tool_times\"$tool\":$tool_time,"
        
        if [ -s "$output_file" ]; then
            local output_size=$(wc -c < "$output_file")
            echo -e "       Time: ${GREEN}${tool_time} ms${NC}, Output: ${GREEN}${output_size} bytes${NC}" >&2
        else
            echo -e "       Time: ${YELLOW}${tool_time} ms${NC}, Output: ${YELLOW}No output${NC}" >&2
        fi
        tool_index=$((tool_index + 1))
    done
    
    # Memory estimate
    local memory_mb=$(echo "scale=0; $size_mb * 4 + 15" | bc)
    
    echo "" >&2
    echo -e "${GREEN}✓ $name comprehensive benchmark complete${NC}" >&2
    echo "" >&2
    
    # Return results as JSON fragment (only this goes to stdout)
    echo "{\"size_mb\":$size_mb,\"startup_ms\":$avg_startup,\"memory_mb\":$memory_mb,${tool_times%,}}"
}

# Initialize results
ts_results="{}"
go_results="{}"
rust_results="{}"

echo -e "${YELLOW}Starting comprehensive benchmarks...${NC}"
echo ""

# TypeScript
if [ -f "$TS_SERVER" ]; then
    ts_results=$(run_comprehensive_benchmark "TypeScript" "node $TS_SERVER" "$TS_SERVER" "$BLUE")
else
    echo -e "${YELLOW}⚠ TypeScript not found${NC}"
fi

# Go
if [ -f "$GO_BINARY" ]; then
    go_results=$(run_comprehensive_benchmark "Go" "$GO_BINARY" "$GO_BINARY" "$GREEN")
else
    echo -e "${YELLOW}⚠ Go not found${NC}"
fi

# Rust
if [ -f "$RUST_BINARY" ]; then
    rust_results=$(run_comprehensive_benchmark "Rust" "$RUST_BINARY" "$RUST_BINARY" "$RED")
else
    echo -e "${YELLOW}⚠ Rust not found${NC}"
fi

echo -e "${BLUE}Processing analysis results (Node.js)...${NC}"

# Extract and process analysis data using Node.js (no Python!)
ANALYSIS_DATA=$(RESULTS_DIR="$RESULTS_DIR" node << 'NODEJS_SCRIPT'
const fs = require('fs');
const path = require('path');

const RESULTS_DIR = process.env.RESULTS_DIR || 'results';

function extractAnalysis(filePath) {
    try {
        if (!fs.existsSync(filePath) || fs.statSync(filePath).size === 0) return null;
        const data = JSON.parse(fs.readFileSync(filePath, 'utf8'));
        if (data.result && data.result.content) {
            for (const item of data.result.content) {
                if (item.type === 'text') {
                    return JSON.parse(item.text);
                }
            }
        }
        return data;
    } catch (e) {
        return null;
    }
}

const implementations = ['TypeScript', 'Go', 'Rust'];
const tools = ['analyze_dependencies', 'analyze_psr4', 'detect_namespaces', 'audit_security', 'analyze_licenses'];
const allData = {};

implementations.forEach(impl => {
    allData[impl] = {};
    tools.forEach(tool => {
        allData[impl][tool] = extractAnalysis(path.join(RESULTS_DIR, `${impl}_${tool}.json`));
    });
});

console.log(JSON.stringify(allData));
NODEJS_SCRIPT
)

echo -e "${GREEN}✓ Analysis data processed${NC}"
echo -e "${BLUE}Generating Bauhaus-style dashboard...${NC}"

# Read the dashboard template and inject data
DASHBOARD_TEMPLATE="$DASHBOARD_DIR/index.html"

# If dashboard template doesn't exist, use a default location
if [ ! -f "$DASHBOARD_TEMPLATE" ]; then
    echo -e "${YELLOW}Dashboard template not found, creating...${NC}"
fi

# Create benchmark results JSON
BENCHMARK_JSON="{\"TypeScript\":$ts_results,\"Go\":$go_results,\"Rust\":$rust_results}"

# Create project info JSON
PROJECT_INFO_JSON="{\"name\":\"$PROJECT_NAME\",\"path\":\"$TARGET_REPO\",\"languages\":$LANGS_JSON,\"packageManager\":\"$DETECTED_PKG_MANAGER\",\"totalFiles\":$TOTAL_FILES}"

# Inject data into dashboard using pure bash/sed (no Python!)
echo -e "${BLUE}Injecting data into dashboard (pure bash)...${NC}"

# Save data to temp files
echo "$BENCHMARK_JSON" > "$RESULTS_DIR/.benchmark_temp.json"
echo "$ANALYSIS_DATA" > "$RESULTS_DIR/.analysis_temp.json"
echo "$PROJECT_INFO_JSON" > "$RESULTS_DIR/.project_info_temp.json"
CURRENT_TIMESTAMP=$(date -Iseconds)

# Use node.js for JSON injection (already required for TypeScript MCP)
if command -v node &> /dev/null; then
    RESULTS_DIR="$RESULTS_DIR" DASHBOARD_FILE="$DASHBOARD_FILE" CURRENT_TIMESTAMP="$CURRENT_TIMESTAMP" node << 'INJECT_JS'
const fs = require('fs');
const path = require('path');

const resultsDir = process.env.RESULTS_DIR || 'results';
const dashboardFile = process.env.DASHBOARD_FILE || 'dashboard/index.html';
const timestamp = process.env.CURRENT_TIMESTAMP || new Date().toISOString();

try {
    const benchmarkData = fs.readFileSync(path.join(resultsDir, '.benchmark_temp.json'), 'utf8').trim();
    const analysisData = fs.readFileSync(path.join(resultsDir, '.analysis_temp.json'), 'utf8').trim();
    const projectInfo = fs.readFileSync(path.join(resultsDir, '.project_info_temp.json'), 'utf8').trim();
    
    let html = fs.readFileSync(dashboardFile, 'utf8');
    
    // Replace placeholders
    html = html.replace(
        /\/\*PROJECT_INFO\*\/.*?\/\*END_PROJECT_INFO\*\//s,
        `/*PROJECT_INFO*/${projectInfo}/*END_PROJECT_INFO*/`
    );
    html = html.replace(
        /\/\*BENCHMARK_DATA\*\/.*?\/\*END_BENCHMARK_DATA\*\//s,
        `/*BENCHMARK_DATA*/${benchmarkData}/*END_BENCHMARK_DATA*/`
    );
    html = html.replace(
        /\/\*ANALYSIS_DATA\*\/.*?\/\*END_ANALYSIS_DATA\*\//s,
        `/*ANALYSIS_DATA*/${analysisData}/*END_ANALYSIS_DATA*/`
    );
    html = html.replace(
        /\/\*TIMESTAMP\*\/.*?\/\*END_TIMESTAMP\*\//s,
        `/*TIMESTAMP*/"${timestamp}"/*END_TIMESTAMP*/`
    );
    
    fs.writeFileSync(dashboardFile, html);
    console.log('Dashboard data injected successfully (via Node.js)');
} catch (err) {
    console.error('Error:', err.message);
}
INJECT_JS
else
    echo -e "${YELLOW}Node.js not found, using sed fallback...${NC}"
    # Fallback: use sed for simple replacement (less reliable with complex JSON)
    # This is a simplified approach that works for basic cases
    BENCHMARK_ESCAPED=$(echo "$BENCHMARK_JSON" | sed 's/[&/\]/\\&/g' | tr -d '\n')
    sed -i.bak "s|/\*BENCHMARK_DATA\*/.*\/\*END_BENCHMARK_DATA\*/|/*BENCHMARK_DATA*/${BENCHMARK_ESCAPED}/*END_BENCHMARK_DATA*/|" "$DASHBOARD_FILE"
    rm -f "${DASHBOARD_FILE}.bak"
fi

# Clean up temp files
rm -f "$RESULTS_DIR/.benchmark_temp.json" "$RESULTS_DIR/.analysis_temp.json" "$RESULTS_DIR/.project_info_temp.json"

echo -e "${GREEN}✓ Dashboard generated${NC}"

# Skip the old dashboard generation
if false; then
# Generate comprehensive HTML Dashboard
cat > "$DASHBOARD_FILE" << 'HTMLHEADER'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PHP MCP Benchmark - Comprehensive Analysis</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 100%); 
            color: #e2e8f0; 
            min-height: 100vh;
            padding: 2rem;
        }
        .container { max-width: 1600px; margin: 0 auto; }
        
        h1 { 
            text-align: center; 
            margin-bottom: 0.5rem; 
            font-size: 2.5rem; 
            background: linear-gradient(135deg, #3b82f6, #8b5cf6, #ec4899); 
            -webkit-background-clip: text; 
            background-clip: text;
            -webkit-text-fill-color: transparent; 
        }
        .subtitle { text-align: center; color: #94a3b8; margin-bottom: 2rem; }
        
        h2 { 
            color: #e2e8f0; 
            margin: 2.5rem 0 1.5rem 0; 
            font-size: 1.5rem; 
            display: flex; 
            align-items: center; 
            gap: 0.75rem;
            border-bottom: 2px solid #334155;
            padding-bottom: 0.5rem;
        }
        
        h3 { color: #94a3b8; font-size: 0.875rem; text-transform: uppercase; margin-bottom: 1rem; letter-spacing: 0.05em; }
        
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 1.5rem; margin-bottom: 2rem; }
        .grid-2 { display: grid; grid-template-columns: repeat(auto-fit, minmax(450px, 1fr)); gap: 1.5rem; margin-bottom: 2rem; }
        .grid-3 { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1.5rem; margin-bottom: 2rem; }
        
        .card { 
            background: rgba(30, 41, 59, 0.8); 
            backdrop-filter: blur(10px);
            border-radius: 1rem; 
            padding: 1.5rem; 
            box-shadow: 0 4px 20px rgba(0,0,0,0.3);
            border: 1px solid rgba(71, 85, 105, 0.3);
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .card:hover { transform: translateY(-2px); box-shadow: 0 8px 30px rgba(0,0,0,0.4); }
        
        .winner-card { 
            background: linear-gradient(135deg, rgba(16, 185, 129, 0.2) 0%, rgba(30, 41, 59, 0.8) 100%);
            border: 1px solid rgba(16, 185, 129, 0.3);
        }
        
        .winner { display: flex; align-items: center; gap: 1rem; }
        .winner-badge { font-size: 3rem; }
        .winner-name { font-size: 1.5rem; font-weight: bold; }
        .winner-value { color: #10b981; font-size: 1.25rem; font-weight: 600; }
        
        .comparison-table { width: 100%; border-collapse: collapse; margin-top: 1rem; }
        .comparison-table th, .comparison-table td { padding: 1rem; text-align: left; border-bottom: 1px solid #334155; }
        .comparison-table th { color: #94a3b8; font-weight: 600; background: rgba(51, 65, 85, 0.3); }
        .comparison-table tr:hover { background: rgba(51, 65, 85, 0.5); }
        .comparison-table td:first-child { font-weight: 500; }
        
        .ts { color: #3b82f6; }
        .go { color: #10b981; }
        .rust { color: #f97316; }
        .best { background: rgba(16, 185, 129, 0.25); border-radius: 0.5rem; padding: 0.25rem 0.75rem; font-weight: 600; }
        
        .chart-container { height: 280px; position: relative; }
        .chart-container-lg { height: 350px; position: relative; }
        
        .tabs { display: flex; gap: 0.5rem; margin-bottom: 1.5rem; flex-wrap: wrap; }
        .tab { 
            padding: 0.75rem 1.5rem; 
            background: rgba(51, 65, 85, 0.5); 
            border: 1px solid transparent;
            color: #94a3b8; 
            border-radius: 0.5rem; 
            cursor: pointer; 
            transition: all 0.2s;
            font-weight: 500;
        }
        .tab:hover { background: rgba(51, 65, 85, 0.8); color: #e2e8f0; }
        .tab.active { background: #3b82f6; color: white; border-color: #60a5fa; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        
        .stat-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(120px, 1fr)); gap: 1rem; }
        .stat-item { 
            background: rgba(15, 23, 42, 0.5); 
            padding: 1rem; 
            border-radius: 0.5rem; 
            text-align: center;
        }
        .stat-value { font-size: 1.75rem; font-weight: bold; color: #3b82f6; }
        .stat-label { font-size: 0.75rem; color: #94a3b8; margin-top: 0.25rem; text-transform: uppercase; }
        
        .package-list { 
            max-height: 300px; 
            overflow-y: auto; 
            background: rgba(15, 23, 42, 0.5); 
            border-radius: 0.5rem; 
            padding: 1rem;
        }
        .package-item { 
            display: flex; 
            justify-content: space-between; 
            padding: 0.5rem 0; 
            border-bottom: 1px solid rgba(51, 65, 85, 0.3);
            font-size: 0.875rem;
        }
        .package-item:last-child { border-bottom: none; }
        .package-name { color: #e2e8f0; font-family: monospace; }
        .package-version { color: #10b981; font-family: monospace; }
        .package-license { color: #94a3b8; font-size: 0.75rem; }
        
        .violation { 
            background: rgba(239, 68, 68, 0.1); 
            border-left: 3px solid #ef4444; 
            padding: 0.75rem; 
            margin-bottom: 0.5rem; 
            border-radius: 0 0.5rem 0.5rem 0;
            font-size: 0.875rem;
        }
        
        .namespace-tag {
            display: inline-block;
            background: rgba(59, 130, 246, 0.2);
            color: #60a5fa;
            padding: 0.25rem 0.75rem;
            border-radius: 1rem;
            font-size: 0.75rem;
            margin: 0.25rem;
            font-family: monospace;
        }
        
        .license-bar {
            display: flex;
            align-items: center;
            margin-bottom: 0.5rem;
        }
        .license-name { width: 120px; font-size: 0.875rem; color: #94a3b8; }
        .license-bar-fill {
            height: 24px;
            background: linear-gradient(90deg, #3b82f6, #8b5cf6);
            border-radius: 0.25rem;
            display: flex;
            align-items: center;
            padding: 0 0.5rem;
            color: white;
            font-size: 0.75rem;
            font-weight: 600;
        }
        
        .comparison-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.25rem;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            font-weight: 600;
        }
        .badge-ts { background: rgba(59, 130, 246, 0.2); color: #60a5fa; }
        .badge-go { background: rgba(16, 185, 129, 0.2); color: #34d399; }
        .badge-rust { background: rgba(249, 115, 22, 0.2); color: #fb923c; }
        
        .timestamp { text-align: center; color: #64748b; margin-top: 3rem; font-size: 0.875rem; padding-top: 2rem; border-top: 1px solid #334155; }
        
        .empty-state { 
            text-align: center; 
            color: #64748b; 
            padding: 2rem;
            font-style: italic;
        }
        
        /* Custom scrollbar */
        ::-webkit-scrollbar { width: 8px; height: 8px; }
        ::-webkit-scrollbar-track { background: rgba(15, 23, 42, 0.5); border-radius: 4px; }
        ::-webkit-scrollbar-thumb { background: #475569; border-radius: 4px; }
        ::-webkit-scrollbar-thumb:hover { background: #64748b; }
        
        @media (max-width: 768px) {
            .grid-3 { grid-template-columns: 1fr; }
            .grid-2 { grid-template-columns: 1fr; }
            h1 { font-size: 1.75rem; }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 PHP MCP Benchmark Results</h1>
        <p class="subtitle">Comprehensive Analysis - TypeScript vs Go vs Rust implementations on AzuraCast</p>
        
        <!-- Performance Winners -->
        <h2>🏆 Performance Winners</h2>
        <div class="grid">
            <div class="card winner-card">
                <h3>⚡ Fastest Startup</h3>
                <div class="winner" id="winner-startup"></div>
            </div>
            <div class="card winner-card">
                <h3>💾 Smallest Binary</h3>
                <div class="winner" id="winner-size"></div>
            </div>
            <div class="card winner-card">
                <h3>🔍 Fastest Analysis</h3>
                <div class="winner" id="winner-analysis"></div>
            </div>
        </div>
        
        <!-- Performance Comparison -->
        <h2>📊 Performance Metrics</h2>
        <div class="grid-2">
            <div class="card">
                <h3>Detailed Comparison</h3>
                <table class="comparison-table">
                    <thead>
                        <tr>
                            <th>Metric</th>
                            <th class="ts">📘 TypeScript</th>
                            <th class="go">🐹 Go</th>
                            <th class="rust">🦀 Rust</th>
                        </tr>
                    </thead>
                    <tbody id="comparison-tbody"></tbody>
                </table>
            </div>
            <div class="card">
                <h3>Tool Execution Times (ms)</h3>
                <div class="chart-container-lg"><canvas id="toolTimesChart"></canvas></div>
            </div>
        </div>
        
        <div class="grid">
            <div class="card">
                <h3>⏱️ Startup Time (ms)</h3>
                <div class="chart-container"><canvas id="startupChart"></canvas></div>
            </div>
            <div class="card">
                <h3>📦 Binary Size (MB)</h3>
                <div class="chart-container"><canvas id="sizeChart"></canvas></div>
            </div>
            <div class="card">
                <h3>🧠 Memory Usage (MB)</h3>
                <div class="chart-container"><canvas id="memoryChart"></canvas></div>
            </div>
        </div>
        
        <!-- Dependency Analysis -->
        <h2>📦 Dependency Analysis</h2>
        <div class="grid">
            <div class="card">
                <h3>Dependencies Overview</h3>
                <div class="stat-grid" id="dep-stats"></div>
            </div>
            <div class="card">
                <h3>Production vs Development</h3>
                <div class="chart-container"><canvas id="depTypeChart"></canvas></div>
            </div>
            <div class="card">
                <h3>License Distribution</h3>
                <div class="chart-container"><canvas id="licenseChart"></canvas></div>
            </div>
        </div>
        
        <div class="card">
            <h3>📋 Package Details</h3>
            <div class="tabs">
                <button class="tab active" onclick="showTab('deps-ts', this)">📘 TypeScript</button>
                <button class="tab" onclick="showTab('deps-go', this)">🐹 Go</button>
                <button class="tab" onclick="showTab('deps-rust', this)">🦀 Rust</button>
            </div>
            <div id="deps-ts" class="tab-content active"><div class="package-list" id="pkg-list-ts"></div></div>
            <div id="deps-go" class="tab-content"><div class="package-list" id="pkg-list-go"></div></div>
            <div id="deps-rust" class="tab-content"><div class="package-list" id="pkg-list-rust"></div></div>
        </div>
        
        <!-- PSR-4 Analysis -->
        <h2>🎯 PSR-4 Autoloading Analysis</h2>
        <div class="grid-2">
            <div class="card">
                <h3>PSR-4 Compliance Stats</h3>
                <div class="stat-grid" id="psr4-stats"></div>
            </div>
            <div class="card">
                <h3>Compliance Rate</h3>
                <div class="chart-container"><canvas id="psr4Chart"></canvas></div>
            </div>
        </div>
        
        <div class="card">
            <h3>⚠️ PSR-4 Violations</h3>
            <div class="tabs">
                <button class="tab active" onclick="showTab('psr4-ts', this)">📘 TypeScript</button>
                <button class="tab" onclick="showTab('psr4-go', this)">🐹 Go</button>
                <button class="tab" onclick="showTab('psr4-rust', this)">🦀 Rust</button>
            </div>
            <div id="psr4-ts" class="tab-content active"><div id="violations-ts"></div></div>
            <div id="psr4-go" class="tab-content"><div id="violations-go"></div></div>
            <div id="psr4-rust" class="tab-content"><div id="violations-rust"></div></div>
        </div>
        
        <!-- Namespace Analysis -->
        <h2>🏷️ Namespace Detection</h2>
        <div class="grid-2">
            <div class="card">
                <h3>Namespace Statistics</h3>
                <div class="stat-grid" id="ns-stats"></div>
            </div>
            <div class="card">
                <h3>Top-Level Namespaces</h3>
                <div id="top-namespaces"></div>
            </div>
        </div>
        
        <!-- License Analysis -->
        <h2>📜 License Analysis</h2>
        <div class="card">
            <h3>License Distribution by Implementation</h3>
            <div class="grid-3" id="license-comparison"></div>
        </div>
        
        <!-- Raw Data -->
        <h2>📄 Raw Analysis Output</h2>
        <div class="card">
            <div class="tabs">
                <button class="tab active" onclick="showTab('raw-ts', this)">📘 TypeScript</button>
                <button class="tab" onclick="showTab('raw-go', this)">🐹 Go</button>
                <button class="tab" onclick="showTab('raw-rust', this)">🦀 Rust</button>
            </div>
            <div id="raw-ts" class="tab-content active">
                <select id="tool-select-ts" onchange="showToolOutput('ts', this.value)" style="margin-bottom: 1rem; padding: 0.5rem; background: #1e293b; color: #e2e8f0; border: 1px solid #475569; border-radius: 0.5rem;">
                    <option value="analyze_dependencies">analyze_dependencies</option>
                    <option value="analyze_psr4">analyze_psr4</option>
                    <option value="detect_namespaces">detect_namespaces</option>
                    <option value="audit_security">audit_security</option>
                    <option value="analyze_licenses">analyze_licenses</option>
                </select>
                <pre style="background: #0f172a; padding: 1rem; border-radius: 0.5rem; overflow: auto; max-height: 400px; font-size: 0.75rem;" id="raw-output-ts"></pre>
            </div>
            <div id="raw-go" class="tab-content">
                <select id="tool-select-go" onchange="showToolOutput('go', this.value)" style="margin-bottom: 1rem; padding: 0.5rem; background: #1e293b; color: #e2e8f0; border: 1px solid #475569; border-radius: 0.5rem;">
                    <option value="analyze_dependencies">analyze_dependencies</option>
                    <option value="analyze_psr4">analyze_psr4</option>
                    <option value="detect_namespaces">detect_namespaces</option>
                    <option value="audit_security">audit_security</option>
                    <option value="analyze_licenses">analyze_licenses</option>
                </select>
                <pre style="background: #0f172a; padding: 1rem; border-radius: 0.5rem; overflow: auto; max-height: 400px; font-size: 0.75rem;" id="raw-output-go"></pre>
            </div>
            <div id="raw-rust" class="tab-content">
                <select id="tool-select-rust" onchange="showToolOutput('rust', this.value)" style="margin-bottom: 1rem; padding: 0.5rem; background: #1e293b; color: #e2e8f0; border: 1px solid #475569; border-radius: 0.5rem;">
                    <option value="analyze_dependencies">analyze_dependencies</option>
                    <option value="analyze_psr4">analyze_psr4</option>
                    <option value="detect_namespaces">detect_namespaces</option>
                    <option value="audit_security">audit_security</option>
                    <option value="analyze_licenses">analyze_licenses</option>
                </select>
                <pre style="background: #0f172a; padding: 1rem; border-radius: 0.5rem; overflow: auto; max-height: 400px; font-size: 0.75rem;" id="raw-output-rust"></pre>
            </div>
        </div>
        
        <p class="timestamp" id="timestamp"></p>
    </div>
    
    <script>
HTMLHEADER

# Inject the benchmark results data
cat >> "$DASHBOARD_FILE" << EOF
        // Performance data
        const benchmarkResults = {
            TypeScript: $ts_results,
            Go: $go_results,
            Rust: $rust_results
        };
        
        // Analysis data
        const analysisData = $ANALYSIS_DATA;
        
        const timestamp = "$(date -Iseconds)";
EOF

# Add the JavaScript logic
cat >> "$DASHBOARD_FILE" << 'HTMLFOOTER'
        
        const implementations = ['TypeScript', 'Go', 'Rust'];
        const emojis = { TypeScript: '📘', Go: '🐹', Rust: '🦀' };
        const colors = { TypeScript: '#3b82f6', Go: '#10b981', Rust: '#f97316' };
        const colorsBg = { TypeScript: 'rgba(59, 130, 246, 0.7)', Go: 'rgba(16, 185, 129, 0.7)', Rust: 'rgba(249, 115, 22, 0.7)' };
        
        // Helper functions
        function findWinner(metric) {
            let winner = null, bestValue = Infinity;
            for (const impl of implementations) {
                const value = benchmarkResults[impl]?.[metric] || 0;
                if (value > 0 && value < bestValue) { bestValue = value; winner = impl; }
            }
            return { name: winner, value: bestValue };
        }
        
        function updateWinner(id, metric, unit) {
            const { name, value } = findWinner(metric);
            const el = document.getElementById(id);
            if (name && value < Infinity) {
                el.innerHTML = `<span class="winner-badge">${emojis[name]}</span><div><div class="winner-name">${name}</div><div class="winner-value">${value} ${unit}</div></div>`;
            } else {
                el.innerHTML = '<span class="empty-state">No data available</span>';
            }
        }
        
        function showTab(tabId, btn) {
            const container = btn.closest('.card');
            container.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            container.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
            btn.classList.add('active');
            document.getElementById(tabId).classList.add('active');
        }
        
        function showToolOutput(impl, tool) {
            const implMap = { ts: 'TypeScript', go: 'Go', rust: 'Rust' };
            const data = analysisData[implMap[impl]]?.[tool];
            const el = document.getElementById(`raw-output-${impl}`);
            el.textContent = data ? JSON.stringify(data, null, 2) : 'No data available';
        }
        
        // Initialize winners
        updateWinner('winner-startup', 'startup_ms', 'ms');
        updateWinner('winner-size', 'size_mb', 'MB');
        
        // Find fastest total analysis time
        let fastestAnalysis = { name: null, value: Infinity };
        for (const impl of implementations) {
            const data = benchmarkResults[impl] || {};
            const total = (data.analyze_dependencies || 0) + (data.analyze_psr4 || 0) + 
                         (data.detect_namespaces || 0) + (data.audit_security || 0) + (data.analyze_licenses || 0);
            if (total > 0 && total < fastestAnalysis.value) {
                fastestAnalysis = { name: impl, value: total };
            }
        }
        document.getElementById('winner-analysis').innerHTML = fastestAnalysis.name 
            ? `<span class="winner-badge">${emojis[fastestAnalysis.name]}</span><div><div class="winner-name">${fastestAnalysis.name}</div><div class="winner-value">${fastestAnalysis.value} ms total</div></div>`
            : '<span class="empty-state">No data available</span>';
        
        // Comparison table
        const metrics = [
            { key: 'startup_ms', label: 'Startup Time', unit: 'ms' },
            { key: 'size_mb', label: 'Binary Size', unit: 'MB' },
            { key: 'memory_mb', label: 'Memory (est.)', unit: 'MB' },
            { key: 'analyze_dependencies', label: 'Dependencies Analysis', unit: 'ms' },
            { key: 'analyze_psr4', label: 'PSR-4 Analysis', unit: 'ms' },
            { key: 'detect_namespaces', label: 'Namespace Detection', unit: 'ms' },
            { key: 'audit_security', label: 'Security Audit', unit: 'ms' },
            { key: 'analyze_licenses', label: 'License Analysis', unit: 'ms' }
        ];
        
        let tableHtml = '';
        for (const m of metrics) {
            const { name: winner } = findWinner(m.key);
            tableHtml += `<tr><td>${m.label}</td>`;
            for (const impl of implementations) {
                const value = benchmarkResults[impl]?.[m.key] || 0;
                const isBest = impl === winner && value > 0;
                tableHtml += `<td class="${impl.toLowerCase()}${isBest ? ' best' : ''}">${value || '-'} ${value ? m.unit : ''}</td>`;
            }
            tableHtml += '</tr>';
        }
        document.getElementById('comparison-tbody').innerHTML = tableHtml;
        
        // Chart defaults
        Chart.defaults.color = '#94a3b8';
        Chart.defaults.borderColor = '#334155';
        
        const chartOptions = {
            responsive: true,
            maintainAspectRatio: false,
            plugins: { legend: { display: false } },
            scales: { 
                y: { beginAtZero: true, grid: { color: '#334155' } }, 
                x: { grid: { display: false } } 
            }
        };
        
        // Performance charts
        function createBarChart(id, metric) {
            new Chart(document.getElementById(id), {
                type: 'bar',
                data: {
                    labels: implementations,
                    datasets: [{
                        data: implementations.map(i => benchmarkResults[i]?.[metric] || 0),
                        backgroundColor: implementations.map(i => colorsBg[i]),
                        borderColor: implementations.map(i => colors[i]),
                        borderWidth: 2,
                        borderRadius: 8
                    }]
                },
                options: chartOptions
            });
        }
        
        createBarChart('startupChart', 'startup_ms');
        createBarChart('sizeChart', 'size_mb');
        createBarChart('memoryChart', 'memory_mb');
        
        // Tool times chart (grouped bar)
        const tools = ['analyze_dependencies', 'analyze_psr4', 'detect_namespaces', 'audit_security', 'analyze_licenses'];
        const toolLabels = ['Dependencies', 'PSR-4', 'Namespaces', 'Security', 'Licenses'];
        
        new Chart(document.getElementById('toolTimesChart'), {
            type: 'bar',
            data: {
                labels: toolLabels,
                datasets: implementations.map(impl => ({
                    label: impl,
                    data: tools.map(t => benchmarkResults[impl]?.[t] || 0),
                    backgroundColor: colorsBg[impl],
                    borderColor: colors[impl],
                    borderWidth: 2,
                    borderRadius: 4
                }))
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { 
                    legend: { 
                        display: true,
                        position: 'top',
                        labels: { usePointStyle: true, padding: 20 }
                    }
                },
                scales: { 
                    y: { beginAtZero: true, grid: { color: '#334155' } }, 
                    x: { grid: { display: false } } 
                }
            }
        });
        
        // Dependency stats
        function getDeps(impl) {
            const deps = analysisData[impl]?.analyze_dependencies;
            if (!deps) return { prod: 0, dev: 0, total: 0 };
            const prod = deps.stats?.totalProduction || Object.keys(deps.production || {}).length;
            const dev = deps.stats?.totalDevelopment || Object.keys(deps.development || {}).length;
            return { prod, dev, total: prod + dev };
        }
        
        const tsD = getDeps('TypeScript'), goD = getDeps('Go'), rustD = getDeps('Rust');
        document.getElementById('dep-stats').innerHTML = `
            <div class="stat-item"><div class="stat-value">${tsD.total || goD.total || rustD.total}</div><div class="stat-label">Total Packages</div></div>
            <div class="stat-item"><div class="stat-value">${tsD.prod || goD.prod || rustD.prod}</div><div class="stat-label">Production</div></div>
            <div class="stat-item"><div class="stat-value">${tsD.dev || goD.dev || rustD.dev}</div><div class="stat-label">Development</div></div>
        `;
        
        // Dependency type chart
        new Chart(document.getElementById('depTypeChart'), {
            type: 'doughnut',
            data: {
                labels: ['Production', 'Development'],
                datasets: [{
                    data: [tsD.prod || goD.prod || rustD.prod, tsD.dev || goD.dev || rustD.dev],
                    backgroundColor: ['#3b82f6', '#8b5cf6'],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { position: 'bottom' } }
            }
        });
        
        // License chart
        function getLicenses(impl) {
            const deps = analysisData[impl]?.analyze_dependencies;
            if (!deps?.tree) return {};
            const licenses = {};
            deps.tree.forEach(pkg => {
                const lic = pkg.license || 'Unknown';
                licenses[lic] = (licenses[lic] || 0) + 1;
            });
            return licenses;
        }
        
        const licenses = getLicenses('TypeScript') || getLicenses('Go') || getLicenses('Rust') || {};
        const licenseLabels = Object.keys(licenses).slice(0, 8);
        const licenseData = licenseLabels.map(l => licenses[l]);
        
        new Chart(document.getElementById('licenseChart'), {
            type: 'pie',
            data: {
                labels: licenseLabels,
                datasets: [{
                    data: licenseData,
                    backgroundColor: ['#3b82f6', '#10b981', '#f97316', '#8b5cf6', '#ec4899', '#14b8a6', '#f59e0b', '#6366f1']
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { position: 'right', labels: { boxWidth: 12, padding: 8 } } }
            }
        });
        
        // Package lists
        function renderPackageList(impl, containerId) {
            const deps = analysisData[impl]?.analyze_dependencies;
            const el = document.getElementById(containerId);
            if (!deps?.tree || deps.tree.length === 0) {
                el.innerHTML = '<div class="empty-state">No package data available</div>';
                return;
            }
            el.innerHTML = deps.tree.slice(0, 50).map(pkg => `
                <div class="package-item">
                    <div>
                        <span class="package-name">${pkg.name}</span>
                        <span class="package-license">${pkg.license || ''}</span>
                    </div>
                    <span class="package-version">${pkg.version}</span>
                </div>
            `).join('');
        }
        
        renderPackageList('TypeScript', 'pkg-list-ts');
        renderPackageList('Go', 'pkg-list-go');
        renderPackageList('Rust', 'pkg-list-rust');
        
        // PSR-4 stats
        function getPsr4Stats(impl) {
            const psr4 = analysisData[impl]?.analyze_psr4;
            if (!psr4?.stats) return { mappings: 0, violations: 0, compliance: 100 };
            const s = psr4.stats;
            const compliance = s.total_mappings > 0 ? Math.round((1 - s.violation_count / s.total_mappings) * 100) : 100;
            return { mappings: s.total_mappings || 0, violations: s.violation_count || 0, compliance };
        }
        
        const psr4Ts = getPsr4Stats('TypeScript'), psr4Go = getPsr4Stats('Go'), psr4Rust = getPsr4Stats('Rust');
        const psr4 = psr4Ts.mappings > 0 ? psr4Ts : (psr4Go.mappings > 0 ? psr4Go : psr4Rust);
        
        document.getElementById('psr4-stats').innerHTML = `
            <div class="stat-item"><div class="stat-value">${psr4.mappings}</div><div class="stat-label">Mappings</div></div>
            <div class="stat-item"><div class="stat-value">${psr4.violations}</div><div class="stat-label">Violations</div></div>
            <div class="stat-item"><div class="stat-value" style="color: ${psr4.compliance > 90 ? '#10b981' : '#f97316'}">${psr4.compliance}%</div><div class="stat-label">Compliance</div></div>
        `;
        
        // PSR-4 compliance chart
        new Chart(document.getElementById('psr4Chart'), {
            type: 'doughnut',
            data: {
                labels: ['Compliant', 'Violations'],
                datasets: [{
                    data: [psr4.mappings - psr4.violations, psr4.violations],
                    backgroundColor: ['#10b981', '#ef4444'],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { position: 'bottom' } }
            }
        });
        
        // PSR-4 violations
        function renderViolations(impl, containerId) {
            const psr4 = analysisData[impl]?.analyze_psr4;
            const el = document.getElementById(containerId);
            if (!psr4?.violations || psr4.violations.length === 0) {
                el.innerHTML = '<div class="empty-state">No violations found ✓</div>';
                return;
            }
            el.innerHTML = psr4.violations.slice(0, 10).map(v => `
                <div class="violation">${typeof v === 'string' ? v : JSON.stringify(v)}</div>
            `).join('');
        }
        
        renderViolations('TypeScript', 'violations-ts');
        renderViolations('Go', 'violations-go');
        renderViolations('Rust', 'violations-rust');
        
        // Namespace stats
        function getNsStats(impl) {
            const ns = analysisData[impl]?.detect_namespaces;
            if (!ns) return { total: 0, topLevel: [] };
            return {
                total: ns.namespaces?.length || ns.total || 0,
                topLevel: ns.top_level || ns.namespaces?.slice(0, 10).map(n => n.namespace?.split('\\')[0]).filter((v, i, a) => a.indexOf(v) === i) || []
            };
        }
        
        const nsTs = getNsStats('TypeScript'), nsGo = getNsStats('Go'), nsRust = getNsStats('Rust');
        const ns = nsTs.total > 0 ? nsTs : (nsGo.total > 0 ? nsGo : nsRust);
        
        document.getElementById('ns-stats').innerHTML = `
            <div class="stat-item"><div class="stat-value">${ns.total}</div><div class="stat-label">Namespaces</div></div>
            <div class="stat-item"><div class="stat-value">${ns.topLevel.length}</div><div class="stat-label">Top-Level</div></div>
        `;
        
        document.getElementById('top-namespaces').innerHTML = ns.topLevel.length > 0 
            ? ns.topLevel.map(n => `<span class="namespace-tag">${n}</span>`).join('')
            : '<div class="empty-state">No namespace data available</div>';
        
        // License comparison
        const licCompHtml = implementations.map(impl => {
            const lics = getLicenses(impl);
            const entries = Object.entries(lics).sort((a, b) => b[1] - a[1]).slice(0, 6);
            const max = Math.max(...entries.map(e => e[1]), 1);
            return `
                <div>
                    <h4 style="margin-bottom: 1rem; color: ${colors[impl]}">${emojis[impl]} ${impl}</h4>
                    ${entries.length > 0 ? entries.map(([name, count]) => `
                        <div class="license-bar">
                            <span class="license-name">${name}</span>
                            <div class="license-bar-fill" style="width: ${count/max*100}%; background: ${colors[impl]}">${count}</div>
                        </div>
                    `).join('') : '<div class="empty-state">No data</div>'}
                </div>
            `;
        }).join('');
        document.getElementById('license-comparison').innerHTML = licCompHtml;
        
        // Initialize raw output
        showToolOutput('ts', 'analyze_dependencies');
        showToolOutput('go', 'analyze_dependencies');
        showToolOutput('rust', 'analyze_dependencies');
        
        // Timestamp
        document.getElementById('timestamp').textContent = `Generated: ${timestamp}`;
    </script>
</body>
</html>
HTMLFOOTER
fi  # End of if false block

echo -e "${GREEN}✓ Rich dashboard generated${NC}"

# ══════════════════════════════════════════════════════════════════════════════
# ASCII TERMINAL REPORT (Claude Code CLI style)
# ══════════════════════════════════════════════════════════════════════════════

print_ascii_report() {
    echo ""
    echo -e "${PURPLE}┌──────────────────────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${PURPLE}│${NC}  ${BOLD}dependency-buster${NC} // Analysis Complete                                      ${PURPLE}│${NC}"
    echo -e "${PURPLE}├──────────────────────────────────────────────────────────────────────────────┤${NC}"
    echo -e "${PURPLE}│${NC}  📦 ${BOLD}$PROJECT_NAME${NC}"
    
    # Format languages nicely
    local langs_str="${DETECTED_LANGS[*]}"
    if [ -z "$langs_str" ] || [ "$langs_str" = "Unknown" ]; then
        langs_str="Auto-detected"
    fi
    
    echo -e "${PURPLE}│${NC}     Languages: ${CYAN}${langs_str}${NC} • Package Manager: ${CYAN}$DETECTED_PKG_MANAGER${NC}"
    echo -e "${PURPLE}│${NC}     Files: $TOTAL_FILES • Analyzed: $(date +'%Y-%m-%d %H:%M:%S')"
    echo -e "${PURPLE}└──────────────────────────────────────────────────────────────────────────────┘${NC}"
    echo ""
    
    # Performance comparison
    echo -e "${BOLD}  ⚡ PERFORMANCE COMPARISON${NC}"
    echo -e "  ${DIM}─────────────────────────────────────────────────────────────────────────${NC}"
    printf "  ${DIM}%-20s${NC} ${BLUE}%-15s${NC} ${GREEN}%-15s${NC} ${RED}%-15s${NC}\n" "Metric" "TypeScript" "Go" "Rust"
    echo -e "  ${DIM}─────────────────────────────────────────────────────────────────────────${NC}"
    
    # Extract metrics from benchmark results
    ts_startup=$(echo "$ts_results" | grep -o '"startup_ms":[0-9]*' | cut -d: -f2)
    go_startup=$(echo "$go_results" | grep -o '"startup_ms":[0-9]*' | cut -d: -f2)
    rust_startup=$(echo "$rust_results" | grep -o '"startup_ms":[0-9]*' | cut -d: -f2)
    
    ts_size=$(echo "$ts_results" | grep -o '"size_mb":[0-9.]*' | cut -d: -f2)
    go_size=$(echo "$go_results" | grep -o '"size_mb":[0-9.]*' | cut -d: -f2)
    rust_size=$(echo "$rust_results" | grep -o '"size_mb":[0-9.]*' | cut -d: -f2)
    
    printf "  %-20s ${BLUE}%-15s${NC} ${GREEN}%-15s${NC} ${RED}%-15s${NC}\n" "Startup Time" "${ts_startup}ms" "${go_startup}ms" "${rust_startup}ms"
    printf "  %-20s ${BLUE}%-15s${NC} ${GREEN}%-15s${NC} ${RED}%-15s${NC}\n" "Binary Size" "${ts_size}MB" "${go_size}MB" "${rust_size}MB"
    
    # Tool execution times
    for tool in "${TOOLS[@]}"; do
        ts_time=$(echo "$ts_results" | grep -o "\"$tool\":[0-9]*" | cut -d: -f2)
        go_time=$(echo "$go_results" | grep -o "\"$tool\":[0-9]*" | cut -d: -f2)
        rust_time=$(echo "$rust_results" | grep -o "\"$tool\":[0-9]*" | cut -d: -f2)
        
        # Determine winner
        winner=""
        min_time=999999
        [ -n "$ts_time" ] && [ "$ts_time" -lt "$min_time" ] && min_time=$ts_time && winner="ts"
        [ -n "$go_time" ] && [ "$go_time" -lt "$min_time" ] && min_time=$go_time && winner="go"
        [ -n "$rust_time" ] && [ "$rust_time" -lt "$min_time" ] && winner="rust"
        
        ts_fmt="${ts_time:-0}ms"
        go_fmt="${go_time:-0}ms"
        rust_fmt="${rust_time:-0}ms"
        
        # Mark winner with checkmark
        [ "$winner" = "ts" ] && ts_fmt="${ts_fmt} ✓"
        [ "$winner" = "go" ] && go_fmt="${go_fmt} ✓"
        [ "$winner" = "rust" ] && rust_fmt="${rust_fmt} ✓"

        tool_short=$(echo "$tool" | sed 's/analyze_//' | sed 's/detect_//' | sed 's/audit_//')
        echo -e "  $(printf '%-20s' "$tool_short") ${BLUE}$(printf '%-15s' "$ts_fmt")${NC} ${GREEN}$(printf '%-15s' "$go_fmt")${NC} ${RED}$(printf '%-15s' "$rust_fmt")${NC}"
    done
    echo -e "  ${DIM}─────────────────────────────────────────────────────────────────────────${NC}"
    echo ""
    
    # Dependency Tree (ASCII)
    echo -e "${BOLD}  📦 DEPENDENCY TREE${NC}"
    echo -e "  ${DIM}─────────────────────────────────────────────────────────────────────────${NC}"
    
    # Try to extract top dependencies from analysis data
    local tree_printed=false
    if [ -f "$RESULTS_DIR/TypeScript_analyze_dependencies.json" ]; then
        echo -e "  ${DIM}(from TypeScript analyzer)${NC}"
        node -e "
        try {
            const fs = require('fs');
            const data = JSON.parse(fs.readFileSync('$RESULTS_DIR/TypeScript_analyze_dependencies.json', 'utf8'));
            const result = data.result?.content?.[0]?.text;
            if (result) {
                const parsed = JSON.parse(result);
                const prod = parsed.production || {};
                const tree = parsed.tree || [];
                
                // Use tree if available, otherwise use production deps
                if (tree.length > 0) {
                    const topDeps = tree.slice(0, 10);
                    topDeps.forEach((dep, i) => {
                        const isLast = i === topDeps.length - 1;
                        const prefix = isLast ? '  └── ' : '  ├── ';
                        console.log(prefix + dep.name + ' @ ' + dep.version);
                    });
                    if (tree.length > 10) {
                        console.log('  ... and ' + (tree.length - 10) + ' more dependencies');
                    }
                } else if (Object.keys(prod).length > 0) {
                    const deps = Object.entries(prod).slice(0, 10);
                    deps.forEach(([name, version], i) => {
                        const isLast = i === deps.length - 1;
                        const prefix = isLast ? '  └── ' : '  ├── ';
                        console.log(prefix + name + ' @ ' + version);
                    });
                    if (Object.keys(prod).length > 10) {
                        console.log('  ... and ' + (Object.keys(prod).length - 10) + ' more');
                    }
                } else {
                    console.log('  (No dependencies found in composer.json)');
                }
            } else {
                console.log('  (Analysis result empty)');
            }
        } catch(e) { 
            console.log('  (Parse error: ' + e.message + ')'); 
        }
        " 2>/dev/null
        tree_printed=true
    fi
    
    if [ "$tree_printed" = false ]; then
        echo -e "  ${DIM}(Run analysis first to see dependency tree)${NC}"
    fi
    echo ""
    
    # Security Summary
    echo -e "${BOLD}  🔒 SECURITY SUMMARY${NC}"
    echo -e "  ${DIM}─────────────────────────────────────────────────────────────────────────${NC}"
    if [ -f "$RESULTS_DIR/TypeScript_audit_security.json" ]; then
        node -e "
        try {
            const data = require('$RESULTS_DIR/TypeScript_audit_security.json');
            const result = data.result?.content?.[0]?.text;
            if (result) {
                const parsed = JSON.parse(result);
                const total = parsed.summary?.total || 0;
                const critical = parsed.summary?.critical || 0;
                const high = parsed.summary?.high || 0;
                const medium = parsed.summary?.medium || 0;
                const low = parsed.summary?.low || 0;
                const safe = total === 0 ? '✓ No vulnerabilities found' : '';
                console.log('  Total Packages: ' + total);
                if (critical > 0) console.log('  ${RED}● Critical: ' + critical + '${NC}');
                if (high > 0) console.log('  ${YELLOW}● High: ' + high + '${NC}');
                if (medium > 0) console.log('  ● Medium: ' + medium);
                if (low > 0) console.log('  ● Low: ' + low);
                if (safe) console.log('  ${GREEN}' + safe + '${NC}');
            }
        } catch(e) { console.log('  (Could not parse security data)'); }
        " 2>/dev/null || echo "  (Could not parse security data)"
    fi
    echo ""
    
    # Agent Suggestions (Compliance Check)
    echo -e "${BOLD}  🤖 AGENT SUGGESTIONS${NC}"
    echo -e "  ${DIM}─────────────────────────────────────────────────────────────────────────${NC}"
    if [ -f "$RESULTS_DIR/TypeScript_audit_security.json" ]; then
        node -e "
        try {
            const data = require('$RESULTS_DIR/TypeScript_audit_security.json');
            const result = data.result?.content?.[0]?.text;
            if (result) {
                const parsed = JSON.parse(result);
                const issues = parsed.vulnerabilities || [];
                const outdated = parsed.outdated || [];
                
                if (issues.length === 0 && outdated.length === 0) {
                    console.log('  🟢 ✓ No compliance issues found');
                } else {
                    if (issues.length > 0) {
                        console.log('  🔴 ✗ ' + issues.length + ' security vulnerabilities');
                        issues.slice(0, 3).forEach(i => {
                            console.log('     • ' + (i.package || i.name) + ': ' + (i.severity || 'unknown'));
                        });
                    }
                    if (outdated.length > 0) {
                        console.log('  🟡 ⚠ ' + outdated.length + ' outdated packages');
                    }
                }
            } else {
                console.log('  ℹ Run check_compliance tool for detailed suggestions');
            }
        } catch(e) { 
            console.log('  ℹ Run get_agent_suggestions for AI agent integration');
        }
        " 2>/dev/null || echo "  ℹ Use MCP tools for detailed compliance checking"
    else
        echo "  ℹ Run analysis to get agent suggestions"
    fi
    echo ""
    echo -e "  ${DIM}To get full suggestions for AI agents:${NC}"
    echo "  echo '{\"jsonrpc\":\"2.0\",\"method\":\"tools/call\",\"params\":{\"name\":\"get_agent_suggestions\",\"arguments\":{\"repo_path\":\"$TARGET_REPO\"}},\"id\":1}' | node dpb-mcp-typescript/build/server.js"
    echo ""
    
    # Generated Files
    echo -e "${BOLD}  📄 GENERATED FILES${NC}"
    echo -e "  ${DIM}─────────────────────────────────────────────────────────────────────────${NC}"
    echo "  JSON Report:  $REPORT_FILE"
    echo "  Dashboard:    $DASHBOARD_FILE"
    echo ""
    
    # Quick Actions
    echo -e "${BOLD}  🚀 QUICK ACTIONS${NC}"
    echo -e "  ${DIM}─────────────────────────────────────────────────────────────────────────${NC}"
    echo "  View Dashboard:  open $DASHBOARD_FILE"
    echo "  View JSON:       cat $REPORT_FILE | jq ."
    echo "  Agent Suggestions: (see command above)"
    echo ""
}

# Print the ASCII report
print_ascii_report
