#!/bin/bash

# Check and fix frontend
cd /d02/projects/amz-web-tools

echo "🔍 Checking frontend status..."

# Check if frontend is running
if pgrep -f "next" > /dev/null; then
    echo "✅ Frontend process found"
    pgrep -f "next"
else
    echo "❌ Frontend process NOT found"
fi

# Check port 3000
if netstat -tlnp 2>/dev/null | grep ":3000" > /dev/null; then
    echo "✅ Port 3000 is in use"
    netstat -tlnp 2>/dev/null | grep ":3000"
else
    echo "❌ Port 3000 is NOT in use"
fi

# Test frontend access
echo "🔍 Testing frontend access..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend accessible locally"
else
    echo "❌ Frontend NOT accessible locally"
fi

if curl -s http://52.206.225.24:3000 > /dev/null; then
    echo "✅ Frontend accessible externally"
else
    echo "❌ Frontend NOT accessible externally"
fi

# Check frontend logs
echo "📄 Frontend logs:"
if [ -f "/d02/logs/frontend.log" ]; then
    tail -20 /d02/logs/frontend.log
else
    echo "❌ Frontend log file not found"
fi

# Kill and restart frontend
echo "🔄 Restarting frontend..."
pkill -f "next" 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
sleep 3

# Start frontend
echo "🚀 Starting frontend..."
cd .next/standalone
nohup node server.js > /d02/logs/frontend.log 2>&1 &
echo $! > /d02/logs/frontend.pid

sleep 10

# Test again
echo "🔍 Testing frontend after restart..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend restarted successfully"
else
    echo "❌ Frontend still not working"
    echo "📄 Recent logs:"
    tail -10 /d02/logs/frontend.log
fi

echo "✅ Frontend check completed!"
