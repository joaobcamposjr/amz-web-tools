#!/bin/bash

echo "🔍 Checking Next.js standalone structure..."

BASE="/d02/projects/amz-web-tools"

echo ""
echo "📁 Checking .next/standalone directory:"
ls -la "$BASE/.next/standalone/" | head -20

echo ""
echo "📁 Checking .next/standalone/.next directory:"
ls -la "$BASE/.next/standalone/.next/" 2>/dev/null | head -20

echo ""
echo "📁 Checking .next/standalone/.next/static directory:"
ls -la "$BASE/.next/standalone/.next/static/" 2>/dev/null | head -20

echo ""
echo "📁 Checking .next/standalone/public directory:"
ls -la "$BASE/.next/standalone/public/" 2>/dev/null | head -20

echo ""
echo "📄 Checking if server.js exists:"
if [ -f "$BASE/.next/standalone/server.js" ]; then
    echo "✅ server.js exists"
else
    echo "❌ server.js NOT FOUND!"
fi

echo ""
echo "📄 Checking frontend logs:"
tail -50 /d02/logs/frontend.log


