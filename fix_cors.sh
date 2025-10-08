#!/bin/bash

echo "🔧 Fixing CORS configuration..."

ENV_FILE="/d02/projects/amz-web-tools/.env"

# Backup current .env
cp "$ENV_FILE" "$ENV_FILE.backup"

# Update CORS_ALLOWED_ORIGINS
sed -i 's/^CORS_ALLOWED_ORIGINS=.*/CORS_ALLOWED_ORIGINS=http:\/\/52.206.225.24:3000,http:\/\/52.206.225.24,https:\/\/52.206.225.24:3000,https:\/\/52.206.225.24,http:\/\/localhost:3000/' "$ENV_FILE"

echo "✅ CORS updated!"
echo ""
echo "📋 New CORS configuration:"
grep CORS_ALLOWED_ORIGINS "$ENV_FILE"
echo ""
echo "🔄 Please restart backend for changes to take effect:"
echo "   cd /d02/projects/amz-web-tools"
echo "   pkill -f backend"
echo "   nohup ./bin/backend > /d02/logs/backend.log 2>&1 & echo \$! > /d02/pids/backend.pid"

