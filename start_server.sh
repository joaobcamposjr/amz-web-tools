#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
mkdir -p "$LOGS" /d02/.cache/go-build /d02/go/pkg/mod

echo "🚀 Backend: build"
cd "$BASE"
if [ -d backend/cmd/server ]; then SRC=./backend/cmd/server; else SRC=./backend; fi
GOCACHE=/d02/.cache/go-build GOMODCACHE=/d02/go/pkg/mod go build -o bin/backend "$SRC"

# Carregar variáveis de ambiente do .env se existir
if [ -f "$BASE/.env" ]; then
  echo "📋 Carregando variáveis de ambiente do .env..."
  set -a
  source "$BASE/.env"
  set +a
  
  # Configurar Oracle Client
  export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
  export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}
fi

# Se usa systemd:
if systemctl list-unit-files | grep -q amz-backend.service; then
  echo "🔁 Backend via systemd"
  sudo systemctl restart amz-backend
  sudo systemctl --no-pager status amz-backend | sed -n '1,10p'
else
  echo "▶️  Backend via nohup"
  pkill -f "$BASE/bin/backend" 2>/dev/null || true
  sleep 2
  
  # Iniciar backend com todas as variáveis de ambiente exportadas
  nohup env $(grep -v '^#' "$BASE/.env" 2>/dev/null | grep -v '^$' | xargs) \
      ORACLE_LIB_DIR="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}" \
      LD_LIBRARY_PATH="${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}:${LD_LIBRARY_PATH:-}" \
      "$BASE/bin/backend" >> "$LOGS/backend.log" 2>&1 & echo $! > "$LOGS/backend.pid"
fi

echo "🚀 Frontend: build+start"
cd "$BASE"
export NODE_ENV=production NEXT_PUBLIC_API_URL=/api/v1 PORT=3000

# Mata todos os processos Next.js primeiro
echo "🔄 Parando processos Next.js..."
pkill -f "next" 2>/dev/null || true
pkill -f "node.*next" 2>/dev/null || true
sleep 2

# Remove cache com sudo se necessário
echo "🗑️ Limpando cache..."
if [ -d .next ]; then
  sudo rm -rf .next 2>/dev/null || rm -rf .next
fi

# Corrige permissões dos arquivos Next.js
echo "🔧 Corrigindo permissões..."
sudo chown -R $(whoami):$(whoami) . 2>/dev/null || true
chmod -R 755 . 2>/dev/null || true

npm run build
nohup npm start > "$LOGS/frontend.log" 2>&1 & echo $! > "$LOGS/frontend.pid"

echo "✅ UP. Portas:"
ss -ltnp | egrep ':8080|:3000' || true
