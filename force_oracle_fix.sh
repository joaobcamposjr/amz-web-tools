#!/bin/bash

echo "🔧 FORCING Oracle configuration fix..."

cd /d02/projects/amz-web-tools

# Stop backend
echo "🛑 Stopping backend..."
pkill -f "bin/backend" 2>/dev/null || true
sleep 3

# Force correct Oracle config
echo "⚙️ Forcing correct Oracle configuration..."
cat > .env << 'EOF'
# =============================================
# Environment Variables for Server Deployment
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

# Oracle Database Configuration - SERVER SETTINGS
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

echo "✅ .env file recreated with correct Oracle settings!"

echo ""
echo "📋 Oracle configuration:"
grep -E "^ORACLE_" .env

echo ""
echo "🚀 Starting backend with correct config..."
cd /d02/projects/amz-web-tools
nohup ./bin/backend > /d02/logs/backend.log 2>&1 &
echo $! > /d02/pids/backend.pid

sleep 5

echo "✅ Backend started!"
echo ""
echo "🔍 Testing Oracle connection..."
sleep 2
tail -10 /d02/logs/backend.log | grep -i oracle || echo "No Oracle logs yet"

echo ""
echo "🎯 Now test Stock search to see if Oracle connects!"

