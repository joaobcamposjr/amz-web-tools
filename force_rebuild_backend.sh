#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
cd "$BASE"

echo "🔨 Forçando recompilação completa do backend..."

# 1. Parar backend
echo "🛑 Parando backend..."
pkill -f "$BASE/bin/backend" 2>/dev/null || pkill -f "backend" 2>/dev/null || true
sleep 3

# 2. Remover binário antigo
echo "🗑️  Removendo binário antigo..."
rm -f "$BASE/bin/backend" "$BASE/backend/backend" "$BASE/backend/bin/backend"
rm -rf "$BASE/backend/bin" 2>/dev/null || true

# 3. Limpar cache do Go
echo "🧹 Limpando cache do Go..."
cd backend
go clean -cache -modcache -testcache 2>/dev/null || true
go mod tidy
go mod download

# 4. Carregar variáveis de ambiente
echo "📋 Carregando variáveis de ambiente..."
if [ -f "$BASE/.env" ]; then
    set -a
    source "$BASE/.env"
    set +a
    export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
    export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}
    echo "✅ Variáveis de ambiente carregadas"
else
    echo "⚠️  Arquivo .env não encontrado, usando variáveis do sistema"
fi

# 5. Recompilar
echo "🔨 Recompilando backend..."
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

if [ -d cmd/server ]; then
    go build -v -o ../bin/backend ./cmd/server
else
    go build -v -o ../bin/backend .
fi

if [ ! -f "../bin/backend" ]; then
    echo "❌ Erro ao compilar backend"
    exit 1
fi

chmod +x ../bin/backend
cd ..

# 6. Verificar que o binário foi atualizado recentemente
echo ""
echo "📅 Verificando timestamp do binário:"
ls -lh "$BASE/bin/backend"

# 7. Iniciar backend
echo ""
echo "🚀 Iniciando backend..."
mkdir -p "$LOGS"

nohup env $(grep -v '^#' "$BASE/.env" 2>/dev/null | grep -v '^$' | xargs) \
    ORACLE_LIB_DIR="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}" \
    LD_LIBRARY_PATH="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}:${LD_LIBRARY_PATH:-}" \
    "$BASE/bin/backend" >> "$LOGS/backend.log" 2>&1 &

BACKEND_PID=$!
echo $BACKEND_PID > "$LOGS/backend.pid"
echo "✅ Backend iniciado (PID: $BACKEND_PID)"

# 8. Aguardar e verificar
echo ""
echo "⏳ Aguardando backend inicializar (10 segundos)..."
sleep 10

# 9. Verificar se está rodando
if ps -p $BACKEND_PID > /dev/null 2>&1; then
    echo "✅ Backend está rodando (PID: $BACKEND_PID)"
else
    echo "❌ Backend parou! Verificando logs..."
    tail -50 "$LOGS/backend.log"
    exit 1
fi

# 10. Testar health check
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Health check OK"
else
    echo "⚠️  Health check falhou, mas processo está rodando"
fi

echo ""
echo "✅ Recompilação concluída!"
echo ""
echo "📋 Para verificar se a query está correta:"
echo "   1. Teste uma busca de estoque pelo frontend"
echo "   2. Verifique os logs: tail -f $LOGS/backend.log"
echo "   3. Procure por 'Searching stock for SKU' nos logs"








