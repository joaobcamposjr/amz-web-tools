#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

echo "💀 Matando TODOS os processos relacionados (modo agressivo)..."

# Matar por porta (múltiplas tentativas)
for port in 3000 8080; do
    echo "   Matando processos na porta $port..."
    for i in {1..10}; do
        if command -v lsof >/dev/null 2>&1; then
            PIDS=$(lsof -ti:$port 2>/dev/null || true)
            if [ -n "$PIDS" ]; then
                echo "      Tentativa $i: Encontrados PIDs: $PIDS"
                echo "$PIDS" | xargs kill -9 2>/dev/null || true
                sleep 1
            else
                echo "      ✅ Porta $port está livre"
                break
            fi
        fi
    done
    
    if command -v fuser >/dev/null 2>&1; then
        fuser -k $port/tcp 2>/dev/null || true
    fi
done

# Matar por nome de processo
echo "   Matando processos por nome..."
pkill -9 -f "$BASE/bin/backend" 2>/dev/null || true
pkill -9 -f "next" 2>/dev/null || true
pkill -9 -f "node.*next" 2>/dev/null || true
pkill -9 -f "node.*3000" 2>/dev/null || true
pkill -9 -f "next-server" 2>/dev/null || true
pkill -9 -f "npm.*build" 2>/dev/null || true
pkill -9 -f "npm.*start" 2>/dev/null || true

# Matar por PID dos arquivos
if [ -f "$LOGS/backend.pid" ]; then
    PID=$(cat "$LOGS/backend.pid" 2>/dev/null || echo "")
    if [ -n "$PID" ]; then
        echo "   Matando backend PID: $PID"
        kill -9 "$PID" 2>/dev/null || true
        rm -f "$LOGS/backend.pid"
    fi
fi

if [ -f "$LOGS/frontend.pid" ]; then
    PID=$(cat "$LOGS/frontend.pid" 2>/dev/null || echo "")
    if [ -n "$PID" ]; then
        echo "   Matando frontend PID: $PID"
        kill -9 "$PID" 2>/dev/null || true
        rm -f "$LOGS/frontend.pid"
    fi
fi

sleep 2

# Verificação final
echo ""
echo "🔍 Verificando processos restantes..."
if command -v lsof >/dev/null 2>&1; then
    PORT3000=$(lsof -ti:3000 2>/dev/null || true)
    PORT8080=$(lsof -ti:8080 2>/dev/null || true)
    
    if [ -n "$PORT3000" ]; then
        echo "⚠️  Porta 3000 ainda em uso por: $PORT3000"
        echo "   Tentando matar novamente..."
        echo "$PORT3000" | xargs kill -9 2>/dev/null || true
    fi
    
    if [ -n "$PORT8080" ]; then
        echo "⚠️  Porta 8080 ainda em uso por: $PORT8080"
        echo "   Tentando matar novamente..."
        echo "$PORT8080" | xargs kill -9 2>/dev/null || true
    fi
    
    if [ -z "$PORT3000" ] && [ -z "$PORT8080" ]; then
        echo "✅ Todas as portas estão livres"
    fi
fi

# Mostrar processos relacionados que ainda estão rodando
echo ""
echo "🔍 Processos relacionados ainda rodando:"
ps aux | grep -E "backend|next|node.*3000" | grep -v grep || echo "   Nenhum processo encontrado"

echo ""
echo "✅ Concluído!"

