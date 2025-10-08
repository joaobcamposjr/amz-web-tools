#!/bin/bash

echo "🔍 Debugging environment variable loading..."

cd /d02/projects/amz-web-tools

echo "📋 Current .env file content:"
echo "============================="
grep -E "^ORACLE_" .env

echo ""
echo "🔍 Checking if backend process is running:"
ps aux | grep "bin/backend" | grep -v grep || echo "❌ Backend not running"

echo ""
echo "🔍 Checking backend binary:"
ls -la bin/backend 2>/dev/null || echo "❌ Backend binary not found"

echo ""
echo "🔍 Testing environment variable loading:"
echo "========================================"
export $(grep -v '^#' .env | xargs)
echo "ORACLE_HOST: '$ORACLE_HOST'"
echo "ORACLE_USER: '$ORACLE_USER'"
echo "ORACLE_SERVICE: '$ORACLE_SERVICE'"
echo "ORACLE_LIB_DIR: '$ORACLE_LIB_DIR'"

echo ""
echo "🔍 Checking if backend is reading .env from correct location:"
pwd
echo "Backend working directory should be: /d02/projects/amz-web-tools"

echo ""
echo "🔍 Checking if .env file is accessible:"
if [ -f ".env" ]; then
    echo "✅ .env file exists and is readable"
    echo "File size: $(wc -c < .env) bytes"
    echo "File permissions: $(ls -la .env)"
else
    echo "❌ .env file not found or not accessible"
fi
