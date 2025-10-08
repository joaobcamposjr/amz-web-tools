#!/bin/bash

echo "🔧 Fixing Next.js standalone configuration..."

cd /d02/projects/amz-web-tools

# Stop everything
pkill -f "node.*standalone" 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
sleep 3

# Clean everything
rm -rf .next
rm -rf node_modules/.cache

# Build
echo "🔨 Building with fixed config..."
npm run build

# Fix standalone structure
echo "📦 Fixing standalone structure..."
cd .next/standalone

# Create proper structure
mkdir -p .next/static
mkdir -p public

# Copy static files correctly
cp -r ../static/* .next/static/ 2>/dev/null || true
cp -r ../../public/* public/ 2>/dev/null || true

# Verify structure
echo "🔍 Verifying structure..."
ls -la .next/static/ | head -5
ls -la public/ | head -5

# Start with proper environment
echo "🚀 Starting with fixed structure..."
cd /d02/projects/amz-web-tools/.next/standalone
HOSTNAME=0.0.0.0 PORT=3000 nohup node server.js > /d02/logs/frontend.log 2>&1 &

sleep 5

# Test
if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend started successfully!"
    echo "🎯 Test: http://52.206.225.24:3000"
else
    echo "❌ Frontend failed. Check logs:"
    tail -10 /d02/logs/frontend.log
fi
