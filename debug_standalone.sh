#!/bin/bash

echo "🔍 Debugging standalone directory structure..."

cd /d02/projects/amz-web-tools

echo "📁 Current directory:"
pwd

echo ""
echo "📁 Checking if .next/standalone exists:"
ls -la .next/ 2>/dev/null || echo "❌ .next directory not found"

echo ""
echo "📁 Checking standalone directory:"
ls -la .next/standalone/ 2>/dev/null || echo "❌ .next/standalone not found"

echo ""
echo "📁 Looking for server.js:"
find .next -name "server.js" 2>/dev/null || echo "❌ server.js not found"

echo ""
echo "📁 Checking if build was successful:"
if [ -d ".next/standalone" ]; then
    echo "✅ standalone directory exists"
    if [ -f ".next/standalone/server.js" ]; then
        echo "✅ server.js exists"
        ls -la .next/standalone/server.js
    else
        echo "❌ server.js missing"
    fi
else
    echo "❌ Build failed - no standalone directory"
fi

echo ""
echo "🔍 Checking build logs:"
echo "Last 10 lines of build output:"
tail -10 /d02/logs/frontend.log 2>/dev/null || echo "No frontend logs found"
