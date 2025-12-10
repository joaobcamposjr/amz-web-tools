#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
cd "$BASE"

echo "🔍 Diagnóstico das conexões de banco de dados..."

# 1. Verificar se o arquivo .env existe
if [ ! -f ".env" ]; then
    echo "⚠️  Arquivo .env não encontrado!"
    echo "📋 Restaurando .env a partir de env.server..."
    cp env.server .env
    chmod 600 .env
    echo "✅ Arquivo .env restaurado"
else
    echo "✅ Arquivo .env existe"
fi

# 2. Verificar conteúdo do .env
echo ""
echo "🔍 Verificando configurações do .env:"
echo "  - ORACLE_HOST: $(grep ORACLE_HOST .env | cut -d'=' -f2 || echo 'NÃO CONFIGURADO')"
echo "  - PG_HOST: $(grep PG_HOST .env | cut -d'=' -f2 || echo 'NÃO CONFIGURADO')"
echo "  - DB_HOST: $(grep DB_HOST .env | cut -d'=' -f2 || echo 'NÃO CONFIGURADO')"

# 3. Corrigir configuração do PostgreSQL (pode estar na porta errada)
if grep -q "PG_PORT=5432" .env; then
    echo "⚠️  Corrigindo porta do PostgreSQL de 5432 para 5433..."
    sed -i 's/PG_PORT=5432/PG_PORT=5433/g' .env
    echo "✅ Porta do PostgreSQL corrigida para 5433"
fi

# 4. Parar o backend atual
echo ""
echo "🛑 Parando backend atual..."
pkill -f "$BASE/bin/backend" 2>/dev/null || true
pkill -f "backend" 2>/dev/null || true
sleep 2

# 5. Carregar variáveis de ambiente
echo ""
echo "📋 Carregando variáveis de ambiente do .env..."
set -a
source .env
set +a

# 6. Verificar variáveis carregadas
echo ""
echo "🔍 Variáveis de ambiente carregadas:"
echo "  - ORACLE_HOST=${ORACLE_HOST:-NÃO CONFIGURADO}"
echo "  - ORACLE_USER=${ORACLE_USER:-NÃO CONFIGURADO}"
echo "  - PG_HOST=${PG_HOST:-NÃO CONFIGURADO}"
echo "  - PG_PORT=${PG_PORT:-NÃO CONFIGURADO}"
echo "  - DB_HOST=${DB_HOST:-NÃO CONFIGURADO}"

# 7. Configurar Oracle Client
echo ""
echo "🔧 Configurando Oracle Client..."
export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}
echo "  - ORACLE_LIB_DIR=$ORACLE_LIB_DIR"
echo "  - LD_LIBRARY_PATH=$LD_LIBRARY_PATH"

# 8. Verificar se as bibliotecas Oracle existem
if [ -f "$ORACLE_LIB_DIR/libclntsh.so" ] || [ -f "$ORACLE_LIB_DIR/libclntsh.so.*" ]; then
    echo "✅ Biblioteca Oracle encontrada"
else
    echo "⚠️  Biblioteca Oracle NÃO encontrada em $ORACLE_LIB_DIR"
    echo "   Verificando alternativas..."
    for libdir in /opt/oracle/instantclient_* /usr/lib/oracle/*; do
        if [ -d "$libdir" ] && [ -f "$libdir/libclntsh.so"* ] 2>/dev/null; then
            echo "   ✅ Encontrada em: $libdir"
            export ORACLE_LIB_DIR="$libdir"
            export LD_LIBRARY_PATH="$libdir:${LD_LIBRARY_PATH:-}"
            sed -i "s|ORACLE_LIB_DIR=.*|ORACLE_LIB_DIR=$libdir|g" .env
            break
        fi
    done
fi

# 9. Testar conectividade de rede com os bancos
echo ""
echo "🌐 Testando conectividade de rede..."

# Testar Oracle
if command -v nc &> /dev/null; then
    echo -n "  - Oracle ($ORACLE_HOST:$ORACLE_PORT): "
    if timeout 5 nc -z "$ORACLE_HOST" "${ORACLE_PORT:-1521}" 2>/dev/null; then
        echo "✅ ACESSÍVEL"
    else
        echo "❌ INACESSÍVEL (verificar firewall/rede)"
    fi
else
    echo "  ⚠️  nc (netcat) não encontrado, pulando teste de conectividade"
fi

# Testar PostgreSQL
if command -v nc &> /dev/null; then
    echo -n "  - PostgreSQL ($PG_HOST:$PG_PORT): "
    if timeout 5 nc -z "$PG_HOST" "${PG_PORT:-5433}" 2>/dev/null; then
        echo "✅ ACESSÍVEL"
    else
        echo "❌ INACESSÍVEL (verificar firewall/rede)"
    fi
fi

# Testar SQL Server
if command -v nc &> /dev/null; then
    echo -n "  - SQL Server ($DB_HOST:$DB_PORT): "
    if timeout 5 nc -z "$DB_HOST" "${DB_PORT:-1433}" 2>/dev/null; then
        echo "✅ ACESSÍVEL"
    else
        echo "❌ INACESSÍVEL (verificar firewall/rede)"
    fi
fi

# 10. Recompilar backend (caso necessário)
echo ""
echo "🔨 Recompilando backend..."
cd backend
go mod tidy
go mod download

# Exportar todas as variáveis necessárias antes do build
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

if [ -d cmd/server ]; then
    go build -o ../bin/backend ./cmd/server
else
    go build -o ../bin/backend .
fi

if [ -f "../bin/backend" ]; then
    echo "✅ Backend compilado com sucesso"
    chmod +x ../bin/backend
else
    echo "❌ Erro ao compilar backend"
    exit 1
fi

cd ..

# 11. Iniciar backend com todas as variáveis de ambiente
echo ""
echo "🚀 Iniciando backend com configurações de banco..."

# Garantir que o diretório de logs existe
mkdir -p "$LOGS"

# Iniciar backend em background com todas as variáveis exportadas
nohup env $(grep -v '^#' .env | grep -v '^$' | xargs) \
    ORACLE_LIB_DIR="$ORACLE_LIB_DIR" \
    LD_LIBRARY_PATH="$LD_LIBRARY_PATH" \
    "$BASE/bin/backend" >> "$LOGS/backend.log" 2>&1 &

BACKEND_PID=$!
echo $BACKEND_PID > "$LOGS/backend.pid"
echo "✅ Backend iniciado (PID: $BACKEND_PID)"

# 12. Aguardar backend inicializar
echo ""
echo "⏳ Aguardando backend inicializar..."
sleep 8

# 13. Verificar logs do backend
echo ""
echo "📄 Últimas linhas do log do backend:"
tail -20 "$LOGS/backend.log" || echo "⚠️  Não foi possível ler o log"

# 14. Testar health check
echo ""
echo "🔍 Testando health check do backend..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Backend respondendo corretamente"
else
    echo "❌ Backend não está respondendo"
    echo ""
    echo "📄 Logs completos do backend:"
    tail -50 "$LOGS/backend.log"
    exit 1
fi

# 15. Verificar conexões de banco nos logs
echo ""
echo "🔍 Verificando conexões de banco nos logs..."
if grep -i "oracle.*connected\|oracle.*sucesso\|oracle.*connection.*established" "$LOGS/backend.log" 2>/dev/null | tail -1; then
    echo "✅ Conexão Oracle estabelecida"
else
    echo "⚠️  Não encontrado log de conexão Oracle bem-sucedida"
fi

if grep -i "postgres.*connected\|postgres.*sucesso\|postgres.*connection.*established" "$LOGS/backend.log" 2>/dev/null | tail -1; then
    echo "✅ Conexão PostgreSQL estabelecida"
else
    echo "⚠️  Não encontrado log de conexão PostgreSQL bem-sucedida"
fi

if grep -i "database.*connected\|sql.*server.*connected\|database.*sucesso" "$LOGS/backend.log" 2>/dev/null | tail -1; then
    echo "✅ Conexão SQL Server estabelecida"
else
    echo "⚠️  Não encontrado log de conexão SQL Server bem-sucedida"
fi

echo ""
echo "✅ Processo de correção concluído!"
echo ""
echo "📋 Próximos passos:"
echo "  1. Verifique os logs em: $LOGS/backend.log"
echo "  2. Teste as funcionalidades no frontend"
echo "  3. Se houver problemas, verifique:"
echo "     - Firewall do servidor"
echo "     - Permissões de arquivo (.env deve ser 600)"
echo "     - Configurações de rede (nginx, firewall)"

