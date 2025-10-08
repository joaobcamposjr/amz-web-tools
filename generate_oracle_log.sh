#!/bin/bash

echo "🔍 Generating Oracle debug log..."

cd /d02/projects/amz-web-tools

# Create a log file
LOG_FILE="oracle_debug.log"
echo "Oracle Debug Log - $(date)" > $LOG_FILE
echo "=================================" >> $LOG_FILE

echo "📋 Current .env file content:" >> $LOG_FILE
echo "=============================" >> $LOG_FILE
cat .env >> $LOG_FILE

echo "" >> $LOG_FILE
echo "🔍 Backend process info:" >> $LOG_FILE
echo "========================" >> $LOG_FILE
ps aux | grep "bin/backend" | grep -v grep >> $LOG_FILE

echo "" >> $LOG_FILE
echo "🔍 Backend logs (last 20 lines):" >> $LOG_FILE
echo "================================" >> $LOG_FILE
tail -20 /d02/logs/backend.log >> $LOG_FILE

echo "" >> $LOG_FILE
echo "🔍 Environment variables test:" >> $LOG_FILE
echo "==============================" >> $LOG_FILE
echo "ORACLE_HOST: '$ORACLE_HOST'" >> $LOG_FILE
echo "ORACLE_USER: '$ORACLE_USER'" >> $LOG_FILE
echo "ORACLE_SERVICE: '$ORACLE_SERVICE'" >> $LOG_FILE

echo "" >> $LOG_FILE
echo "🔍 File permissions:" >> $LOG_FILE
echo "====================" >> $LOG_FILE
ls -la .env >> $LOG_FILE
ls -la bin/backend >> $LOG_FILE

echo "" >> $LOG_FILE
echo "🔍 Current directory:" >> $LOG_FILE
echo "====================" >> $LOG_FILE
pwd >> $LOG_FILE

echo "✅ Log file created: $LOG_FILE"
echo "📋 Copy and paste the content below:"
echo "===================================="
cat $LOG_FILE
