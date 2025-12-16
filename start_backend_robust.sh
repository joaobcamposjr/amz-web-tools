#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="$BASE/logs"
PID_FILE="$LOGS/backend.pid"

mkdir -p "$LOGS"

echo "🚀 Iniciando Backend (modo robusto)..."

# Matar processos anteriores
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    if ps -p "$OLD_PID" > /dev/null 2>&1; then
        echo "🔄 Matando processo anterior (PID: $OLD_PID)..."
        kill "$OLD_PID" 2>/dev/null || true
        sleep 2
        kill -9 "$OLD_PID" 2>/dev/null || true
    fi
fi

pkill -f "bin/backend" 2>/dev/null || true
sleep 2

cd "$BASE"

# Carregar variáveis de ambiente
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

export SERVER_HOST=${SERVER_HOST:-0.0.0.0}
export SERVER_PORT=${SERVER_PORT:-8080}

# Iniciar backend em background com nohup
echo "▶️  Iniciando backend na porta $SERVER_PORT..."
nohup ./bin/backend >> "$LOGS/backend.log" 2>&1 &
BACKEND_PID=$!

echo "$BACKEND_PID" > "$PID_FILE"
echo "✅ Backend iniciado (PID: $BACKEND_PID)"

# Aguardar até 30 segundos para verificar se iniciou corretamente
echo "⏳ Aguardando inicialização..."
for i in {1..30}; do
    if ss -tlnp | grep -q ":$SERVER_PORT "; then
        echo "✅ Backend está escutando na porta $SERVER_PORT!"
        curl -s http://localhost:$SERVER_PORT/api/v1/health > /dev/null 2>&1 && echo "✅ Health check OK" || echo "⚠️  Health check falhou, mas porta está aberta"
        exit 0
    fi
    sleep 1
    # Verificar se processo ainda está vivo
    if ! ps -p "$BACKEND_PID" > /dev/null 2>&1; then
        echo "❌ Backend morreu durante inicialização!"
        echo "Últimas linhas do log:"
        tail -20 "$LOGS/backend.log"
        exit 1
    fi
done

echo "⚠️  Backend não está escutando após 30 segundos, mas processo está vivo"
echo "Verificando logs..."
tail -30 "$LOGS/backend.log"
exit 1

