#!/bin/bash
set -euo pipefail

# =============================================
# AMZ Web Tools - Apply Working Oracle Fix
# Based on commit 442b5df8 that worked!
# =============================================

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔒 APPLYING WORKING ORACLE FIX (from commit 442b5df8)...${NC}"
echo ""

cd "$BASE"

# Stop backend
echo -e "${YELLOW}🛑 Stopping backend...${NC}"
pkill -f "bin/backend" 2>/dev/null || true
pkill -f "amz-web-tools" 2>/dev/null || true
sleep 3

# Create LOCKED .env with working configuration
echo -e "${YELLOW}🔒 Creating LOCKED .env with WORKING server settings...${NC}"
cat > .env << 'EOF'
# =============================================
# LOCKED Configuration - WORKING VERSION
# Based on commit 442b5df8 that fixed Oracle!
# =============================================

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

# CORS (Allow external access from frontend)
CORS_ALLOWED_ORIGINS=http://52.206.225.24:3000,http://52.206.225.24,https://52.206.225.24:3000,https://52.206.225.24

# Logging
LOG_LEVEL=info

# ORACLE CONFIGURATION - LOCKED TO SERVER (WORKING!)
ORACLE_HOST=164.152.40.38
ORACLE_PORT=1521
ORACLE_USER=dashjc
ORACLE_PASSWORD=@Joao1225
ORACLE_SERVICE=nbs
ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7

# PostgreSQL Database Configuration (RESTORED ORIGINAL)
PG_HOST=shared-codako-nlb-3f3ad9f6c528c4a6.elb.us-east-1.amazonaws.com
PG_PORT=5433
PG_USER=codako_bi
PG_PASSWORD=lNkIXKc9CQuyv28B
PG_DATABASE=codako_bi
PG_SSL_MODE=disable

# Next.js Public API URL (Frontend to Backend)
NEXT_PUBLIC_API_URL=http://52.206.225.24:8080/api/v1

# Server Host (for binding to all interfaces)
SERVER_HOST=0.0.0.0
FRONTEND_HOST=0.0.0.0
EOF

# Make file read-only to prevent changes (THIS WAS KEY!)
chmod 444 .env

echo -e "${GREEN}✅ .env LOCKED with working server Oracle settings!${NC}"
echo -e "${YELLOW}📋 Oracle config (LOCKED):${NC}"
grep -E "^ORACLE_" .env

# Set environment variables explicitly (THIS WAS ALSO KEY!)
echo -e "${YELLOW}🔧 Setting environment variables explicitly...${NC}"
export $(grep -v '^#' .env | xargs)

echo -e "${GREEN}🔍 Environment variables set:${NC}"
echo "ORACLE_HOST: '$ORACLE_HOST'"
echo "ORACLE_USER: '$ORACLE_USER'"
echo "ORACLE_SERVICE: '$ORACLE_SERVICE'"
echo "PG_HOST: '$PG_HOST'"
echo "PG_USER: '$PG_USER'"

# Create symlink in backend directory
echo -e "${YELLOW}🔗 Creating symlink in backend directory...${NC}"
ln -sf "$BASE/.env" "$BASE/backend/.env"

# Start backend with locked config
echo -e "${YELLOW}🚀 Starting backend with LOCKED working config...${NC}"

# Kill any existing processes on port 8080
fuser -k 8080/tcp 2>/dev/null || true
sleep 2

# Start backend
cd "$BASE/backend"
nohup ./bin/backend > "$LOGS/backend.log" 2>&1 &
echo $! > "$LOGS/backend.pid"

sleep 5

echo -e "${GREEN}✅ Backend started with LOCKED working Oracle config!${NC}"
echo ""

# Check Oracle connection logs
echo -e "${YELLOW}🔍 Checking Oracle connection logs:${NC}"
tail -20 "$LOGS/backend.log" | grep -E "(Oracle|Config|Host|PostgreSQL)" || echo "No Oracle logs found yet"

echo ""
echo -e "${BLUE}🎯 WORKING ORACLE FIX APPLIED!${NC}"
echo -e "${GREEN}✅ Oracle should now connect to 164.152.40.38${NC}"
echo -e "${GREEN}✅ PostgreSQL should connect to shared-codako-nlb${NC}"
echo ""
echo -e "${YELLOW}Test now:${NC}"
echo "1. Stock search (Oracle)"
echo "2. XML Integrator (PostgreSQL + Oracle)"
echo ""
echo -e "${YELLOW}To check logs:${NC}"
echo "tail -f $LOGS/backend.log"
