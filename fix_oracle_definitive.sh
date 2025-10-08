#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Definitive Oracle Fix
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${RED}🚨 DEFINITIVE ORACLE FIX${NC}"
echo ""

cd "$BASE"

# Stop ALL backend processes
echo -e "${YELLOW}🛑 Stopping ALL backend processes...${NC}"
pkill -f "backend" 2>/dev/null || true
pkill -f "amz-web-tools" 2>/dev/null || true
fuser -k 8080/tcp 2>/dev/null || true
sleep 3

# Verify .env file exists and is readable
echo -e "${YELLOW}🔍 Checking .env file...${NC}"
if [ ! -f ".env" ]; then
    echo -e "${RED}❌ .env file not found, creating it...${NC}"
    cat > .env << 'EOF'
# Database Configuration
DB_HOST=54.204.42.134
DB_PORT=1433
DB_USER=sa
DB_PASSWORD=321@Mudar@7089341@
DB_NAME=integration
DB_SSL_MODE=disable

# Server Configuration
SERVER_PORT=8080
FRONTEND_PORT=3000
JWT_SECRET=amz-web-tools-secret-key-2024
JWT_EXPIRE_HOURS=168

# API Configuration (Car Plate)
PLATE_API_URL=https://wdapi2.com.br/consulta/PLACA/your-api-key
PLATE_API_KEY=your-api-key

# Environment
ENVIRONMENT=production

# CORS
CORS_ALLOWED_ORIGINS=http://52.206.225.24:3000,http://52.206.225.24

# Logging
LOG_LEVEL=info

# ORACLE CONFIGURATION - DEFINITIVE
ORACLE_HOST=164.152.40.38
ORACLE_PORT=1521
ORACLE_USER=dashjc
ORACLE_PASSWORD=@Joao1225
ORACLE_SERVICE=nbs
ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7

# PostgreSQL Database Configuration
PG_HOST=shared-codako-nlb-3f3ad9f6c528c4a6.elb.us-east-1.amazonaws.com
PG_PORT=5433
PG_USER=codako_bi
PG_PASSWORD=lNkIXKc9CQuyv28B
PG_DATABASE=codako_bi
PG_SSL_MODE=disable

# Next.js Public API URL
NEXT_PUBLIC_API_URL=http://52.206.225.24:8080/api/v1

# Server Host
SERVER_HOST=0.0.0.0
FRONTEND_HOST=0.0.0.0
EOF
fi

# Make sure .env is readable
chmod 644 .env

# Verify Oracle config in .env
echo -e "${YELLOW}🔍 Verifying Oracle config in .env...${NC}"
if grep -q "ORACLE_HOST=164.152.40.38" .env; then
    echo -e "${GREEN}✅ Oracle config found in .env${NC}"
    grep -E "^ORACLE_" .env
else
    echo -e "${RED}❌ Oracle config missing from .env${NC}"
fi

# Create backend startup script with hardcoded Oracle config
echo -e "${YELLOW}🔧 Creating backend startup script...${NC}"
cat > start_backend_oracle.sh << 'EOF'
#!/bin/bash

# Hardcoded Oracle configuration
export ORACLE_HOST=164.152.40.38
export ORACLE_PORT=1521
export ORACLE_USER=dashjc
export ORACLE_PASSWORD=@Joao1225
export ORACLE_SERVICE=nbs
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export ORACLE_HOME=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}

# Other required variables
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

# Database config
export DB_HOST=54.204.42.134
export DB_PORT=1433
export DB_USER=sa
export DB_PASSWORD=321@Mudar@7089341@
export DB_NAME=integration

# PostgreSQL config
export PG_HOST=shared-codako-nlb-3f3ad9f6c528c4a6.elb.us-east-1.amazonaws.com
export PG_PORT=5433
export PG_USER=codako_bi
export PG_PASSWORD=lNkIXKc9CQuyv28B
export PG_DATABASE=codako_bi

# JWT config
export JWT_SECRET=amz-web-tools-secret-key-2024
export JWT_EXPIRE_HOURS=168

echo "🔧 Starting backend with DEFINITIVE Oracle config:"
echo "ORACLE_HOST: $ORACLE_HOST"
echo "ORACLE_USER: $ORACLE_USER"
echo "ORACLE_SERVICE: $ORACLE_SERVICE"

cd /d02/projects/amz-web-tools/backend
exec ./bin/backend
EOF

chmod +x start_backend_oracle.sh

# Start backend with definitive Oracle config
echo -e "${YELLOW}🚀 Starting backend with DEFINITIVE Oracle config...${NC}"
nohup ./start_backend_oracle.sh > "$LOGS/backend.log" 2>&1 &
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

# Test backend health
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

# Test Oracle configuration in backend
echo -e "${YELLOW}🔍 Testing Oracle configuration in backend...${NC}"
sleep 2
tail -20 "$LOGS/backend.log" | grep -i "oracle\|config" || echo "No Oracle logs yet"

echo ""
echo -e "${BLUE}🔧 DEFINITIVE ORACLE FIX COMPLETED!${NC}"
echo ""
echo -e "${GREEN}Oracle should now work! Test:${NC}"
echo "1. Login: http://52.206.225.24:3000"
echo "2. Stock search with any SKU"
echo "3. Check logs: tail -f $LOGS/backend.log"
