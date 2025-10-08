#!/bin/bash

# Fix go.sum and Oracle
cd /d02/projects/amz-web-tools

echo "🔧 Fixing go.sum and Oracle..."

# Stop backend
pkill -f "backend" 2>/dev/null || true
sleep 2

# Fix go.sum
cd backend
echo "🔧 Generating go.sum..."
go mod tidy
go mod download

# Build
echo "🔨 Building backend..."
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}

go build -o bin/backend .

if [ -f "bin/backend" ]; then
    echo "✅ Backend built successfully"
else
    echo "❌ Build failed"
    exit 1
fi

# Start with Oracle config
echo "🚀 Starting backend with Oracle config..."
export ORACLE_HOST=164.152.40.38
export ORACLE_USER=dashjc
export ORACLE_SERVICE=nbs
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

nohup ./bin/backend > /d02/logs/backend.log 2>&1 &
echo $! > /d02/logs/backend.pid

sleep 5

# Test
echo "🔍 Testing backend..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ Backend health check passed"
else
    echo "❌ Backend health check failed"
    echo "📄 Backend logs:"
    tail -10 /d02/logs/backend.log
fi

echo "✅ Done! Test: http://52.206.225.24:3000"
