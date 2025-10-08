#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Quick Fix Login
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
PID_DIR="/d02/pids"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${RED}🚨 QUICK FIX LOGIN ISSUE${NC}"
echo ""

cd "$BASE"

# Check if backend is running
echo -e "${YELLOW}🔍 Checking if backend is running...${NC}"
if pgrep -f "backend" > /dev/null; then
    echo -e "${GREEN}✅ Backend process found${NC}"
    pgrep -f "backend"
else
    echo -e "${RED}❌ Backend process NOT found${NC}"
fi

# Check if port 8080 is in use
echo ""
echo -e "${YELLOW}🔍 Checking port 8080...${NC}"
if netstat -tlnp 2>/dev/null | grep ":8080" > /dev/null; then
    echo -e "${GREEN}✅ Port 8080 is in use${NC}"
    netstat -tlnp 2>/dev/null | grep ":8080"
else
    echo -e "${RED}❌ Port 8080 is NOT in use${NC}"
fi

# Kill all backend processes
echo ""
echo -e "${YELLOW}🛑 Killing all backend processes...${NC}"
pkill -f "backend" 2>/dev/null || true
pkill -f "amz-web-tools" 2>/dev/null || true
fuser -k 8080/tcp 2>/dev/null || true
sleep 3

# Check if .env is readable
echo ""
echo -e "${YELLOW}🔍 Checking .env file...${NC}"
if [ -f ".env" ]; then
    if [ -r ".env" ]; then
        echo -e "${GREEN}✅ .env file is readable${NC}"
        echo "Oracle Host: $(grep '^ORACLE_HOST=' .env | cut -d'=' -f2)"
    else
        echo -e "${RED}❌ .env file is NOT readable (might be locked)${NC}"
        echo -e "${YELLOW}🔓 Making .env readable...${NC}"
        chmod 644 .env
    fi
else
    echo -e "${RED}❌ .env file NOT found${NC}"
fi

# Start backend
echo ""
echo -e "${YELLOW}🚀 Starting backend...${NC}"

# Set environment variables
export $(grep -v '^#' .env | xargs) 2>/dev/null || true

# Start backend
cd "$BASE/backend"
nohup ./bin/backend > "$LOGS/backend.log" 2>&1 &
echo $! > "$LOGS/backend.pid"

sleep 5

# Check if backend started
echo ""
echo -e "${YELLOW}🔍 Checking if backend started...${NC}"
if pgrep -f "backend" > /dev/null; then
    echo -e "${GREEN}✅ Backend started successfully${NC}"
else
    echo -e "${RED}❌ Backend failed to start${NC}"
fi

# Test backend health
echo ""
echo -e "${YELLOW}🔍 Testing backend health...${NC}"
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✅ Backend health check passed${NC}"
else
    echo -e "${RED}❌ Backend health check failed${NC}"
fi

# Test external access
echo ""
echo -e "${YELLOW}🔍 Testing external access...${NC}"
if curl -s http://52.206.225.24:8080/health > /dev/null; then
    echo -e "${GREEN}✅ External backend access working${NC}"
else
    echo -e "${RED}❌ External backend access failed${NC}"
fi

# Show recent logs
echo ""
echo -e "${YELLOW}📄 Recent backend logs:${NC}"
tail -10 "$LOGS/backend.log" 2>/dev/null || echo "No backend logs found"

echo ""
echo -e "${BLUE}🔧 Quick fix completed!${NC}"
echo ""
echo -e "${YELLOW}Test login now:${NC}"
echo "http://52.206.225.24:3000"
