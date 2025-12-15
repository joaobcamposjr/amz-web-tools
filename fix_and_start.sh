#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

echo "🔧 Corrigindo e iniciando serviços..."

# Função para matar processos em uma porta de forma mais agressiva
kill_port() {
    local port=$1
    echo "🛑 Matando processos na porta $port..."
    
    # Loop para matar múltiplas vezes
    for i in {1..5}; do
        # Matar por lsof
        if command -v lsof >/dev/null 2>&1; then
            PIDS=$(lsof -ti:$port 2>/dev/null || true)
            if [ -n "$PIDS" ]; then
                echo "$PIDS" | xargs kill -9 2>/dev/null || true
                sleep 1
            fi
        fi
        
        # Matar por fuser
        if command -v fuser >/dev/null 2>&1; then
            fuser -k $port/tcp 2>/dev/null || true
            sleep 1
        fi
        
        # Verificar se ainda está em uso
        if command -v lsof >/dev/null 2>&1; then
            if ! lsof -ti:$port >/dev/null 2>&1; then
                break  # Porta está livre
            fi
        fi
    done
    
    # Matar por PID dos arquivos
    if [ -f "$LOGS/backend.pid" ] && [ "$port" == "8080" ]; then
        PID=$(cat "$LOGS/backend.pid" 2>/dev/null || echo "")
        if [ -n "$PID" ]; then
            kill -9 "$PID" 2>/dev/null || true
            rm -f "$LOGS/backend.pid"
        fi
    fi
    
    if [ -f "$LOGS/frontend.pid" ] && [ "$port" == "3000" ]; then
        PID=$(cat "$LOGS/frontend.pid" 2>/dev/null || echo "")
        if [ -n "$PID" ]; then
            kill -9 "$PID" 2>/dev/null || true
            rm -f "$LOGS/frontend.pid"
        fi
    fi
    
    # Matar por nome de processo (mais agressivo)
    pkill -9 -f "$BASE/bin/backend" 2>/dev/null || true
    pkill -9 -f "next" 2>/dev/null || true
    pkill -9 -f "node.*next" 2>/dev/null || true
    pkill -9 -f "node.*3000" 2>/dev/null || true
    pkill -9 -f "next-server" 2>/dev/null || true
    
    sleep 3
    
    # Verificação final
    if command -v lsof >/dev/null 2>&1; then
        REMAINING=$(lsof -ti:$port 2>/dev/null || true)
        if [ -n "$REMAINING" ]; then
            echo "⚠️  Porta $port ainda em uso após múltiplas tentativas: $REMAINING"
            echo "   Tentando matar novamente..."
            echo "$REMAINING" | xargs kill -9 2>/dev/null || true
            sleep 2
        else
            echo "✅ Porta $port está livre"
        fi
    fi
}

echo "🛑 Parando processos antigos..."

# Matar tudo primeiro antes de verificar portas
echo "   Matando todos os processos relacionados..."
pkill -9 -f "$BASE/bin/backend" 2>/dev/null || true
pkill -9 -f "next" 2>/dev/null || true
pkill -9 -f "node.*next" 2>/dev/null || true
pkill -9 -f "node.*3000" 2>/dev/null || true
pkill -9 -f "next-server" 2>/dev/null || true

# Limpar PIDs antigos
rm -f "$LOGS/backend.pid" "$LOGS/frontend.pid"

sleep 2

# Agora matar por porta
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

# Verificar e corrigir fontes do Google antes do build
LAYOUT_FILE="$BASE/src/app/layout.tsx"
if grep -q "next/font/google" "$LAYOUT_FILE" 2>/dev/null; then
    echo "   🔧 Removendo Google Fonts do layout.tsx..."
    
    # Backup
    cp "$LAYOUT_FILE" "${LAYOUT_FILE}.backup.$(date +%Y%m%d_%H%M%S)" 2>/dev/null || true
    
    # Remover import do Google Fonts
    sed -i.tmp '/import { Inter } from .next\/font\/google/d' "$LAYOUT_FILE"
    sed -i.tmp '/^const inter = Inter({ subsets: .latin. })/d' "$LAYOUT_FILE"
    sed -i.tmp 's/${inter.className}//g' "$LAYOUT_FILE"
    sed -i.tmp 's/className={`/className="/g' "$LAYOUT_FILE"
    sed -i.tmp 's/`} bg-gray-50 antialiased`/" bg-gray-50 antialiased font-sans"/g' "$LAYOUT_FILE"
    sed -i.tmp 's/className="bg-gray-50 antialiased`/"bg-gray-50 antialiased font-sans"/g' "$LAYOUT_FILE"
    
    # Garantir que o body tenha a classe correta
    if ! grep -q 'className="bg-gray-50 antialiased font-sans"' "$LAYOUT_FILE"; then
        sed -i.tmp 's/<body className="[^"]*">/<body className="bg-gray-50 antialiased font-sans">/g' "$LAYOUT_FILE"
    fi
    
    rm -f "${LAYOUT_FILE}.tmp"
    echo "   ✅ Google Fonts removido"
fi

# Adicionar fontes do sistema no globals.css se necessário
GLOBALS_FILE="$BASE/src/app/globals.css"
if [ -f "$GLOBALS_FILE" ] && ! grep -q "font-family:" "$GLOBALS_FILE" 2>/dev/null; then
    echo "   🔧 Adicionando fontes do sistema no globals.css..."
    sed -i.tmp '/@apply bg-gray-50 text-gray-900;/a\
    font-family: -apple-system, BlinkMacSystemFont, '\''Segoe UI'\'', '\''Roboto'\'', '\''Oxygen'\'', '\''Ubuntu'\'', '\''Cantarell'\'', '\''Fira Sans'\'', '\''Droid Sans'\'', '\''Helvetica Neue'\'', sans-serif;\
    -webkit-font-smoothing: antialiased;\
    -moz-osx-font-smoothing: grayscale;' "$GLOBALS_FILE"
    rm -f "${GLOBALS_FILE}.tmp"
    echo "   ✅ Fontes do sistema adicionadas"
fi

# Limpar cache
echo "   Limpando cache Next.js..."
if [ -d .next ]; then
    sudo rm -rf .next 2>/dev/null || rm -rf .next
fi

# Limpar também cache do node_modules
if [ -d node_modules/.cache ]; then
    rm -rf node_modules/.cache
fi

# Corrigir permissões
echo "   Corrigindo permissões..."
sudo chown -R $(whoami):$(whoami) . 2>/dev/null || true
chmod -R 755 . 2>/dev/null || true

# Build e start do frontend
echo "   Fazendo build do frontend (pode levar alguns minutos)..."
echo "   ⚠️  Se travar, você pode verificar o progresso com: tail -f $LOGS/frontend_build.log"

# Fazer build com timeout de 20 minutos e desabilitar download de fontes
NEXT_TELEMETRY_DISABLED=1 npm run build > "$LOGS/frontend_build.log" 2>&1 &
BUILD_PID=$!

# Aguardar com timeout de 20 minutos
TIMEOUT=1200
ELAPSED=0
while kill -0 $BUILD_PID 2>/dev/null; do
    sleep 5
    ELAPSED=$((ELAPSED + 5))
    
    # A cada minuto, mostrar progresso
    if [ $((ELAPSED % 60)) -eq 0 ]; then
        echo "   ⏱️  Build em andamento... ${ELAPSED}s decorridos"
    fi
    
    if [ $ELAPSED -ge $TIMEOUT ]; then
        echo "   ❌ Build excedeu 20 minutos, cancelando..."
        kill -9 $BUILD_PID 2>/dev/null || true
        echo "❌ Build do frontend excedeu 20 minutos e foi cancelado"
        echo "   Verifique os logs: tail -f $LOGS/frontend_build.log"
        exit 1
    fi
done

# Verificar resultado do build
wait $BUILD_PID
BUILD_EXIT_CODE=$?

if [ $BUILD_EXIT_CODE -ne 0 ]; then
    echo "❌ Build do frontend falhou com código $BUILD_EXIT_CODE"
    echo "   Verifique os logs: tail -f $LOGS/frontend_build.log"
    echo ""
    echo "   Últimas linhas do log:"
    tail -20 "$LOGS/frontend_build.log"
    exit 1
fi

if [ ! -d "$BASE/.next" ]; then
    echo "❌ Build não gerou pasta .next, verifique os logs"
    exit 1
fi

echo "✅ Build do frontend concluído"

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

