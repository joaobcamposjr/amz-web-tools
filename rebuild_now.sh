#!/bin/bash

echo "🚀 FINAL FIX - Rebuilding frontend with hardcoded URL..."

cd /d02/projects/amz-web-tools

# Stop everything
pkill -f "node.*standalone" 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
sleep 3

# Clean and rebuild
rm -rf .next
npm run build

# Copy files
cp -r .next/static .next/standalone/.next/
cp -r public .next/standalone/

# Start
cd .next/standalone
HOSTNAME=0.0.0.0 PORT=3000 nohup node server.js > /d02/logs/frontend.log 2>&1 &

echo "✅ DONE! Frontend should now call: http://52.206.225.24:8080/api/v1"
