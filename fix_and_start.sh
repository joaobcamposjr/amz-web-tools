#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

echo "🔧 Corrigindo e iniciando serviços..."

# Função para matar processos em uma porta
kill_port() {
    local port=$1
    echo "🛑 Matando processos na porta $port..."
    
    # Tentar múltiplos métodos
    if command -v lsof >/dev/null 2>&1; then
        lsof -ti:$port 2>/dev/null | xargs kill -9 2>/dev/null || true
    fi
    
    if command -v fuser >/dev/null 2>&1; then
        fuser -k $port/tcp 2>/dev/null || true
    fi
    
    # Matar por PID dos arquivos
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
    
    # Matar por nome de processo
    pkill -f "$BASE/bin/backend" 2>/dev/null || true
    pkill -f "next" 2>/dev/null || true
    pkill -f "node.*next" 2>/dev/null || true
    pkill -f "node.*3000" 2>/dev/null || true
    
    sleep 2
    
    # Verificar se ainda está em uso
    if command -v lsof >/dev/null 2>&1; then
        if lsof -ti:$port >/dev/null 2>&1; then
            echo "⚠️  Porta $port ainda em uso, tentando novamente..."
            sleep 2
            lsof -ti:$port | xargs kill -9 2>/dev/null || true
            sleep 2
        fi
    fi
}

echo "🛑 Parando processos antigos..."
kill_port 8080
kill_port 3000

# Verificar conectividade com o banco
echo ""
echo "🔍 Verificando conectividade com o banco de dados..."
if [ -f "$BASE/.env" ]; then
    source "$BASE/.env"
    
    # Testar conexão com sqlcmd se disponível
    if command -v sqlcmd >/dev/null 2>&1; then
        echo "   Testando conexão com sqlcmd..."
        sqlcmd -S "$DB_HOST" -U "$DB_USER" -P "$DB_PASSWORD" -d "$DB_NAME" -C -Q "SELECT 1" >/dev/null 2>&1 && {
            echo "✅ Conexão com banco OK"
        } || {
            echo "⚠️  Aviso: Não foi possível conectar ao banco via sqlcmd"
            echo "   Isso pode ser normal se o sqlcmd não estiver disponível"
        }
    else
        echo "   sqlcmd não disponível, pulando teste de conexão"
    fi
    
    # Testar conectividade básica de rede
    if command -v nc >/dev/null 2>&1 || command -v telnet >/dev/null 2>&1; then
        echo "   Testando conectividade de rede na porta 1433..."
        if command -v nc >/dev/null 2>&1; then
            timeout 5 nc -z "$DB_HOST" 1433 2>/dev/null && {
                echo "✅ Porta 1433 acessível"
            } || {
                echo "⚠️  Aviso: Porta 1433 não acessível (timeout ou recusada)"
            }
        fi
    fi
fi

echo ""
echo "🚀 Iniciando Backend..."
cd "$BASE"

# Build do backend
if [ -d backend/cmd/server ]; then 
    SRC=./backend/cmd/server
else 
    SRC=./backend
fi

echo "   Compilando backend..."
GOCACHE=/d02/.cache/go-build GOMODCACHE=/d02/go/pkg/mod go build -o bin/backend "$SRC"

# Carregar variáveis de ambiente
if [ -f "$BASE/.env" ]; then
    set -a
    source "$BASE/.env"
    set +a
    
    export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
    export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}
fi

# Iniciar backend
echo "   Iniciando backend..."
nohup env $(grep -v '^#' "$BASE/.env" 2>/dev/null | grep -v '^$' | xargs) \
    ORACLE_LIB_DIR="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}" \
    LD_LIBRARY_PATH="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}:${LD_LIBRARY_PATH:-}" \
    "$BASE/bin/backend" >> "$LOGS/backend.log" 2>&1 & 
echo $! > "$LOGS/backend.pid"

sleep 3

# Verificar se backend está rodando
if kill -0 $(cat "$LOGS/backend.pid") 2>/dev/null; then
    echo "✅ Backend iniciado (PID: $(cat $LOGS/backend.pid))"
else
    echo "⚠️  Backend pode não ter iniciado corretamente, verifique os logs:"
    echo "   tail -f $LOGS/backend.log"
fi

echo ""
echo "🚀 Iniciando Frontend..."
cd "$BASE"

export NODE_ENV=production
export NEXT_PUBLIC_API_URL=/api/v1
export PORT=3000

# Limpar cache
echo "   Limpando cache Next.js..."
if [ -d .next ]; then
    sudo rm -rf .next 2>/dev/null || rm -rf .next
fi

# Corrigir permissões
echo "   Corrigindo permissões..."
sudo chown -R $(whoami):$(whoami) . 2>/dev/null || true
chmod -R 755 . 2>/dev/null || true

# Build e start do frontend
echo "   Fazendo build do frontend..."
npm run build

echo "   Iniciando frontend..."
nohup npm start >> "$LOGS/frontend.log" 2>&1 &
echo $! > "$LOGS/frontend.pid"

sleep 3

# Verificar se frontend está rodando
if kill -0 $(cat "$LOGS/frontend.pid") 2>/dev/null; then
    echo "✅ Frontend iniciado (PID: $(cat $LOGS/frontend.pid))"
else
    echo "⚠️  Frontend pode não ter iniciado corretamente, verifique os logs:"
    echo "   tail -f $LOGS/frontend.log"
fi

echo ""
echo "🔍 Verificando portas..."
if command -v ss >/dev/null 2>&1; then
    ss -ltnp | grep -E ':8080|:3000' || echo "   Nenhuma porta encontrada (pode levar alguns segundos para aparecer)"
elif command -v netstat >/dev/null 2>&1; then
    netstat -tlnp | grep -E ':8080|:3000' || echo "   Nenhuma porta encontrada (pode levar alguns segundos para aparecer)"
fi

echo ""
echo "✅ Processo concluído!"
echo ""
echo "📄 Logs:"
echo "   Backend:  tail -f $LOGS/backend.log"
echo "   Frontend: tail -f $LOGS/frontend.log"
echo ""
echo "🔍 Para verificar processos:"
echo "   ps aux | grep -E 'backend|next'"

