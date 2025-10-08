#!/bin/bash

# Start frontend
cd /d02/projects/amz-web-tools

echo "🚀 Starting frontend..."

# Kill any existing frontend processes
pkill -f "next" 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
sleep 3

# Check if standalone exists
if [ -d ".next/standalone" ]; then
    echo "✅ Standalone directory found"
    cd .next/standalone
    
    # Check if server.js exists
    if [ -f "server.js" ]; then
        echo "✅ server.js found"
        
        # Start frontend
        nohup node server.js > /d02/logs/frontend.log 2>&1 &
        echo $! > /d02/logs/frontend.pid
        
        sleep 10
        
        # Test
        if curl -s http://localhost:3000 > /dev/null; then
            echo "✅ Frontend started successfully"
        else
            echo "❌ Frontend failed to start"
            echo "📄 Frontend logs:"
            tail -10 /d02/logs/frontend.log
        fi
    else
        echo "❌ server.js not found in standalone"
        echo "📁 Standalone contents:"
        ls -la
    fi
else
    echo "❌ Standalone directory not found"
    echo "📁 .next contents:"
    ls -la .next/ 2>/dev/null || echo "No .next directory"
    
    # Try to build frontend
    echo "🔨 Building frontend..."
    npm run build
    
    if [ -d ".next/standalone" ]; then
        echo "✅ Frontend built successfully"
        cd .next/standalone
        nohup node server.js > /d02/logs/frontend.log 2>&1 &
        echo $! > /d02/logs/frontend.pid
        sleep 10
        echo "✅ Frontend started"
    else
        echo "❌ Frontend build failed"
    fi
fi

echo "✅ Frontend start completed!"
echo "Test: http://52.206.225.24:3000"
