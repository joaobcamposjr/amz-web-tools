#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Check Oracle Status
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 CHECKING ORACLE STATUS${NC}"
echo ""

cd "$BASE"

# Check if backend is running
echo -e "${YELLOW}🔍 Checking backend process...${NC}"
if pgrep -f "backend" > /dev/null; then
    echo -e "${GREEN}✅ Backend is running${NC}"
    pgrep -f "backend"
else
    echo -e "${RED}❌ Backend is NOT running${NC}"
fi

# Check backend logs for Oracle
echo ""
echo -e "${YELLOW}📄 Recent backend logs (Oracle related):${NC}"
if [ -f "$LOGS/backend.log" ]; then
    echo "=== Last 30 lines ==="
    tail -30 "$LOGS/backend.log"
    echo ""
    echo "=== Oracle specific logs ==="
    tail -100 "$LOGS/backend.log" | grep -i "oracle\|config\|host\|user" || echo "No Oracle logs found"
else
    echo -e "${RED}❌ Backend log file not found${NC}"
fi

# Test backend health
echo ""
echo -e "${YELLOW}🔍 Testing backend health...${NC}"
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✅ Backend health check passed${NC}"
else
    echo -e "${RED}❌ Backend health check failed${NC}"
fi

# Test Oracle API endpoint
echo ""
echo -e "${YELLOW}🔍 Testing Oracle API endpoint...${NC}"
echo "Testing stock search endpoint..."

# Create test request
TEST_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/test/stock/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $(echo 'test-token')" \
  -d '{"sku":"14329101"}' 2>/dev/null || echo "API_ERROR")

if [ "$TEST_RESPONSE" = "API_ERROR" ]; then
    echo -e "${RED}❌ Stock API endpoint failed${NC}"
else
    echo -e "${GREEN}✅ Stock API endpoint responded${NC}"
    echo "Response: $TEST_RESPONSE"
fi

# Check environment variables in running process
echo ""
echo -e "${YELLOW}🔍 Checking environment variables in backend process...${NC}"
BACKEND_PID=$(pgrep -f "backend" | head -1)
if [ -n "$BACKEND_PID" ]; then
    echo "Backend PID: $BACKEND_PID"
    echo "Oracle environment variables in process:"
    cat /proc/$BACKEND_PID/environ 2>/dev/null | tr '\0' '\n' | grep -i oracle || echo "No Oracle environment variables found in process"
else
    echo -e "${RED}❌ No backend process found${NC}"
fi

# Check current .env file
echo ""
echo -e "${YELLOW}🔍 Checking .env file...${NC}"
if [ -f ".env" ]; then
    echo "Oracle variables in .env:"
    grep -E "^ORACLE_" .env || echo "No Oracle variables in .env"
else
    echo -e "${RED}❌ .env file not found${NC}"
fi

# Check if Oracle libraries are accessible
echo ""
echo -e "${YELLOW}🔍 Checking Oracle libraries...${NC}"
ORACLE_LIB_DIR="/opt/oracle/instantclient_21_7"
if [ -d "$ORACLE_LIB_DIR" ]; then
    echo "Oracle lib directory: $ORACLE_LIB_DIR"
    echo "Libraries:"
    ls -la "$ORACLE_LIB_DIR"/libclntsh* 2>/dev/null || echo "No libclntsh files"
    ls -la "$ORACLE_LIB_DIR"/liboci* 2>/dev/null || echo "No liboci files"
else
    echo -e "${RED}❌ Oracle lib directory not found${NC}"
fi

echo ""
echo -e "${BLUE}🔍 Oracle status check completed!${NC}"
echo ""
echo -e "${YELLOW}What to do next:${NC}"
echo "1. If backend is not running: ./fix_oracle_env_backend.sh"
echo "2. If Oracle config is empty: check .env file"
echo "3. If libraries are missing: install Oracle client"
echo "4. If API fails: check backend logs"
