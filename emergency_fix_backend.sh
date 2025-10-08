#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Emergency Backend Fix
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${RED}🚨 EMERGENCY BACKEND FIX${NC}"
echo ""

cd "$BASE"

# AGGRESSIVE CLEANUP
echo -e "${YELLOW}🧹 AGGRESSIVE CLEANUP...${NC}"

# Kill everything
pkill -f "backend" 2>/dev/null || true
pkill -f "amz-web-tools" 2>/dev/null || true
pkill -f "go run" 2>/dev/null || true

# Kill ports
fuser -k 8080/tcp 2>/dev/null || true
lsof -ti:8080 | xargs kill -9 2>/dev/null || true

sleep 3

# Make sure .env is readable
echo -e "${YELLOW}🔓 Fixing .env permissions...${NC}"
chmod 644 .env 2>/dev/null || true

# Verify .env exists and is readable
if [ ! -r ".env" ]; then
    echo -e "${RED}❌ .env not readable, creating basic one...${NC}"
    cat > .env << 'EOF'
# Basic config for emergency
DB_HOST=54.204.42.134
DB_PORT=1433
DB_USER=sa
DB_PASSWORD=321@Mudar@7089341@
DB_NAME=integration
DB_SSL_MODE=disable
SERVER_PORT=8080
JWT_SECRET=amz-web-tools-secret-key-2024
ORACLE_HOST=164.152.40.38
ORACLE_PORT=1521
ORACLE_USER=dashjc
ORACLE_PASSWORD=@Joao1225
ORACLE_SERVICE=nbs
ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
PG_HOST=shared-codako-nlb-3f3ad9f6c528c4a6.elb.us-east-1.amazonaws.com
PG_PORT=5433
PG_USER=codako_bi
PG_PASSWORD=lNkIXKc9CQuyv28B
PG_DATABASE=codako_bi
CORS_ALLOWED_ORIGINS=http://52.206.225.24:3000,http://52.206.225.24
NEXT_PUBLIC_API_URL=http://52.206.225.24:8080/api/v1
SERVER_HOST=0.0.0.0
FRONTEND_HOST=0.0.0.0
EOF
fi

# Check if backend binary exists
echo -e "${YELLOW}🔍 Checking backend binary...${NC}"
if [ -f "backend/bin/backend" ]; then
    echo -e "${GREEN}✅ Backend binary found${NC}"
else
    echo -e "${RED}❌ Backend binary NOT found${NC}"
    echo -e "${YELLOW}🔨 Building backend...${NC}"
    cd backend
    go build -o bin/backend .
    cd ..
fi

# Start backend with explicit config
echo -e "${YELLOW}🚀 Starting backend...${NC}"

# Load environment
source .env 2>/dev/null || true

# Set explicit environment variables
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080
export ORACLE_HOST=164.152.40.38
export ORACLE_USER=dashjc
export ORACLE_SERVICE=nbs
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7

cd backend
nohup ./bin/backend > "$LOGS/backend.log" 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > "$LOGS/backend.pid"

sleep 5

# Check if backend is running
echo -e "${YELLOW}🔍 Checking backend status...${NC}"
if ps -p $BACKEND_PID > /dev/null; then
    echo -e "${GREEN}✅ Backend process running (PID: $BACKEND_PID)${NC}"
else
    echo -e "${RED}❌ Backend process died${NC}"
    echo -e "${YELLOW}📄 Backend logs:${NC}"
    tail -20 "$LOGS/backend.log"
    exit 1
fi

# Test port 8080
echo -e "${YELLOW}🔍 Testing port 8080...${NC}"
if netstat -tlnp 2>/dev/null | grep ":8080" > /dev/null; then
    echo -e "${GREEN}✅ Port 8080 is listening${NC}"
    netstat -tlnp 2>/dev/null | grep ":8080"
else
    echo -e "${RED}❌ Port 8080 is NOT listening${NC}"
fi

# Test health endpoint
echo -e "${YELLOW}🔍 Testing health endpoint...${NC}"
for i in {1..5}; do
    if curl -s http://localhost:8080/health > /dev/null; then
        echo -e "${GREEN}✅ Health check passed${NC}"
        break
    else
        echo -e "${YELLOW}⏳ Attempt $i/5 failed, waiting...${NC}"
        sleep 2
    fi
done

# Test external access
echo -e "${YELLOW}🔍 Testing external access...${NC}"
if curl -s http://52.206.225.24:8080/health > /dev/null; then
    echo -e "${GREEN}✅ External access working${NC}"
else
    echo -e "${RED}❌ External access failed${NC}"
    echo -e "${YELLOW}📄 Recent logs:${NC}"
    tail -10 "$LOGS/backend.log"
fi

echo ""
echo -e "${BLUE}🔧 Emergency fix completed!${NC}"
echo -e "${GREEN}Try login now: http://52.206.225.24:3000${NC}"
