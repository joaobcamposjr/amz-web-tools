#!/bin/bash

echo "🔒 LOCKING Oracle configuration to SERVER settings..."

cd /d02/projects/amz-web-tools

# Stop backend
echo "🛑 Stopping backend..."
pkill -f "bin/backend" 2>/dev/null || true
sleep 3

# Make .env READ-ONLY and force correct config
echo "🔒 Creating LOCKED .env with server settings..."
cat > .env << 'EOF'
# =============================================
# LOCKED Oracle Configuration for SERVER
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

# ORACLE CONFIGURATION - LOCKED TO SERVER
ORACLE_HOST=164.152.40.38
ORACLE_PORT=1521
ORACLE_USER=dashjc
ORACLE_PASSWORD=@Joao1225
ORACLE_SERVICE=nbs
ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7

# PostgreSQL Database Configuration (for integrator)
PG_HOST=your-postgres-host
PG_PORT=5433
PG_USER=your-postgres-user
PG_PASSWORD=your-postgres-password
PG_DATABASE=your-postgres-database
PG_SSL_MODE=disable

# Next.js Public API URL (Frontend to Backend)
NEXT_PUBLIC_API_URL=http://52.206.225.24:8080/api/v1

# Server Host (for binding to all interfaces)
SERVER_HOST=0.0.0.0
FRONTEND_HOST=0.0.0.0
EOF

# Make file read-only to prevent changes
chmod 444 .env

echo "✅ .env LOCKED with server Oracle settings!"
echo "📋 Oracle config (LOCKED):"
grep -E "^ORACLE_" .env

# Set environment variables explicitly
echo "🔧 Setting environment variables..."
export $(grep -v '^#' .env | xargs)

echo "🔍 Environment variables set:"
echo "ORACLE_HOST: '$ORACLE_HOST'"
echo "ORACLE_USER: '$ORACLE_USER'"
echo "ORACLE_SERVICE: '$ORACLE_SERVICE'"

# Start backend
echo "🚀 Starting backend with LOCKED server config..."
nohup ./bin/backend > /d02/logs/backend.log 2>&1 &
echo $! > /d02/pids/backend.pid

sleep 5

echo "✅ Backend started with LOCKED Oracle config!"
echo ""
echo "🔍 Checking Oracle connection logs:"
tail -15 /d02/logs/backend.log | grep -E "(Oracle|Config|Host)"

echo ""
echo "🎯 Now test Stock search - Oracle should connect to 164.152.40.38!"

