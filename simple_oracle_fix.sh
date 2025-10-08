#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Simple Oracle Fix
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 SIMPLE ORACLE FIX${NC}"
echo ""

cd "$BASE"

# Stop backend
echo -e "${YELLOW}🛑 Stopping backend...${NC}"
pkill -f "backend" 2>/dev/null || true
sleep 3

# Build backend
echo -e "${YELLOW}🔨 Building backend...${NC}"
cd backend

# Set Oracle environment
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}

# Build
go build -o bin/backend .

if [ -f "bin/backend" ]; then
    echo -e "${GREEN}✅ Backend built successfully${NC}"
else
    echo -e "${RED}❌ Backend build failed${NC}"
    exit 1
fi

cd ..

# Start backend with Oracle config
echo -e "${YELLOW}🚀 Starting backend with Oracle config...${NC}"

# Set Oracle environment variables
export ORACLE_HOST=164.152.40.38
export ORACLE_PORT=1521
export ORACLE_USER=dashjc
export ORACLE_PASSWORD=@Joao1225
export ORACLE_SERVICE=nbs
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export ORACLE_HOME=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

cd backend
nohup ./bin/backend > "$LOGS/backend.log" 2>&1 &
echo $! > "$LOGS/backend.pid"

sleep 5

# Check if backend started
if pgrep -f "backend" > /dev/null; then
    echo -e "${GREEN}✅ Backend started successfully${NC}"
else
    echo -e "${RED}❌ Backend failed to start${NC}"
    tail -10 "$LOGS/backend.log"
    exit 1
fi

# Test health
echo -e "${YELLOW}🔍 Testing backend health...${NC}"
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✅ Backend health check passed${NC}"
else
    echo -e "${RED}❌ Backend health check failed${NC}"
fi

echo ""
echo -e "${BLUE}🔧 SIMPLE ORACLE FIX COMPLETED!${NC}"
echo ""
echo -e "${GREEN}Test now:${NC}"
echo "1. Login: http://52.206.225.24:3000"
echo "2. Stock search with any SKU"
echo "3. Check logs: tail -f $LOGS/backend.log"
