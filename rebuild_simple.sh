#!/bin/bash

echo "🔨 Simple rebuild and restart..."

cd /d02/projects/amz-web-tools

# Stop frontend
pkill -f "node.*standalone" 2>/dev/null || true
sleep 2

# Build
npm run build

# Create standalone directory if it doesn't exist
mkdir -p .next/standalone/.next

# Copy files
cp -r .next/static .next/standalone/.next/
cp -r public .next/standalone/

# Start
cd .next/standalone
HOSTNAME=0.0.0.0 PORT=3000 nohup node server.js > /d02/logs/frontend.log 2>&1 &

echo "✅ Frontend rebuilt and started!"
