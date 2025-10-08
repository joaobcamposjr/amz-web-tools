#!/bin/bash

echo "🔍 Checking Backend Status..."
echo ""

# Check if backend process is running
echo "📊 Backend Process:"
ps aux | grep -E "backend|8080" | grep -v grep || echo "❌ No backend process found"
echo ""

# Check if port 8080 is open
echo "🔌 Port 8080 Status:"
if lsof -i:8080 >/dev/null 2>&1; then
    echo "✅ Port 8080 is in use:"
    lsof -i:8080
else
    echo "❌ Port 8080 is NOT in use"
fi
echo ""

# Check backend logs
echo "📋 Backend Logs (last 30 lines):"
tail -30 /d02/logs/backend.log
echo ""

# Test backend health endpoint
echo "🏥 Testing Backend Health:"
curl -v http://localhost:8080/health 2>&1 | tail -20
echo ""

# Test backend from external IP
echo "🌐 Testing Backend from External IP:"
curl -v http://52.206.225.24:8080/health 2>&1 | tail -20
echo ""

# Check .env configuration
echo "⚙️ Backend Configuration:"
echo "SERVER_HOST: $(grep SERVER_HOST /d02/projects/amz-web-tools/.env | cut -d'=' -f2)"
echo "SERVER_PORT: $(grep SERVER_PORT /d02/projects/amz-web-tools/.env | cut -d'=' -f2)"
echo "CORS_ALLOWED_ORIGINS: $(grep CORS_ALLOWED_ORIGINS /d02/projects/amz-web-tools/.env | cut -d'=' -f2)"
echo ""

# Check if binary exists
echo "📦 Backend Binary:"
if [ -f /d02/projects/amz-web-tools/bin/backend ]; then
    echo "✅ Backend binary exists"
    ls -lh /d02/projects/amz-web-tools/bin/backend
else
    echo "❌ Backend binary NOT found"
fi


