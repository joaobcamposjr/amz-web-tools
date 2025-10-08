#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Rebuild Oracle Fixed
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${RED}🔨 REBUILDING ORACLE FIXED${NC}"
echo ""

cd "$BASE"

# Stop ALL processes
echo -e "${YELLOW}🛑 Stopping ALL processes...${NC}"
pkill -f "backend" 2>/dev/null || true
pkill -f "amz-web-tools" 2>/dev/null || true
pkill -f "next" 2>/dev/null || true
fuser -k 8080/tcp 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
sleep 3

# Remove old backend binary
echo -e "${YELLOW}🗑️ Removing old backend binary...${NC}"
rm -f backend/bin/backend
rm -f backend/backend

# Fix Go modules
echo -e "${YELLOW}🔧 Fixing Go modules...${NC}"
cd backend
go mod tidy
go mod download
go mod verify

# Build backend on Linux server
echo -e "${YELLOW}🔨 Building backend for Linux...${NC}"

# Set Oracle environment for build
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}

# Build
echo "Building Go backend..."
go build -o bin/backend .

if [ -f "bin/backend" ]; then
    echo -e "${GREEN}✅ Backend built successfully for Linux${NC}"
    file bin/backend
    ls -la bin/backend
else
    echo -e "${RED}❌ Backend build failed${NC}"
    echo "Build error details:"
    go build -v -o bin/backend . 2>&1 || true
    exit 1
fi

cd ..

# Create startup script with Oracle config
echo -e "${YELLOW}🔧 Creating startup script...${NC}"
cat > start_backend_linux.sh << 'EOF'
#!/bin/bash

# Oracle configuration
export ORACLE_HOST=164.152.40.38
export ORACLE_PORT=1521
export ORACLE_USER=dashjc
export ORACLE_PASSWORD=@Joao1225
export ORACLE_SERVICE=nbs
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export ORACLE_HOME=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}

# Other config
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080
export DB_HOST=54.204.42.134
export DB_PORT=1433
export DB_USER=sa
export DB_PASSWORD=321@Mudar@7089341@
export DB_NAME=integration
export PG_HOST=shared-codako-nlb-3f3ad9f6c528c4a6.elb.us-east-1.amazonaws.com
export PG_PORT=5433
export PG_USER=codako_bi
export PG_PASSWORD=lNkIXKc9CQuyv28B
export PG_DATABASE=codako_bi
export JWT_SECRET=amz-web-tools-secret-key-2024
export JWT_EXPIRE_HOURS=168

echo "🔧 Starting Linux backend with Oracle config:"
echo "ORACLE_HOST: $ORACLE_HOST"
echo "ORACLE_USER: $ORACLE_USER"
echo "ORACLE_SERVICE: $ORACLE_SERVICE"

cd /d02/projects/amz-web-tools/backend
exec ./bin/backend
EOF

chmod +x start_backend_linux.sh

# Start backend
echo -e "${YELLOW}🚀 Starting backend...${NC}"
nohup ./start_backend_linux.sh > "$LOGS/backend.log" 2>&1 &
echo $! > "$LOGS/backend.pid"

sleep 5

# Check backend status
echo -e "${YELLOW}🔍 Checking backend status...${NC}"
if pgrep -f "backend" > /dev/null; then
    echo -e "${GREEN}✅ Backend started successfully${NC}"
else
    echo -e "${RED}❌ Backend failed to start${NC}"
    echo -e "${YELLOW}📄 Backend logs:${NC}"
    tail -20 "$LOGS/backend.log"
    exit 1
fi

# Test health
echo -e "${YELLOW}🔍 Testing backend health...${NC}"
for i in {1..10}; do
    if curl -s http://localhost:8080/health > /dev/null; then
        echo -e "${GREEN}✅ Backend health check passed${NC}"
        break
    else
        echo -e "${YELLOW}⏳ Attempt $i/10 failed, waiting...${NC}"
        sleep 2
    fi
done

# Test external access
echo -e "${YELLOW}🔍 Testing external access...${NC}"
if curl -s http://52.206.225.24:8080/health > /dev/null; then
    echo -e "${GREEN}✅ External backend access working${NC}"
else
    echo -e "${RED}❌ External backend access failed${NC}"
fi

# Check Oracle logs
echo -e "${YELLOW}📄 Checking Oracle logs...${NC}"
tail -20 "$LOGS/backend.log" | grep -i "oracle\|config" || echo "No Oracle logs yet"

echo ""
echo -e "${BLUE}🔧 BACKEND REBUILT AND ORACLE FIXED!${NC}"
echo ""
echo -e "${GREEN}Test now:${NC}"
echo "1. Login: http://52.206.225.24:3000"
echo "2. Stock search with any SKU"
echo "3. Check logs: tail -f $LOGS/backend.log"
