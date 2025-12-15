#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
mkdir -p "$LOGS" /d02/.cache/go-build /d02/go/pkg/mod

echo "🔧 Iniciando serviços..."

# Função para matar processos em uma porta
kill_port() {
    local port=$1
    echo "🛑 Matando processos na porta $port..."
    
    # Tenta matar por PID dos arquivos
    if [ -f "$LOGS/backend.pid" ] && [ "$port" == "8080" ]; then
        PID=$(cat "$LOGS/backend.pid" 2>/dev/null || echo "")
        if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
            kill -9 "$PID" 2>/dev/null || true
        fi
    fi
    
    if [ -f "$LOGS/frontend.pid" ] && [ "$port" == "3000" ]; then
        PID=$(cat "$LOGS/frontend.pid" 2>/dev/null || echo "")
        if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
            kill -9 "$PID" 2>/dev/null || true
        fi
    fi
    
    # Mata processos usando lsof se disponível
    if command -v lsof >/dev/null 2>&1; then
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
    fi
    
    # Mata processos usando fuser se disponível
    if command -v fuser >/dev/null 2>&1; then
        fuser -k $port/tcp 2>/dev/null || true
    fi
    
    # Mata processos por nome
    pkill -f "$BASE/bin/backend" 2>/dev/null || true
    pkill -f "next" 2>/dev/null || true
    pkill -f "node.*next" 2>/dev/null || true
    pkill -f "node.*3000" 2>/dev/null || true
    
    # Aguarda para garantir que os processos foram finalizados
    sleep 3
    
    # Verifica se a porta ainda está em uso
    if command -v lsof >/dev/null 2>&1; then
        if lsof -ti:$port >/dev/null 2>&1; then
            echo "⚠️  Porta $port ainda está em uso, tentando novamente..."
            sleep 2
            lsof -ti:$port | xargs kill -9 2>/dev/null || true
            sleep 2
        fi
    fi
}

echo "🛑 Parando serviços antigos..."
kill_port 8080
kill_port 3000

echo ""
echo "🚀 Iniciando Backend..."
cd "$BASE"

# Build do backend
if [ -d backend/cmd/server ]; then 
    SRC=./backend/cmd/server
else 
    SRC=./backend
fi

GOCACHE=/d02/.cache/go-build GOMODCACHE=/d02/go/pkg/mod go build -o bin/backend "$SRC"

# Carregar variáveis de ambiente do .env
if [ -f "$BASE/.env" ]; then
    echo "📋 Carregando variáveis de ambiente do .env..."
    set -a
    source "$BASE/.env"
    set +a
    
    export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
    export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}
fi

# Iniciar backend
nohup env $(grep -v '^#' "$BASE/.env" 2>/dev/null | grep -v '^$' | xargs) \
    ORACLE_LIB_DIR="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}" \
    LD_LIBRARY_PATH="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}:${LD_LIBRARY_PATH:-}" \
    "$BASE/bin/backend" >> "$LOGS/backend.log" 2>&1 & 
echo $! > "$LOGS/backend.pid"

sleep 2

echo ""
echo "🚀 Iniciando Frontend..."
cd "$BASE"

export NODE_ENV=production
export NEXT_PUBLIC_API_URL=/api/v1
export PORT=3000

# Limpar cache
echo "🗑️ Limpando cache Next.js..."
if [ -d .next ]; then
    sudo rm -rf .next 2>/dev/null || rm -rf .next
fi

# Corrigir permissões
echo "🔧 Corrigindo permissões..."
sudo chown -R $(whoami):$(whoami) . 2>/dev/null || true
chmod -R 755 . 2>/dev/null || true

# Build e start do frontend
npm run build
nohup npm start >> "$LOGS/frontend.log" 2>&1 &
echo $! > "$LOGS/frontend.pid"

sleep 3

echo ""
echo "🔍 Testando serviços..."
echo ""
echo "Backend:"
curl -s http://localhost:8080/api/v1/health || echo "❌ Backend não respondeu"

echo ""
echo "Frontend:"
curl -s -I http://localhost:3000 | head -1 || echo "❌ Frontend não respondeu"

echo ""
echo ""
echo "✅ Serviços iniciados!"
echo ""
echo "📄 Logs:"
echo "   Backend:  tail -f $LOGS/backend.log"
echo "   Frontend: tail -f $LOGS/frontend.log"




