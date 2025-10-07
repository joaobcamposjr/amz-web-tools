#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Server Initialization
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
CACHE="/d02/.cache"
PID_DIR="/d02/pids"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 AMZ Web Tools - Server Initialization${NC}"
echo -e "${BLUE}=========================================${NC}"

# =============================================
# 1. Setup Environment Variables
# =============================================
echo -e "${YELLOW}🔧 Setting up environment variables...${NC}"

# Oracle Environment
export ORACLE_HOME=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH="$ORACLE_HOME:$LD_LIBRARY_PATH"
export PATH="$ORACLE_HOME:$PATH"

# Go Environment
export GOCACHE="$CACHE/go-build"
export GOMODCACHE="/d02/go/pkg/mod"
export CGO_ENABLED=1

# Node.js Environment
export NODE_ENV=production
export NEXT_PUBLIC_API_URL=http://52.206.225.24:8080/api/v1
export SERVER_HOST=0.0.0.0
export FRONTEND_HOST=0.0.0.0

# Log environment setup
echo "✅ Oracle Home: $ORACLE_HOME"
echo "✅ Go Cache: $GOCACHE"
echo "✅ Go Module Cache: $GOMODCACHE"

# =============================================
# 2. Create Directories
# =============================================
echo -e "${YELLOW}📁 Creating directories...${NC}"
mkdir -p "$LOGS" "$CACHE/go-build" "/d02/go/pkg/mod" "$PID_DIR"
echo "✅ Directories created"

# =============================================
# 3. Kill Existing Processes
# =============================================
echo -e "${YELLOW}🔄 Killing existing processes...${NC}"

# Kill by port
kill_port() {
    local port=$1
    local pids=$(lsof -ti:$port 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo "🔫 Killing processes on port $port: $pids"
        kill -9 $pids 2>/dev/null || true
    else
        echo "✅ Port $port is free"
    fi
}

# Kill by process name
kill_process() {
    local process=$1
    local pids=$(pgrep -f "$process" 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo "🔫 Killing $process processes: $pids"
        kill -9 $pids 2>/dev/null || true
    else
        echo "✅ No $process processes found"
    fi
}

# Kill all relevant processes
kill_port 8080  # Backend
kill_port 3000  # Frontend
kill_port 3001  # Frontend alt
kill_port 3002  # Frontend alt
kill_port 3003  # Frontend alt

kill_process "go run main.go"
kill_process "next dev"
kill_process "next start"
kill_process "node.*next"
kill_process "amz-web-tools"

sleep 2
echo "✅ Process cleanup completed"

# =============================================
# 4. Clean Cache and Build Artifacts
# =============================================
echo -e "${YELLOW}🗑️ Cleaning cache and build artifacts...${NC}"

cd "$BASE"

# Clean Next.js cache
if [ -d .next ]; then
    echo "🧹 Removing .next directory"
    sudo rm -rf .next || rm -rf .next
fi

# Clean Node.js cache
if [ -d node_modules/.cache ]; then
    echo "🧹 Removing node_modules/.cache"
    rm -rf node_modules/.cache
fi

# Clean Go cache
echo "🧹 Cleaning Go cache"
go clean -cache 2>/dev/null || true
go clean -modcache 2>/dev/null || true

echo "✅ Cache cleanup completed"

# =============================================
# 5. Fix Permissions
# =============================================
echo -e "${YELLOW}🔧 Fixing permissions...${NC}"

# Fix ownership
sudo chown -R ec2-user:ec2-user "$BASE" 2>/dev/null || true
sudo chown -R ec2-user:ec2-user "$LOGS" 2>/dev/null || true
sudo chown -R ec2-user:ec2-user "$CACHE" 2>/dev/null || true

# Fix permissions
chmod -R 755 "$BASE" 2>/dev/null || true
chmod -R 755 "$LOGS" 2>/dev/null || true
chmod -R 755 "$CACHE" 2>/dev/null || true

echo "✅ Permissions fixed"

# =============================================
# 6. Install Dependencies
# =============================================
echo -e "${YELLOW}📦 Installing dependencies...${NC}"

# Go dependencies
echo "📦 Installing Go dependencies..."
cd "$BASE/backend"
go mod download
go mod tidy
echo "✅ Go dependencies installed"

# Node.js dependencies
echo "📦 Installing Node.js dependencies..."
cd "$BASE"
npm install --production=false
echo "✅ Node.js dependencies installed"

# =============================================
# 7. Build Applications
# =============================================
echo -e "${YELLOW}🔨 Building applications...${NC}"

# Build Go backend
echo "🔨 Building Go backend..."
cd "$BASE/backend"
go build -o "$BASE/bin/backend" .
echo "✅ Backend built successfully"

# Build Next.js frontend
echo "🔨 Building Next.js frontend..."
cd "$BASE"
npm run build
echo "✅ Frontend built successfully"

# =============================================
# 8. Start Services
# =============================================
echo -e "${YELLOW}🚀 Starting services...${NC}"

# Start Backend
echo "🚀 Starting backend..."
cd "$BASE"
nohup ./bin/backend > "$LOGS/backend.log" 2>&1 & echo $! > "$PID_DIR/backend.pid"
sleep 3

# Check if backend started
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✅ Backend started successfully on port 8080${NC}"
else
    echo -e "${RED}❌ Backend failed to start${NC}"
    echo "📋 Backend logs:"
    tail -10 "$LOGS/backend.log"
    exit 1
fi

# Start Frontend
echo "🚀 Starting frontend..."
cd "$BASE"
nohup node .next/standalone/server.js --hostname 0.0.0.0 --port 3000 > "$LOGS/frontend.log" 2>&1 & echo $! > "$PID_DIR/frontend.pid"
sleep 5

# Check if frontend started
if curl -s http://localhost:3000 > /dev/null; then
    echo -e "${GREEN}✅ Frontend started successfully on port 3000${NC}"
else
    echo -e "${RED}❌ Frontend failed to start${NC}"
    echo "📋 Frontend logs:"
    tail -10 "$LOGS/frontend.log"
    exit 1
fi

# =============================================
# 9. Health Checks
# =============================================
echo -e "${YELLOW}🏥 Running health checks...${NC}"

# Backend health check
echo "🏥 Checking backend health..."
BACKEND_HEALTH=$(curl -s http://localhost:8080/health || echo "FAILED")
if [ "$BACKEND_HEALTH" = '{"status":"ok"}' ]; then
    echo -e "${GREEN}✅ Backend health check passed${NC}"
else
    echo -e "${RED}❌ Backend health check failed: $BACKEND_HEALTH${NC}"
fi

# Frontend health check
echo "🏥 Checking frontend health..."
FRONTEND_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 || echo "FAILED")
if [ "$FRONTEND_HEALTH" = "200" ]; then
    echo -e "${GREEN}✅ Frontend health check passed${NC}"
else
    echo -e "${RED}❌ Frontend health check failed: HTTP $FRONTEND_HEALTH${NC}"
fi

# =============================================
# 10. Final Status
# =============================================
echo ""
echo -e "${GREEN}🎉 AMZ Web Tools Initialized Successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo -e "${BLUE}🌐 Services:${NC}"
echo -e "   Frontend: http://52.206.225.24:3000"
echo -e "   Backend:  http://52.206.225.24:8080"
echo -e "   Health:   http://52.206.225.24:8080/health"
echo -e "   API:      http://52.206.225.24:8080/api/v1"
echo ""
echo -e "${BLUE}📝 Logs:${NC}"
echo -e "   Backend:  tail -f $LOGS/backend.log"
echo -e "   Frontend: tail -f $LOGS/frontend.log"
echo ""
echo -e "${BLUE}🛑 Stop Services:${NC}"
echo -e "   ./stop_server.sh"
echo ""
echo -e "${BLUE}📊 Process IDs:${NC}"
echo -e "   Backend PID:  $(cat $PID_DIR/backend.pid 2>/dev/null || echo 'N/A')"
echo -e "   Frontend PID: $(cat $PID_DIR/frontend.pid 2>/dev/null || echo 'N/A')"
echo ""
echo -e "${BLUE}🔍 Debug Commands:${NC}"
echo -e "   Check ports:  lsof -i :8080 -i :3000"
echo -e "   Check processes: ps aux | grep -E '(backend|next)'"
echo -e "   Check Oracle:  echo \$ORACLE_HOME"
echo -e "   Check Go:      go version"
echo -e "   Check Node:    node --version"
echo ""
