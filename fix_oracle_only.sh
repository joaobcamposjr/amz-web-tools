#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Fix Oracle Only
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 FIXING ORACLE ONLY (without breaking login)${NC}"
echo ""

cd "$BASE"

# Load environment variables
source .env 2>/dev/null || true

echo -e "${YELLOW}📊 Current Oracle Configuration:${NC}"
echo "Oracle Host: ${ORACLE_HOST:-'NOT SET'}"
echo "Oracle User: ${ORACLE_USER:-'NOT SET'}"
echo "Oracle Service: ${ORACLE_SERVICE:-'NOT SET'}"
echo "Oracle Lib Dir: ${ORACLE_LIB_DIR:-'NOT SET'}"
echo ""

# Check Oracle directory and libraries
echo -e "${YELLOW}🔍 Checking Oracle client libraries...${NC}"
ORACLE_LIB_DIR="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}"

if [ -d "$ORACLE_LIB_DIR" ]; then
    echo -e "${GREEN}✅ Oracle directory exists: $ORACLE_LIB_DIR${NC}"
    
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
    find /opt -name "*oracle*" -type d 2>/dev/null | head -10 || echo "No Oracle directories found in /opt"
fi

# Set Oracle environment
echo ""
echo -e "${YELLOW}🔧 Setting Oracle environment...${NC}"
export ORACLE_HOME="$ORACLE_LIB_DIR"
export LD_LIBRARY_PATH="$ORACLE_LIB_DIR:${LD_LIBRARY_PATH:-}"

echo "ORACLE_HOME: $ORACLE_HOME"
echo "LD_LIBRARY_PATH: $LD_LIBRARY_PATH"

# Test Oracle connection with a simple Go program
echo ""
echo -e "${YELLOW}🔍 Testing Oracle connection with Go...${NC}"

# Create a simple Oracle test program
cat > /tmp/test_oracle_go.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    _ "github.com/sijms/go-ora/v2"
)

func main() {
    // Get Oracle config from environment
    host := os.Getenv("ORACLE_HOST")
    user := os.Getenv("ORACLE_USER")
    password := os.Getenv("ORACLE_PASSWORD")
    service := os.Getenv("ORACLE_SERVICE")
    port := os.Getenv("ORACLE_PORT")
    
    if port == "" {
        port = "1521"
    }
    
    fmt.Printf("Testing Oracle connection...\n")
    fmt.Printf("Host: %s\n", host)
    fmt.Printf("User: %s\n", user)
    fmt.Printf("Service: %s\n", service)
    fmt.Printf("Port: %s\n", port)
    
    // Build DSN
    dsn := fmt.Sprintf("oracle://%s:%s@%s:%s/%s", user, password, host, port, service)
    fmt.Printf("DSN: oracle://%s:****@%s:%s/%s\n", user, host, port, service)
    
    // Try to connect
    db, err := sql.Open("oracle", dsn)
    if err != nil {
        fmt.Printf("❌ Error opening Oracle connection: %v\n", err)
        return
    }
    defer db.Close()
    
    // Test connection
    err = db.Ping()
    if err != nil {
        fmt.Printf("❌ Error pinging Oracle: %v\n", err)
        return
    }
    
    fmt.Printf("✅ Oracle connection successful!\n")
    
    // Try a simple query
    var result string
    err = db.QueryRow("SELECT 'Oracle connection test' FROM DUAL").Scan(&result)
    if err != nil {
        fmt.Printf("⚠️ Error executing query: %v\n", err)
        return
    }
    
    fmt.Printf("✅ Query result: %s\n", result)
}
EOF

# Compile and run the test
cd /tmp
if go run test_oracle_go.go; then
    echo -e "${GREEN}✅ Oracle connection test passed!${NC}"
else
    echo -e "${RED}❌ Oracle connection test failed${NC}"
fi

# Clean up
rm -f /tmp/test_oracle_go.go

cd "$BASE"

# Check backend logs for Oracle errors
echo ""
echo -e "${YELLOW}📄 Checking backend logs for Oracle errors...${NC}"
if [ -f "$LOGS/backend.log" ]; then
    echo "=== Recent Oracle logs ==="
    tail -50 "$LOGS/backend.log" | grep -i "oracle" || echo "No Oracle logs found"
else
    echo -e "${RED}❌ Backend log file not found${NC}"
fi

echo ""
echo -e "${BLUE}🔧 Oracle fix completed!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Test Stock search in the frontend"
echo "2. Check backend logs: tail -f $LOGS/backend.log"
echo "3. If still not working, check Oracle client installation"
