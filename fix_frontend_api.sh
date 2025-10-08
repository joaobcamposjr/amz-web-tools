#!/bin/bash

echo "🔧 Fixing frontend API configuration..."

ENV_FILE="/d02/projects/amz-web-tools/.env"

# Check current NEXT_PUBLIC_API_URL
echo "📋 Current NEXT_PUBLIC_API_URL:"
grep NEXT_PUBLIC_API_URL "$ENV_FILE" || echo "❌ NOT FOUND"

echo ""
echo "🔧 Setting correct NEXT_PUBLIC_API_URL..."

# Update NEXT_PUBLIC_API_URL to use backend port 8080
sed -i 's/^NEXT_PUBLIC_API_URL=.*/NEXT_PUBLIC_API_URL=http:\/\/52.206.225.24:8080\/api\/v1/' "$ENV_FILE"

echo "✅ Updated NEXT_PUBLIC_API_URL!"
echo ""
echo "📋 New configuration:"
grep NEXT_PUBLIC_API_URL "$ENV_FILE"

echo ""
echo "🔄 Now rebuilding frontend..."
cd /d02/projects/amz-web-tools
npm run build

echo ""
echo "📦 Copying static files..."
cp -r .next/static .next/standalone/.next/
cp -r public .next/standalone/

echo ""
echo "🔄 Restarting frontend..."
pkill -f "node.*standalone.*server.js" 2>/dev/null || true
cd .next/standalone
HOSTNAME=0.0.0.0 PORT=3000 nohup node server.js > /d02/logs/frontend.log 2>&1 &

echo "✅ Frontend restarted!"
echo ""
echo "🎯 Frontend should now call: http://52.206.225.24:8080/api/v1/..."

