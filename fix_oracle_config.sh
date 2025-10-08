#!/bin/bash

echo "🔧 Fixing Oracle configuration for server..."

cd /d02/projects/amz-web-tools

# Backup current config
echo "💾 Backing up current .env..."
cp .env .env.backup

# Fix Oracle configuration
echo "⚙️ Updating Oracle configuration..."
sed -i 's/^# ORACLE_HOST=164.152.40.38/ORACLE_HOST=164.152.40.38/' .env
sed -i 's/^ORACLE_HOST=10.13.2.159/# ORACLE_HOST=10.13.2.159/' .env
sed -i 's|ORACLE_LIB_DIR=/Applications/oracle/client/instantclient_23|ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7|' .env

echo "✅ Oracle configuration updated!"
echo ""
echo "📋 New Oracle config:"
grep -E "^ORACLE_" .env

echo ""
echo "🔄 Restarting backend to apply changes..."
pkill -f "bin/backend" 2>/dev/null || true
sleep 2

# Start backend
cd /d02/projects/amz-web-tools
nohup ./bin/backend > /d02/logs/backend.log 2>&1 &
echo $! > /d02/pids/backend.pid

sleep 3

echo "✅ Backend restarted with new Oracle config!"
echo ""
echo "🔍 Testing Oracle connection..."
sleep 2
tail -10 /d02/logs/backend.log | grep -i oracle || echo "No Oracle logs yet"
