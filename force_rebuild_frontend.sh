#!/bin/bash

echo "🔄 Force rebuilding frontend with correct API URL..."

cd /d02/projects/amz-web-tools

# Stop frontend
echo "🛑 Stopping frontend..."
pkill -f "node.*standalone.*server.js" 2>/dev/null || true
sleep 2

# Set environment variable
echo "⚙️ Setting NEXT_PUBLIC_API_URL..."
export NEXT_PUBLIC_API_URL="http://52.206.225.24:8080/api/v1"
echo "NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL" > .env.local

# Update .env file
sed -i 's/^NEXT_PUBLIC_API_URL=.*/NEXT_PUBLIC_API_URL=http:\/\/52.206.225.24:8080\/api\/v1/' .env

# Clean build
echo "🧹 Cleaning build..."
rm -rf .next
rm -rf node_modules/.cache

# Rebuild
echo "🔨 Rebuilding frontend..."
npm run build

# Copy static files
echo "📦 Copying static files..."
cp -r .next/static .next/standalone/.next/
cp -r public .next/standalone/

# Start frontend
echo "🚀 Starting frontend..."
cd .next/standalone
HOSTNAME=0.0.0.0 PORT=3000 nohup node server.js > /d02/logs/frontend.log 2>&1 &

echo "✅ Frontend rebuilt and started!"
echo ""
echo "🎯 Now frontend should call: http://52.206.225.24:8080/api/v1/..."
echo ""
echo "📋 Check logs:"
echo "tail -f /d02/logs/frontend.log"


