#!/bin/bash
set -euo pipefail

BASE="/Users/user/GItProjects/amz-web-tools"
LOGS="$BASE/logs"
mkdir -p "$LOGS"

echo "🚀 AMZ Web Tools - macOS Development"
echo "===================================="

# Mata processos existentes
echo "🔄 Parando processos existentes..."
pkill -f "go run main.go" 2>/dev/null || true
pkill -f "next dev" 2>/dev/null || true
pkill -f "node.*next" 2>/dev/null || true
sleep 2

# Limpa cache
echo "🗑️ Limpando cache..."
if [ -d .next ]; then
  rm -rf .next
fi

# Backend
echo "🔧 Iniciando Backend (Go)..."
cd "$BASE/backend"
nohup go run main.go > "$LOGS/backend.log" 2>&1 & echo $! > "$LOGS/backend.pid"
sleep 3

# Verifica se backend subiu
if curl -s http://localhost:8080/health > /dev/null; then
  echo "✅ Backend rodando na porta 8080"
else
  echo "❌ Backend falhou ao iniciar"
  exit 1
fi

# Frontend
echo "🔧 Iniciando Frontend (Next.js)..."
cd "$BASE"
export NODE_ENV=development
export NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
export PORT=3000

nohup npm run dev > "$LOGS/frontend.log" 2>&1 & echo $! > "$LOGS/frontend.pid"
sleep 5

# Verifica se frontend subiu
if curl -s http://localhost:3000 > /dev/null; then
  echo "✅ Frontend rodando na porta 3000"
else
  echo "❌ Frontend falhou ao iniciar"
  exit 1
fi

echo ""
echo "🎉 SISTEMA INICIADO COM SUCESSO!"
echo "=================================="
echo "🌐 Frontend: http://localhost:3000"
echo "🔧 Backend:  http://localhost:8080"
echo "📊 Health:   http://localhost:8080/health"
echo ""
echo "📝 Logs:"
echo "   Backend:  tail -f $LOGS/backend.log"
echo "   Frontend: tail -f $LOGS/frontend.log"
echo ""
echo "🛑 Para parar:"
echo "   pkill -f 'go run main.go'"
echo "   pkill -f 'next dev'"
echo ""
