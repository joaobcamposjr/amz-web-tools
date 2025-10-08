#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Fix Oracle Client
# =============================================

BASE="/d02/projects/amz-web-tools"
ORACLE_LIB_DIR="/opt/oracle/instantclient_21_7"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Fixing Oracle Client Issues...${NC}"
echo ""

# Load environment variables
source "$BASE/.env"

echo -e "${YELLOW}📊 Current Oracle Configuration:${NC}"
echo "Oracle Lib Dir: ${ORACLE_LIB_DIR}"
echo "Oracle Host: ${ORACLE_HOST}"
echo "Oracle User: ${ORACLE_USER}"
echo "Oracle Service: ${ORACLE_SERVICE}"
echo ""

# Check Oracle directory
if [ -d "$ORACLE_LIB_DIR" ]; then
    echo -e "${GREEN}✅ Oracle directory exists: $ORACLE_LIB_DIR${NC}"
    
    echo -e "${YELLOW}📁 Oracle directory contents:${NC}"
    ls -la "$ORACLE_LIB_DIR" | head -20
    
    echo ""
    echo -e "${YELLOW}🔍 Looking for Oracle libraries:${NC}"
    
    # Check for libclntsh
    if [ -f "$ORACLE_LIB_DIR/libclntsh.so" ]; then
        echo -e "${GREEN}✅ libclntsh.so found${NC}"
    elif [ -f "$ORACLE_LIB_DIR/libclntsh.so.21.1" ]; then
        echo -e "${YELLOW}⚠️ libclntsh.so.21.1 found, creating symlink...${NC}"
        ln -sf "$ORACLE_LIB_DIR/libclntsh.so.21.1" "$ORACLE_LIB_DIR/libclntsh.so"
        echo -e "${GREEN}✅ Symlink created${NC}"
    elif [ -f "$ORACLE_LIB_DIR/libclntsh.so.21" ]; then
        echo -e "${YELLOW}⚠️ libclntsh.so.21 found, creating symlink...${NC}"
        ln -sf "$ORACLE_LIB_DIR/libclntsh.so.21" "$ORACLE_LIB_DIR/libclntsh.so"
        echo -e "${GREEN}✅ Symlink created${NC}"
    else
        echo -e "${RED}❌ libclntsh.so not found${NC}"
        echo "Available libclntsh files:"
        ls -la "$ORACLE_LIB_DIR"/libclntsh* 2>/dev/null || echo "No libclntsh files found"
    fi
    
    # Check for oci library
    if [ -f "$ORACLE_LIB_DIR/libociicus.so" ]; then
        echo -e "${GREEN}✅ libociicus.so found${NC}"
    else
        echo -e "${YELLOW}⚠️ libociicus.so not found, looking for alternatives...${NC}"
        
        # Look for other oci libraries
        OCI_FILES=$(find "$ORACLE_LIB_DIR" -name "*oci*" -type f 2>/dev/null | head -5)
        if [ -n "$OCI_FILES" ]; then
            echo "Found OCI files:"
            echo "$OCI_FILES"
            
            # Try to create symlink to the first oci file found
            FIRST_OCI=$(echo "$OCI_FILES" | head -1)
            if [ -n "$FIRST_OCI" ]; then
                echo -e "${YELLOW}Creating symlink from $FIRST_OCI to libociicus.so...${NC}"
                ln -sf "$FIRST_OCI" "$ORACLE_LIB_DIR/libociicus.so"
                echo -e "${GREEN}✅ Symlink created${NC}"
            fi
        else
            echo -e "${RED}❌ No OCI files found${NC}"
        fi
    fi
    
else
    echo -e "${RED}❌ Oracle directory not found: $ORACLE_LIB_DIR${NC}"
    echo "Checking for alternative Oracle directories..."
    
    # Look for Oracle directories
    find /opt -name "*oracle*" -type d 2>/dev/null | head -10 || echo "No Oracle directories found in /opt"
    find /usr -name "*oracle*" -type d 2>/dev/null | head -10 || echo "No Oracle directories found in /usr"
fi

# Test Oracle environment
echo ""
echo -e "${YELLOW}🔧 Testing Oracle Environment...${NC}"

export ORACLE_HOME="$ORACLE_LIB_DIR"
export LD_LIBRARY_PATH="$ORACLE_LIB_DIR:${LD_LIBRARY_PATH:-}"

echo "ORACLE_HOME: $ORACLE_HOME"
echo "LD_LIBRARY_PATH: $LD_LIBRARY_PATH"

# Test if libraries can be found
if command -v ldd >/dev/null 2>&1; then
    echo ""
    echo -e "${YELLOW}🔍 Testing library dependencies:${NC}"
    
    # Create a simple test program to check Oracle libraries
    cat > /tmp/test_oracle.c << 'EOF'
#include <stdio.h>
int main() {
    printf("Oracle client test\n");
    return 0;
}
EOF
    
    # Try to compile with Oracle libraries
    if gcc -o /tmp/test_oracle /tmp/test_oracle.c -L"$ORACLE_LIB_DIR" -lclntsh 2>/dev/null; then
        echo -e "${GREEN}✅ Oracle libraries can be linked successfully${NC}"
        rm -f /tmp/test_oracle /tmp/test_oracle.c
    else
        echo -e "${RED}❌ Oracle libraries cannot be linked${NC}"
        echo "Compilation error (this is expected if Oracle client is not properly installed)"
    fi
else
    echo -e "${YELLOW}⚠️ ldd not available, cannot test library dependencies${NC}"
fi

# Create a simple Oracle connection test
echo ""
echo -e "${YELLOW}🔧 Creating Oracle connection test...${NC}"
cat > "$BASE/test_oracle_connection.sh" << EOF
#!/bin/bash
source "$BASE/.env"

export ORACLE_HOME="$ORACLE_LIB_DIR"
export LD_LIBRARY_PATH="$ORACLE_LIB_DIR:\${LD_LIBRARY_PATH:-}"

echo "Testing Oracle connection..."
echo "Oracle DSN: oracle://${ORACLE_USER}:****@${ORACLE_HOST}:${ORACLE_PORT:-1521}/${ORACLE_SERVICE}"

# Test with sqlplus if available
if command -v sqlplus >/dev/null 2>&1; then
    echo "sqlplus found, testing connection..."
    echo "exit;" | sqlplus "${ORACLE_USER}/${ORACLE_PASSWORD}@${ORACLE_HOST}:${ORACLE_PORT:-1521}/${ORACLE_SERVICE}" 2>/dev/null && echo "✅ Oracle connection successful" || echo "❌ Oracle connection failed"
else
    echo "sqlplus not available, cannot test direct connection"
fi
EOF

chmod +x "$BASE/test_oracle_connection.sh"
echo -e "${GREEN}✅ Oracle connection test script created${NC}"

echo ""
echo -e "${BLUE}🔧 Oracle client fix completed!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Run: ./test_oracle_connection.sh"
echo "2. Restart backend: ./stop_server.sh && ./init_server.sh"
echo "3. Check backend logs: tail -f /d02/logs/backend.log"
