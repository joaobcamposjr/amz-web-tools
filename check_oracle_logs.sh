#!/bin/bash

echo "🔍 Checking Oracle connection logs..."

cd /d02/projects/amz-web-tools

echo "📋 Backend logs (last 50 lines):"
echo "=================================="
tail -50 /d02/logs/backend.log 2>/dev/null || echo "❌ Backend logs not found"

echo ""
echo "🔍 Looking for Oracle-related errors:"
echo "======================================"
grep -i "oracle" /d02/logs/backend.log 2>/dev/null || echo "No Oracle mentions found"

echo ""
echo "🔍 Looking for connection errors:"
echo "================================="
grep -i "connection\|error\|failed" /d02/logs/backend.log 2>/dev/null | tail -10 || echo "No connection errors found"

echo ""
echo "🔍 Current Oracle configuration:"
echo "================================"
grep -i "ORACLE" .env 2>/dev/null || echo "No Oracle config found in .env"

