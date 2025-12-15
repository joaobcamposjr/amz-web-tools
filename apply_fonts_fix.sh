#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"

echo "🔧 Aplicando correção de fontes do Google..."

cd "$BASE"

echo "📥 Fazendo git pull..."
git pull origin main

echo "📦 Verificando dependências..."
npm install

echo "🗑️ Limpando cache Next.js..."
if [ -d .next ]; then
    sudo rm -rf .next 2>/dev/null || rm -rf .next
fi

echo "🔨 Fazendo build do frontend..."
export NODE_ENV=production
export NEXT_PUBLIC_API_URL=/api/v1
export PORT=3000
npm run build

echo "✅ Correção aplicada com sucesso!"
echo ""
echo "🔄 Para reiniciar os serviços, execute:"
echo "   ./start_services.sh"

