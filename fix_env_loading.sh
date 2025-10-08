#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Fix Environment Loading
# =============================================

BASE="/d02/projects/amz-web-tools"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Fixing Environment Loading Issues...${NC}"
echo ""

# Check if .env exists in project root
if [ -f "$BASE/.env" ]; then
    echo -e "${GREEN}✅ .env file found in project root${NC}"
else
    echo -e "${RED}❌ .env file NOT found in project root${NC}"
    echo "Creating .env from env.server..."
    cp "$BASE/env.server" "$BASE/.env"
    echo -e "${GREEN}✅ .env file created${NC}"
fi

# Check if .env exists in backend directory
if [ -f "$BASE/backend/.env" ]; then
    echo -e "${GREEN}✅ .env file found in backend directory${NC}"
else
    echo -e "${YELLOW}⚠️ .env file NOT found in backend directory${NC}"
    echo "Creating symlink to project root .env..."
    ln -sf "$BASE/.env" "$BASE/backend/.env"
    echo -e "${GREEN}✅ Symlink created${NC}"
fi

# Verify environment variables
echo ""
echo -e "${YELLOW}🔍 Verifying Environment Variables...${NC}"
source "$BASE/.env"

echo "Oracle Host: ${ORACLE_HOST:-'NOT SET'}"
echo "Oracle User: ${ORACLE_USER:-'NOT SET'}"
echo "Oracle Service: ${ORACLE_SERVICE:-'NOT SET'}"
echo ""
echo "PostgreSQL Host: ${PG_HOST:-'NOT SET'}"
echo "PostgreSQL User: ${PG_USER:-'NOT SET'}"
echo "PostgreSQL Database: ${PG_DATABASE:-'NOT SET'}"

# Fix Oracle client library issue
echo ""
echo -e "${YELLOW}🔧 Fixing Oracle Client Library...${NC}"
ORACLE_LIB_DIR="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}"

if [ -d "$ORACLE_LIB_DIR" ]; then
    echo -e "${GREEN}✅ Oracle lib directory exists: $ORACLE_LIB_DIR${NC}"
    
    # Check for libclntsh.so
    if [ -f "$ORACLE_LIB_DIR/libclntsh.so" ]; then
        echo -e "${GREEN}✅ libclntsh.so found${NC}"
    else
        echo -e "${RED}❌ libclntsh.so NOT found${NC}"
        echo "Looking for alternative libclntsh files..."
        ls -la "$ORACLE_LIB_DIR"/libclntsh* 2>/dev/null || echo "No libclntsh files found"
    fi
    
    # Check for libociicus.so
    if [ -f "$ORACLE_LIB_DIR/libociicus.so" ]; then
        echo -e "${GREEN}✅ libociicus.so found${NC}"
    else
        echo -e "${YELLOW}⚠️ libociicus.so NOT found${NC}"
        echo "Looking for alternative oci files..."
        ls -la "$ORACLE_LIB_DIR"/liboci* 2>/dev/null || echo "No oci files found"
        
        # Try to find the correct oci library
        echo "Checking for other oci libraries..."
        find "$ORACLE_LIB_DIR" -name "*oci*" -type f 2>/dev/null | head -5 || echo "No oci libraries found"
    fi
else
    echo -e "${RED}❌ Oracle lib directory NOT found: $ORACLE_LIB_DIR${NC}"
fi

# Create a test environment loading script
echo ""
echo -e "${YELLOW}🔧 Creating environment test script...${NC}"
cat > "$BASE/test_env_loading.sh" << 'EOF'
#!/bin/bash
echo "Testing environment loading..."

# Test loading from project root
echo "Loading from project root..."
source /d02/projects/amz-web-tools/.env 2>/dev/null || echo "Failed to load .env from project root"

echo "Oracle Host: ${ORACLE_HOST:-'NOT SET'}"
echo "Oracle User: ${ORACLE_USER:-'NOT SET'}"
echo "PostgreSQL Host: ${PG_HOST:-'NOT SET'}"
echo "PostgreSQL User: ${PG_USER:-'NOT SET'}"

# Test loading from backend directory
echo ""
echo "Loading from backend directory..."
source /d02/projects/amz-web-tools/backend/.env 2>/dev/null || echo "Failed to load .env from backend directory"

echo "Oracle Host: ${ORACLE_HOST:-'NOT SET'}"
echo "Oracle User: ${ORACLE_USER:-'NOT SET'}"
echo "PostgreSQL Host: ${PG_HOST:-'NOT SET'}"
echo "PostgreSQL User: ${PG_USER:-'NOT SET'}"
EOF

chmod +x "$BASE/test_env_loading.sh"
echo -e "${GREEN}✅ Environment test script created${NC}"

echo ""
echo -e "${BLUE}🔧 Environment loading fix completed!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Run: ./test_env_loading.sh"
echo "2. Restart backend: ./stop_server.sh && ./init_server.sh"
echo "3. Check logs: tail -f /d02/logs/backend.log"
