#!/bin/bash

echo "🛑 Stopping AMZ Web Tools..."

PID_DIR="/d02/pids"

# Kill by PID files
if [ -f "$PID_DIR/backend.pid" ]; then
    PID=$(cat "$PID_DIR/backend.pid")
    echo "🔫 Killing backend (PID: $PID)"
    kill -9 $PID 2>/dev/null || true
    rm -f "$PID_DIR/backend.pid"
fi

if [ -f "$PID_DIR/frontend.pid" ]; then
    PID=$(cat "$PID_DIR/frontend.pid")
    echo "🔫 Killing frontend (PID: $PID)"
    kill -9 $PID 2>/dev/null || true
    rm -f "$PID_DIR/frontend.pid"
fi

# Kill by port
echo "🔫 Killing processes on ports..."
fuser -k 8080/tcp 2>/dev/null || true
fuser -k 3000/tcp 2>/dev/null || true
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
lsof -ti:3000 | xargs kill -9 2>/dev/null || true

# Kill by process name
echo "🔫 Killing processes by name..."
pkill -9 -f "backend" 2>/dev/null || true
pkill -9 -f "node.*standalone.*server.js" 2>/dev/null || true
pkill -9 -f "node .next/standalone/server.js" 2>/dev/null || true

# Extra aggressive cleanup
echo "🔫 Extra aggressive cleanup..."
ps aux | grep "server.js" | grep -v grep | awk '{print $2}' | xargs kill -9 2>/dev/null || true
ps aux | grep "backend" | grep -v grep | awk '{print $2}' | xargs kill -9 2>/dev/null || true

sleep 2

# Verify ports are free
echo ""
echo "🔍 Verifying ports are free..."
if lsof -i:8080 >/dev/null 2>&1; then
    echo "⚠️ Port 8080 still in use:"
    lsof -i:8080
else
    echo "✅ Port 8080 is free"
fi

if lsof -i:3000 >/dev/null 2>&1; then
    echo "⚠️ Port 3000 still in use:"
    lsof -i:3000
else
    echo "✅ Port 3000 is free"
fi

echo ""
echo "✅ All services stopped"
