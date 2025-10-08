#!/bin/bash

# Fix frontend binding to external IP
cd /d02/projects/amz-web-tools

echo "🔧 Fixing frontend binding to external IP..."

# Kill existing frontend
pkill -f "next" 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
sleep 3

# Start frontend with explicit host binding
echo "🚀 Starting frontend with external IP binding..."
cd .next/standalone

# Set environment variables for external binding
export HOSTNAME=0.0.0.0
export PORT=3000

# Start frontend
nohup node server.js > /d02/logs/frontend.log 2>&1 &
echo $! > /d02/logs/frontend.pid

sleep 10

# Test local access
echo "🔍 Testing local access..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend accessible locally"
else
    echo "❌ Frontend NOT accessible locally"
fi

# Test external access
echo "🔍 Testing external access..."
if curl -s http://52.206.225.24:3000 > /dev/null; then
    echo "✅ Frontend accessible externally"
else
    echo "❌ Frontend NOT accessible externally"
fi

# Check what's listening on port 3000
echo "🔍 Checking port 3000 binding..."
netstat -tlnp 2>/dev/null | grep ":3000" || echo "No process listening on port 3000"

# Show frontend logs
echo "📄 Recent frontend logs:"
tail -10 /d02/logs/frontend.log

echo "✅ Frontend binding fix completed!"
echo "Test: http://52.206.225.24:3000"
