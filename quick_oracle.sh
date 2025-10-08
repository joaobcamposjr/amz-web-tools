#!/bin/bash

# Quick Oracle fix
cd /d02/projects/amz-web-tools

# Pull latest
git pull

# Stop backend
pkill -f "backend" 2>/dev/null || true
sleep 3

# Build and start
cd backend
export ORACLE_LIB_DIR=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_7:${LD_LIBRARY_PATH:-}
go build -o bin/backend .

export ORACLE_HOST=164.152.40.38
export ORACLE_USER=dashjc
export ORACLE_SERVICE=nbs
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

nohup ./bin/backend > /d02/logs/backend.log 2>&1 &
echo $! > /d02/logs/backend.pid

sleep 5
echo "✅ Backend started with Oracle config"
echo "Test: http://52.206.225.24:3000"
