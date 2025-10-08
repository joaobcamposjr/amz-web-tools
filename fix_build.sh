#!/bin/bash

echo "🔧 Fixing Next.js build to create standalone..."

cd /d02/projects/amz-web-tools

# Stop everything
pkill -f "node.*server" 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
sleep 3

# Clean everything
echo "🧹 Cleaning everything..."
rm -rf .next
rm -rf node_modules/.cache

# Check next.config.js
echo "🔍 Checking next.config.js..."
cat next.config.js

echo ""
echo "🔨 Building with standalone output..."

# Build with explicit standalone
NEXT_PUBLIC_API_URL="http://52.206.225.24:8080/api/v1" npm run build

echo ""
echo "🔍 Checking if standalone was created..."
if [ -d ".next/standalone" ]; then
    echo "✅ standalone directory created!"
    ls -la .next/standalone/
    
    if [ -f ".next/standalone/server.js" ]; then
        echo "✅ server.js exists!"
        
        # Copy static files
        echo "📦 Copying static files..."
        cp -r .next/static .next/standalone/.next/
        cp -r public .next/standalone/
        
        # Start
        echo "🚀 Starting frontend..."
        cd .next/standalone
        HOSTNAME=0.0.0.0 PORT=3000 nohup node server.js > /d02/logs/frontend.log 2>&1 &
        
        sleep 5
        
        # Test
        if curl -s http://localhost:3000 > /dev/null; then
            echo "✅ Frontend started successfully!"
        else
            echo "❌ Frontend failed. Check logs:"
            tail -10 /d02/logs/frontend.log
        fi
    else
        echo "❌ server.js not found in standalone"
    fi
else
    echo "❌ standalone directory not created!"
    echo "🔍 Available in .next:"
    ls -la .next/
fi

