#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

echo "🛑 Matando todos os processos nas portas 3000 e 8080..."

# Matar porta 3000
echo "   Matando processos na porta 3000..."
if command -v lsof >/dev/null 2>&1; then
    PIDS=$(lsof -ti:3000 2>/dev/null || true)
    if [ -n "$PIDS" ]; then
        echo "$PIDS" | xargs kill -9 2>/dev/null || true
        echo "   ✅ Processos na porta 3000 finalizados"
    else
        echo "   ℹ️  Nenhum processo na porta 3000"
    fi
fi

# Matar porta 8080
echo "   Matando processos na porta 8080..."
if command -v lsof >/dev/null 2>&1; then
    PIDS=$(lsof -ti:8080 2>/dev/null || true)
    if [ -n "$PIDS" ]; then
        echo "$PIDS" | xargs kill -9 2>/dev/null || true
        echo "   ✅ Processos na porta 8080 finalizados"
    else
        echo "   ℹ️  Nenhum processo na porta 8080"
    fi
fi

# Matar por nome de processo
echo "   Matando processos por nome..."
pkill -9 -f "$BASE/bin/backend" 2>/dev/null || true
pkill -9 -f "next" 2>/dev/null || true
pkill -9 -f "node.*next" 2>/dev/null || true
pkill -9 -f "node.*3000" 2>/dev/null || true

# Matar por PID dos arquivos
if [ -f "$LOGS/backend.pid" ]; then
    PID=$(cat "$LOGS/backend.pid" 2>/dev/null || echo "")
    if [ -n "$PID" ]; then
        kill -9 "$PID" 2>/dev/null || true
        echo "   ✅ Backend PID $PID finalizado"
    fi
fi

if [ -f "$LOGS/frontend.pid" ]; then
    PID=$(cat "$LOGS/frontend.pid" 2>/dev/null || echo "")
    if [ -n "$PID" ]; then
        kill -9 "$PID" 2>/dev/null || true
        echo "   ✅ Frontend PID $PID finalizado"
    fi
fi

sleep 2

echo ""
echo "🔍 Verificando portas..."
if command -v lsof >/dev/null 2>&1; then
    PORT3000=$(lsof -ti:3000 2>/dev/null || true)
    PORT8080=$(lsof -ti:8080 2>/dev/null || true)
    
    if [ -z "$PORT3000" ] && [ -z "$PORT8080" ]; then
        echo "✅ Portas 3000 e 8080 livres"
    else
        if [ -n "$PORT3000" ]; then
            echo "⚠️  Porta 3000 ainda em uso: $PORT3000"
        fi
        if [ -n "$PORT8080" ]; then
            echo "⚠️  Porta 8080 ainda em uso: $PORT8080"
        fi
    fi
else
    echo "ℹ️  lsof não disponível, não foi possível verificar"
fi

echo ""
echo "✅ Concluído!"

