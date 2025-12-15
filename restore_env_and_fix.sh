#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
cd "$BASE"

echo "🔧 Restaurando .env e corrigindo conexões de banco..."

# 1. Restaurar .env a partir de env.server
if [ ! -f ".env" ]; then
    echo "⚠️  Arquivo .env não encontrado - restaurando..."
    cp env.server .env
    chmod 600 .env
    echo "✅ .env restaurado a partir de env.server"
elif ! grep -q "ORACLE_HOST=164.152.40.38" .env 2>/dev/null; then
    echo "⚠️  Arquivo .env parece estar incorreto - restaurando..."
    cp env.server .env
    chmod 600 .env
    echo "✅ .env restaurado a partir de env.server"
else
    echo "✅ Arquivo .env já existe e parece correto"
fi

# 2. Corrigir porta do PostgreSQL (pode estar na porta errada)
if grep -q "PG_PORT=5432" .env; then
    echo "⚠️  Ajustando porta do PostgreSQL (5432 -> 5433)..."
    sed -i 's/PG_PORT=5432/PG_PORT=5433/g' .env
    echo "✅ Porta do PostgreSQL ajustada para 5433"
fi

# 3. Carregar variáveis de ambiente
echo "📋 Carregando variáveis de ambiente..."
set -a
source .env
set +a

# 4. Verificar configurações críticas
echo ""
echo "🔍 Configurações carregadas:"
echo "  ✅ ORACLE_HOST: ${ORACLE_HOST:-❌ NÃO CONFIGURADO}"
echo "  ✅ PG_HOST: ${PG_HOST:-❌ NÃO CONFIGURADO}"
echo "  ✅ DB_HOST: ${DB_HOST:-❌ NÃO CONFIGURADO}"
echo "  ✅ ORACLE_LIB_DIR: ${ORACLE_LIB_DIR:-❌ NÃO CONFIGURADO}"

# 5. Configurar Oracle Client
export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}

# 6. Parar backend atual
echo ""
echo "🛑 Parando backend atual..."
pkill -f "$BASE/bin/backend" 2>/dev/null || pkill -f "backend" 2>/dev/null || true
sleep 2

# 7. Recompilar backend (se necessário)
echo ""
echo "🔨 Recompilando backend..."
cd backend
go mod tidy > /dev/null 2>&1 || true
go mod download > /dev/null 2>&1 || true

if [ -d cmd/server ]; then
    go build -o ../bin/backend ./cmd/server
else
    go build -o ../bin/backend .
fi

if [ ! -f "../bin/backend" ]; then
    echo "❌ Erro ao compilar backend"
    exit 1
fi

chmod +x ../bin/backend
cd ..
echo "✅ Backend recompilado"

# 8. Iniciar backend com todas as variáveis de ambiente
echo ""
echo "🚀 Iniciando backend com configurações corretas..."
mkdir -p "$LOGS"

# Exportar todas as variáveis do .env antes de iniciar
nohup env $(grep -v '^#' .env | grep -v '^$' | xargs) \
    ORACLE_LIB_DIR="$ORACLE_LIB_DIR" \
    LD_LIBRARY_PATH="$LD_LIBRARY_PATH" \
    "$BASE/bin/backend" >> "$LOGS/backend.log" 2>&1 &

BACKEND_PID=$!
echo $BACKEND_PID > "$LOGS/backend.pid"
echo "✅ Backend iniciado (PID: $BACKEND_PID)"

# 9. Aguardar e verificar
echo ""
echo "⏳ Aguardando backend inicializar (10 segundos)..."
sleep 10

# 10. Verificar status
echo ""
echo "🔍 Verificando status do backend..."

# Verificar se o processo ainda está rodando
if ps -p $BACKEND_PID > /dev/null 2>&1; then
    echo "✅ Backend ainda está rodando (PID: $BACKEND_PID)"
else
    echo "❌ Backend parou! Verificando logs..."
    tail -30 "$LOGS/backend.log"
    exit 1
fi

# Verificar health check
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Health check OK"
else
    echo "❌ Health check falhou"
    echo ""
    echo "📄 Últimos logs do backend:"
    tail -50 "$LOGS/backend.log"
    exit 1
fi

# 11. Verificar conexões nos logs
echo ""
echo "🔍 Verificando conexões de banco nos logs:"
if grep -i "oracle.*connected\|oracle.*connection.*established\|Oracle.*available" "$LOGS/backend.log" 2>/dev/null | tail -1; then
    echo "  ✅ Oracle: Conexão estabelecida"
else
    echo "  ⚠️  Oracle: Não encontrado log de conexão bem-sucedida"
    grep -i "oracle" "$LOGS/backend.log" 2>/dev/null | tail -5 || echo "    (sem logs Oracle)"
fi

if grep -i "postgres.*connected\|postgres.*connection.*established\|PostgreSQL.*available" "$LOGS/backend.log" 2>/dev/null | tail -1; then
    echo "  ✅ PostgreSQL: Conexão estabelecida"
else
    echo "  ⚠️  PostgreSQL: Não encontrado log de conexão bem-sucedida"
    grep -i "postgres" "$LOGS/backend.log" 2>/dev/null | tail -5 || echo "    (sem logs PostgreSQL)"
fi

if grep -i "database.*connected\|sql.*server.*connected\|Database.*connected" "$LOGS/backend.log" 2>/dev/null | tail -1; then
    echo "  ✅ SQL Server: Conexão estabelecida"
else
    echo "  ⚠️  SQL Server: Não encontrado log de conexão bem-sucedida"
    grep -i "database\|sql.*server" "$LOGS/backend.log" 2>/dev/null | tail -5 || echo "    (sem logs SQL Server)"
fi

echo ""
echo "✅ Processo concluído!"
echo ""
echo "📋 Próximos passos:"
echo "  1. Verifique os logs completos: tail -f $LOGS/backend.log"
echo "  2. Teste uma funcionalidade que usa banco de dados"
echo "  3. Se ainda houver problemas:"
echo "     - Verifique firewall: sudo iptables -L -n"
echo "     - Verifique conectividade: nc -zv $ORACLE_HOST $ORACLE_PORT"
echo "     - Verifique nginx: sudo systemctl status nginx"









