#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

echo "🚀 Iniciando build do frontend em background..."

cd "$BASE"

# Limpar cache
echo "🗑️ Limpando cache..."
if [ -d .next ]; then
    sudo rm -rf .next 2>/dev/null || rm -rf .next
fi

if [ -d node_modules/.cache ]; then
    rm -rf node_modules/.cache
fi

# Corrigir permissões
sudo chown -R $(whoami):$(whoami) . 2>/dev/null || true

# Build em background
export NODE_ENV=production
export NEXT_PUBLIC_API_URL=/api/v1
export PORT=3000
export NEXT_TELEMETRY_DISABLED=1

echo "📦 Iniciando build (pode levar alguns minutos)..."
echo "   📄 Acompanhe os logs em tempo real com:"
echo "   tail -f $LOGS/frontend_build.log"
echo ""
echo "   🔍 Ou verifique o progresso com:"
echo "   ps aux | grep 'npm run build'"
echo ""

# Iniciar build em background
nohup npm run build > "$LOGS/frontend_build.log" 2>&1 &
BUILD_PID=$!

echo "✅ Build iniciado em background (PID: $BUILD_PID)"
echo "   O terminal não será bloqueado!"
echo ""
echo "📊 Para verificar o status:"
echo "   # Ver se ainda está rodando:"
echo "   ps -p $BUILD_PID"
echo ""
echo "   # Ver últimos logs:"
echo "   tail -f $LOGS/frontend_build.log"
echo ""
echo "   # Quando terminar, iniciar frontend com:"
echo "   nohup npm start >> $LOGS/frontend.log 2>&1 &"
echo "   echo \$! > $LOGS/frontend.pid"

