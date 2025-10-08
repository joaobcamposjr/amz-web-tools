#!/bin/bash

# Fix git pull conflict
cd /d02/projects/amz-web-tools

echo "🔧 Fixing git pull conflict..."

# Remove conflicting files
rm -f go.sum
rm -f go.mod

# Pull latest
git pull

# Fix go.mod if needed
cd backend
if [ -f "../go.mod" ]; then
    echo "✅ go.mod pulled successfully"
else
    echo "🔧 Creating go.mod..."
    cat > ../go.mod << 'EOF'
module amz-web-tools

go 1.21

require (
	github.com/gin-contrib/cors v1.5.0
	github.com/gin-gonic/gin v1.7.7
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.5.0
	github.com/gorilla/websocket v1.5.3
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/microsoft/go-mssqldb v1.6.0
	github.com/sijms/go-ora/v2 v2.9.0
	golang.org/x/crypto v0.17.0
)

exclude (
	github.com/bytedance/sonic v1.10.1
	github.com/chenzhuoyu/iasm v0.9.0
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d
)
EOF
fi

cd ..

# Build and start backend
echo "🔨 Building backend..."
cd backend
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}

# Stop existing backend
pkill -f "backend" 2>/dev/null || true
sleep 2

# Build
go build -o bin/backend .

# Start with Oracle config
export ORACLE_HOST=164.152.40.38
export ORACLE_USER=dashjc
export ORACLE_SERVICE=nbs
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

nohup ./bin/backend > /d02/logs/backend.log 2>&1 &
echo $! > /d02/logs/backend.pid

sleep 5
echo "✅ Backend started with Oracle config"
echo "Test: http://52.206.225.24:3000"
