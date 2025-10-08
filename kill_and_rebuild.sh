#!/bin/bash

echo "🔄 Killing all processes and rebuilding frontend..."

cd /d02/projects/amz-web-tools

# Kill ALL processes on port 3000
echo "🔫 Killing all processes on port 3000..."
fuser -k 3000/tcp 2>/dev/null || true
lsof -ti:3000 | xargs kill -9 2>/dev/null || true
pkill -f "node.*3000" 2>/dev/null || true
pkill -f "next.*3000" 2>/dev/null || true
pkill -f "standalone.*server.js" 2>/dev/null || true

# Wait a bit
sleep 3

# Double check port is free
echo "🔍 Checking port 3000 is free..."
if lsof -i :3000 > /dev/null; then
    echo "❌ Port 3000 still in use, killing more aggressively..."
    netstat -tlnp 2>/dev/null | grep :3000 | awk '{print $7}' | cut -d'/' -f1 | xargs kill -9 2>/dev/null || true
    sleep 2
fi

# Set environment
export NEXT_PUBLIC_API_URL="http://52.206.225.24:8080/api/v1"
echo "NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL" > .env.local

# Update .env
sed -i 's/^NEXT_PUBLIC_API_URL=.*/NEXT_PUBLIC_API_URL=http:\/\/52.206.225.24:8080\/api\/v1/' .env

# Clean everything
echo "🧹 Cleaning everything..."
rm -rf .next
rm -rf node_modules/.cache

# Rebuild
echo "🔨 Rebuilding..."
npm run build

# Copy static files
echo "📦 Copying static files..."
cp -r .next/static .next/standalone/.next/
cp -r public .next/standalone/

# Start frontend
echo "🚀 Starting frontend..."
cd .next/standalone
HOSTNAME=0.0.0.0 PORT=3000 nohup node server.js > /d02/logs/frontend.log 2>&1 &
echo $! > /d02/pids/frontend.pid

sleep 5

# Check if it started
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend started successfully!"
else
    echo "❌ Frontend failed to start. Check logs:"
    tail -10 /d02/logs/frontend.log
fi

