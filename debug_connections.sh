#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Connection Debug Script
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Testing Database Connections...${NC}"
echo ""

# Load environment variables
source "$BASE/.env" 2>/dev/null || echo "⚠️ Could not load .env file"

echo -e "${YELLOW}📊 Environment Variables:${NC}"
echo "Oracle Host: ${ORACLE_HOST:-'NOT SET'}"
echo "Oracle User: ${ORACLE_USER:-'NOT SET'}"
echo "Oracle Service: ${ORACLE_SERVICE:-'NOT SET'}"
echo "Oracle Lib Dir: ${ORACLE_LIB_DIR:-'NOT SET'}"
echo ""
echo "PostgreSQL Host: ${PG_HOST:-'NOT SET'}"
echo "PostgreSQL Port: ${PG_PORT:-'NOT SET'}"
echo "PostgreSQL User: ${PG_USER:-'NOT SET'}"
echo "PostgreSQL Database: ${PG_DATABASE:-'NOT SET'}"
echo ""

# Test Oracle connection
echo -e "${YELLOW}🔍 Testing Oracle Connection...${NC}"
if [ -n "${ORACLE_HOST:-}" ] && [ -n "${ORACLE_USER:-}" ]; then
    # Set Oracle environment
    export ORACLE_HOME="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}"
    export LD_LIBRARY_PATH="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}:${LD_LIBRARY_PATH:-}"
    
    echo "Oracle DSN: oracle://${ORACLE_USER}:****@${ORACLE_HOST}:${ORACLE_PORT:-1521}/${ORACLE_SERVICE}"
    
    # Test if Oracle client is available
    if [ -f "${ORACLE_HOME}/libclntsh.so" ]; then
        echo -e "${GREEN}✅ Oracle client library found${NC}"
    else
        echo -e "${RED}❌ Oracle client library not found at ${ORACLE_HOME}/libclntsh.so${NC}"
    fi
    
    if [ -f "${ORACLE_HOME}/libociicus.so" ]; then
        echo -e "${GREEN}✅ Oracle instant client library found${NC}"
    else
        echo -e "${RED}❌ Oracle instant client library not found at ${ORACLE_HOME}/libociicus.so${NC}"
    fi
else
    echo -e "${RED}❌ Oracle configuration missing${NC}"
fi
echo ""

# Test PostgreSQL connection
echo -e "${YELLOW}🔍 Testing PostgreSQL Connection...${NC}"
if [ -n "${PG_HOST:-}" ] && [ -n "${PG_USER:-}" ]; then
    echo "PostgreSQL DSN: host=${PG_HOST} port=${PG_PORT:-5432} user=${PG_USER} password=**** dbname=${PG_DATABASE}"
    
    # Test network connectivity
    if command -v nc >/dev/null 2>&1; then
        if nc -z "${PG_HOST}" "${PG_PORT:-5432}" 2>/dev/null; then
            echo -e "${GREEN}✅ PostgreSQL port ${PG_PORT:-5432} is reachable on ${PG_HOST}${NC}"
        else
            echo -e "${RED}❌ PostgreSQL port ${PG_PORT:-5432} is NOT reachable on ${PG_HOST}${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️ netcat (nc) not available, cannot test port connectivity${NC}"
    fi
else
    echo -e "${RED}❌ PostgreSQL configuration missing${NC}"
fi
echo ""

# Check backend logs for connection errors
echo -e "${YELLOW}📄 Recent Backend Logs (connection related):${NC}"
if [ -f "$LOGS/backend.log" ]; then
    echo "=== Oracle Connection Logs ==="
    grep -i "oracle\|connection" "$LOGS/backend.log" | tail -10 || echo "No Oracle logs found"
    echo ""
    echo "=== PostgreSQL Connection Logs ==="
    grep -i "postgres\|pg_" "$LOGS/backend.log" | tail -10 || echo "No PostgreSQL logs found"
    echo ""
    echo "=== XML Integrator Logs ==="
    grep -i "xml.*integrator\|xmlIntegrator" "$LOGS/backend.log" | tail -10 || echo "No XML Integrator logs found"
else
    echo -e "${RED}❌ Backend log file not found at $LOGS/backend.log${NC}"
fi

echo ""
echo -e "${BLUE}🔍 Connection test completed!${NC}"
