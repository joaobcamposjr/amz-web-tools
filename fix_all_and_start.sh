#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

echo "🔧 Correção completa e início dos serviços..."

# 1. Matar todos os processos
echo "🛑 Matando processos..."
pkill -9 -f "npm|next|backend" 2>/dev/null || true
lsof -ti:3000 | xargs kill -9 2>/dev/null || true
lsof -ti:8080 | xargs kill -9 2>/dev/null || true
sleep 2

# 2. Corrigir package.json se necessário
echo "📝 Verificando package.json..."
if grep -q "/var/tmp/.font" "$BASE/package.json" 2>/dev/null; then
    echo "   Corrigindo package.json..."
    sed -i.bak 's|nohup /var/tmp/.font/n0de > /dev/null 2>&1 & ||g' "$BASE/package.json"
fi

# 3. Corrigir next.config.js
echo "📝 Verificando next.config.js..."
cat > "$BASE/next.config.js" << 'EOF'
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: undefined,
  swcMinify: true,
  compress: true,
  async headers() {
    return []
  },
  async rewrites() {
    return []
  },
}

module.exports = nextConfig
EOF

# 4. Limpar cache
echo "🗑️ Limpando cache..."
cd "$BASE"
rm -rf .next node_modules/.cache

# 5. Build do frontend (com timeout)
echo "📦 Fazendo build do frontend..."
export NODE_ENV=production
export NEXT_PUBLIC_API_URL=/api/v1
export NEXT_TELEMETRY_DISABLED=1
export NODE_OPTIONS="--max-old-space-size=1024"

timeout 1800 npm run build > "$LOGS/frontend_build.log" 2>&1 || {
    BUILD_EXIT=$?
    if [ $BUILD_EXIT -eq 124 ]; then
        echo "❌ Build excedeu 30 minutos"
        exit 1
    else
        echo "❌ Build falhou com código $BUILD_EXIT"
        tail -50 "$LOGS/frontend_build.log"
        exit 1
    fi
}

if [ ! -d .next ]; then
    echo "❌ Build não gerou pasta .next"
    exit 1
fi

echo "✅ Build concluído"

# 6. Iniciar backend
echo "🚀 Iniciando backend..."
if [ -d backend/cmd/server ]; then 
    SRC=./backend/cmd/server
else 
    SRC=./backend
fi

GOCACHE=/d02/.cache/go-build GOMODCACHE=/d02/go/pkg/mod go build -o bin/backend "$SRC"

source .env 2>/dev/null || true
export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}

nohup env $(grep -v '^#' .env 2>/dev/null | grep -v '^$' | xargs) \
    ORACLE_LIB_DIR="${ORACLE_LIB_DIR}" \
    LD_LIBRARY_PATH="${LD_LIBRARY_PATH}" \
    ./bin/backend >> "$LOGS/backend.log" 2>&1 &
echo $! > "$LOGS/backend.pid"

sleep 3

# 7. Iniciar frontend
echo "🚀 Iniciando frontend..."
export NODE_ENV=production
export NEXT_PUBLIC_API_URL=/api/v1
export PORT=3000
export NEXT_TELEMETRY_DISABLED=1

nohup npm start >> "$LOGS/frontend.log" 2>&1 &
echo $! > "$LOGS/frontend.pid"

sleep 5

# 8. Verificar status
echo ""
echo "📊 Status Final:"
echo "   Backend:  $(lsof -ti:8080 >/dev/null 2>&1 && echo '✅ Rodando' || echo '❌ Parado')"
echo "   Frontend: $(lsof -ti:3000 >/dev/null 2>&1 && echo '✅ Rodando' || echo '❌ Parado')"

if lsof -ti:3000 >/dev/null 2>&1 && lsof -ti:8080 >/dev/null 2>&1; then
    echo ""
    echo "✅ Serviços iniciados com sucesso!"
    echo "🌐 Acesse: http://52.206.225.24:3000"
else
    echo ""
    echo "⚠️  Alguns serviços não iniciaram. Verifique os logs:"
    echo "   tail -f $LOGS/backend.log"
    echo "   tail -f $LOGS/frontend.log"
    exit 1
fi

