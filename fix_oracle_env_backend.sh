#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Fix Oracle Environment in Backend
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 FIXING ORACLE ENVIRONMENT IN BACKEND${NC}"
echo ""

cd "$BASE"

# Stop backend first
echo -e "${YELLOW}🛑 Stopping backend...${NC}"
pkill -f "backend" 2>/dev/null || true
sleep 3

# Load environment variables
echo -e "${YELLOW}🔍 Loading environment variables...${NC}"
source .env 2>/dev/null || true

echo "Oracle Host: ${ORACLE_HOST:-'NOT SET'}"
echo "Oracle User: ${ORACLE_USER:-'NOT SET'}"
echo "Oracle Service: ${ORACLE_SERVICE:-'NOT SET'}"

# Set Oracle environment variables explicitly
echo -e "${YELLOW}🔧 Setting Oracle environment variables...${NC}"
export ORACLE_HOST=164.152.40.38
export ORACLE_PORT=1521
export ORACLE_USER=dashjc
export ORACLE_PASSWORD=@Joao1225
export ORACLE_SERVICE=nbs
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export ORACLE_HOME=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}

echo "✅ Oracle environment variables set:"
echo "ORACLE_HOST: $ORACLE_HOST"
echo "ORACLE_USER: $ORACLE_USER"
echo "ORACLE_SERVICE: $ORACLE_SERVICE"
echo "ORACLE_LIB_DIR: $ORACLE_LIB_DIR"

# Create a startup script that preserves environment
echo -e "${YELLOW}🔧 Creating backend startup script with Oracle env...${NC}"
cat > start_backend_with_oracle.sh << 'EOF'
#!/bin/bash
cd /d02/projects/amz-web-tools

# Load environment variables
source .env 2>/dev/null || true

# Set Oracle environment explicitly
export ORACLE_HOST=164.152.40.38
export ORACLE_PORT=1521
export ORACLE_USER=dashjc
export ORACLE_PASSWORD=@Joao1225
export ORACLE_SERVICE=nbs
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export ORACLE_HOME=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}

# Set other required variables
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

echo "🔧 Starting backend with Oracle environment:"
echo "ORACLE_HOST: $ORACLE_HOST"
echo "ORACLE_USER: $ORACLE_USER"
echo "ORACLE_SERVICE: $ORACLE_SERVICE"

# Start backend
cd backend
exec ./bin/backend
EOF

chmod +x start_backend_with_oracle.sh

# Start backend with Oracle environment
echo -e "${YELLOW}🚀 Starting backend with Oracle environment...${NC}"
nohup ./start_backend_with_oracle.sh > "$LOGS/backend.log" 2>&1 &
echo $! > "$LOGS/backend.pid"

sleep 5

# Check if backend started
echo -e "${YELLOW}🔍 Checking backend status...${NC}"
if pgrep -f "backend" > /dev/null; then
    echo -e "${GREEN}✅ Backend started successfully${NC}"
else
    echo -e "${RED}❌ Backend failed to start${NC}"
    echo -e "${YELLOW}📄 Backend logs:${NC}"
    tail -20 "$LOGS/backend.log"
    exit 1
fi

# Test health endpoint
echo -e "${YELLOW}🔍 Testing backend health...${NC}"
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✅ Backend health check passed${NC}"
else
    echo -e "${RED}❌ Backend health check failed${NC}"
fi

# Check Oracle logs
echo -e "${YELLOW}📄 Checking Oracle logs...${NC}"
tail -20 "$LOGS/backend.log" | grep -i "oracle\|config" || echo "No Oracle logs found yet"

echo ""
echo -e "${BLUE}🔧 Oracle environment fix completed!${NC}"
echo ""
echo -e "${YELLOW}Test now:${NC}"
echo "1. Login should still work"
echo "2. Stock search should work with Oracle"
echo "3. Check logs: tail -f $LOGS/backend.log"
