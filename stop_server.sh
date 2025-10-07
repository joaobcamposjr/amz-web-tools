#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Stop Server
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
PID_DIR="/d02/pids"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🛑 AMZ Web Tools - Stopping Services${NC}"
echo -e "${BLUE}=====================================${NC}"

# =============================================
# Stop Services by PID
# =============================================
echo -e "${YELLOW}🔄 Stopping services by PID...${NC}"

# Stop Backend
if [ -f "$PID_DIR/backend.pid" ]; then
    BACKEND_PID=$(cat "$PID_DIR/backend.pid")
    if kill -0 "$BACKEND_PID" 2>/dev/null; then
        echo "🛑 Stopping backend (PID: $BACKEND_PID)"
        kill -TERM "$BACKEND_PID" 2>/dev/null || true
        sleep 2
        kill -KILL "$BACKEND_PID" 2>/dev/null || true
        echo -e "${GREEN}✅ Backend stopped${NC}"
    else
        echo "⚠️ Backend PID file exists but process not running"
    fi
    rm -f "$PID_DIR/backend.pid"
else
    echo "⚠️ Backend PID file not found"
fi

# Stop Frontend
if [ -f "$PID_DIR/frontend.pid" ]; then
    FRONTEND_PID=$(cat "$PID_DIR/frontend.pid")
    if kill -0 "$FRONTEND_PID" 2>/dev/null; then
        echo "🛑 Stopping frontend (PID: $FRONTEND_PID)"
        kill -TERM "$FRONTEND_PID" 2>/dev/null || true
        sleep 2
        kill -KILL "$FRONTEND_PID" 2>/dev/null || true
        echo -e "${GREEN}✅ Frontend stopped${NC}"
    else
        echo "⚠️ Frontend PID file exists but process not running"
    fi
    rm -f "$PID_DIR/frontend.pid"
else
    echo "⚠️ Frontend PID file not found"
fi

# =============================================
# Kill by Port
# =============================================
echo -e "${YELLOW}🔄 Killing processes by port...${NC}"

kill_port() {
    local port=$1
    local pids=$(lsof -ti:$port 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo "🔫 Killing processes on port $port: $pids"
        kill -9 $pids 2>/dev/null || true
        echo -e "${GREEN}✅ Port $port cleared${NC}"
    else
        echo "✅ Port $port is already free"
    fi
}

kill_port 8080  # Backend
kill_port 3000  # Frontend
kill_port 3001  # Frontend alt
kill_port 3002  # Frontend alt
kill_port 3003  # Frontend alt

# =============================================
# Kill by Process Name
# =============================================
echo -e "${YELLOW}🔄 Killing processes by name...${NC}"

kill_process() {
    local process=$1
    local pids=$(pgrep -f "$process" 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo "🔫 Killing $process processes: $pids"
        kill -9 $pids 2>/dev/null || true
        echo -e "${GREEN}✅ $process processes killed${NC}"
    else
        echo "✅ No $process processes found"
    fi
}

kill_process "go run main.go"
kill_process "next dev"
kill_process "next start"
kill_process "node.*next"
kill_process "amz-web-tools"

# =============================================
# Final Status
# =============================================
echo ""
echo -e "${GREEN}🎉 All services stopped successfully!${NC}"
echo ""
echo -e "${BLUE}🔍 Verify cleanup:${NC}"
echo -e "   Check ports:  lsof -i :8080 -i :3000"
echo -e "   Check processes: ps aux | grep -E '(backend|next)'"
echo ""
