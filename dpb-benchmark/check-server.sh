#!/bin/bash
# Check if dashboard server is running

PORT=8080
SERVER_BINARY="dashboard-server"

echo "Checking dashboard server status..."
echo ""

# Check if port is in use
if lsof -ti:$PORT > /dev/null 2>&1; then
    PID=$(lsof -ti:$PORT)
    PROCESS=$(ps -p $PID -o comm= 2>/dev/null || echo "unknown")
    echo "✓ Server is running"
    echo "  Port: $PORT"
    echo "  PID: $PID"
    echo "  Process: $PROCESS"
    echo ""
    echo "Dashboard URL: http://localhost:$PORT"
    echo ""
    echo "To stop the server:"
    echo "  kill $PID"
    echo "  or: lsof -ti:$PORT | xargs kill"
else
    echo "✗ Server is not running"
    echo ""
    echo "To start the server:"
    echo "  cd dpb-benchmark"
    echo "  make serve"
    echo ""
    echo "Or manually:"
    echo "  cd dpb-benchmark/server"
    echo "  go build -o dashboard-server ."
    echo "  ./dashboard-server"
fi

echo ""
echo "Checking for server binary..."
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BINARY_PATH="$SCRIPT_DIR/server/dashboard-server"
if [ -f "$BINARY_PATH" ]; then
    echo "✓ Server binary found"
    BINARY_SIZE=$(du -h "$BINARY_PATH" | cut -f1)
    BINARY_DATE=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" "$BINARY_PATH" 2>/dev/null || stat -c "%y" "$BINARY_PATH" 2>/dev/null | cut -d' ' -f1-2)
    echo "  Size: $BINARY_SIZE"
    echo "  Modified: $BINARY_DATE"
    echo "  Path: $BINARY_PATH"
else
    echo "✗ Server binary not found at expected location"
    echo "  Expected: $BINARY_PATH"
    echo "  Build it with: cd $SCRIPT_DIR && make build"
fi
