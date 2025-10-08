#!/bin/bash

echo "🔧 Rebuilding backend with correct .env loading..."

cd /d02/projects/amz-web-tools

# Stop backend
echo "🛑 Stopping backend..."
pkill -f "bin/backend" 2>/dev/null || true
sleep 3

# Verify .env exists
echo "🔍 Verifying .env file..."
if [ -f ".env" ]; then
    echo "✅ .env file exists"
    echo "📋 Oracle config in .env:"
    grep -E "^ORACLE_" .env
else
    echo "❌ .env file not found!"
    exit 1
fi

# Rebuild backend
echo "🔨 Rebuilding backend..."
cd backend
go build -o ../bin/backend .
cd ..

echo "✅ Backend rebuilt!"

# Start backend with explicit .env loading
echo "🚀 Starting backend with explicit environment..."
cd /d02/projects/amz-web-tools

# Set environment variables explicitly
export $(grep -v '^#' .env | xargs)

echo "🔍 Testing environment variables:"
echo "ORACLE_HOST: '$ORACLE_HOST'"
echo "ORACLE_USER: '$ORACLE_USER'"
echo "ORACLE_SERVICE: '$ORACLE_SERVICE'"

# Start backend
nohup ./bin/backend > /d02/logs/backend.log 2>&1 &
echo $! > /d02/pids/backend.pid

sleep 5

echo "✅ Backend started with explicit environment variables!"
echo ""
echo "🔍 Checking startup logs:"
tail -15 /d02/logs/backend.log | grep -E "(Loaded .env|Config loaded|Oracle)"

echo ""
echo "🎯 Now test Stock search to see if Oracle connects!"
