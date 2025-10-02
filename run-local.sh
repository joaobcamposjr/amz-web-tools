#!/bin/bash

# AMZ Web Tools - Local Development Script
# Este script roda o backend e frontend diretamente no servidor

set -e

echo "🚀 AMZ Web Tools - Iniciando em modo local..."

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para log
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Função para verificar se um processo está rodando
is_running() {
    local port=$1
    if lsof -i :$port >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Função para matar processo na porta
kill_port() {
    local port=$1
    if is_running $port; then
        log "Matando processo na porta $port..."
        lsof -ti :$port | xargs kill -9 2>/dev/null || true
        sleep 2
    fi
}

# Função para iniciar backend
start_backend() {
    log "🔧 Iniciando Backend (Go)..."
    
    cd backend
    
    # Verifica se o .env existe
    if [ ! -f "../.env" ]; then
        error "Arquivo .env não encontrado! Crie o arquivo .env primeiro."
        exit 1
    fi
    
    # Carrega variáveis do .env
    export $(cat ../.env | grep -v '^#' | xargs)
    
    # Mata processo na porta 8080 se existir
    kill_port 8080
    
    # Compila o backend
    log "Compilando backend..."
    go build -o amz-web-tools
    
    if [ $? -ne 0 ]; then
        error "Erro ao compilar backend!"
        exit 1
    fi
    
    # Inicia o backend em background
    log "Iniciando backend na porta 8080..."
    nohup ./amz-web-tools > ../logs/backend.log 2>&1 &
    BACKEND_PID=$!
    
    # Salva PID para poder parar depois
    echo $BACKEND_PID > ../logs/backend.pid
    
    # Aguarda backend iniciar
    sleep 5
    
    if is_running 8080; then
        log "✅ Backend iniciado com sucesso! (PID: $BACKEND_PID)"
    else
        error "❌ Falha ao iniciar backend!"
        cat ../logs/backend.log
        exit 1
    fi
    
    cd ..
}

# Função para iniciar frontend
start_frontend() {
    log "🎨 Iniciando Frontend (Next.js)..."
    
    cd frontend
    
    # Verifica se node_modules existe
    if [ ! -d "node_modules" ]; then
        log "Instalando dependências do frontend..."
        npm install
    fi
    
    # Mata processo na porta 3000 se existir
    kill_port 3000
    
    # Configura variáveis de ambiente
    export NODE_ENV=production
    export NEXT_PUBLIC_API_URL=http://localhost:8080
    
    # Inicia o frontend em background
    log "Iniciando frontend na porta 3000..."
    nohup npm start > ../logs/frontend.log 2>&1 &
    FRONTEND_PID=$!
    
    # Salva PID para poder parar depois
    echo $FRONTEND_PID > ../logs/frontend.pid
    
    # Aguarda frontend iniciar
    sleep 10
    
    if is_running 3000; then
        log "✅ Frontend iniciado com sucesso! (PID: $FRONTEND_PID)"
    else
        error "❌ Falha ao iniciar frontend!"
        cat ../logs/frontend.log
        exit 1
    fi
    
    cd ..
}

# Função para parar tudo
stop_all() {
    log "🛑 Parando todos os serviços..."
    
    # Para backend
    if [ -f "logs/backend.pid" ]; then
        BACKEND_PID=$(cat logs/backend.pid)
        if kill -0 $BACKEND_PID 2>/dev/null; then
            log "Parando backend (PID: $BACKEND_PID)..."
            kill $BACKEND_PID
        fi
        rm -f logs/backend.pid
    fi
    
    # Para frontend
    if [ -f "logs/frontend.pid" ]; then
        FRONTEND_PID=$(cat logs/frontend.pid)
        if kill -0 $FRONTEND_PID 2>/dev/null; then
            log "Parando frontend (PID: $FRONTEND_PID)..."
            kill $FRONTEND_PID
        fi
        rm -f logs/frontend.pid
    fi
    
    # Mata processos nas portas
    kill_port 8080
    kill_port 3000
    
    log "✅ Todos os serviços parados!"
}

# Função para mostrar status
show_status() {
    log "📊 Status dos serviços:"
    
    if is_running 8080; then
        echo -e "  ${GREEN}✅ Backend:${NC} Rodando na porta 8080"
    else
        echo -e "  ${RED}❌ Backend:${NC} Não está rodando"
    fi
    
    if is_running 3000; then
        echo -e "  ${GREEN}✅ Frontend:${NC} Rodando na porta 3000"
    else
        echo -e "  ${RED}❌ Frontend:${NC} Não está rodando"
    fi
}

# Função para mostrar logs
show_logs() {
    local service=$1
    
    case $service in
        "backend")
            if [ -f "logs/backend.log" ]; then
                log "📋 Logs do Backend:"
                tail -f logs/backend.log
            else
                error "Arquivo de log do backend não encontrado!"
            fi
            ;;
        "frontend")
            if [ -f "logs/frontend.log" ]; then
                log "📋 Logs do Frontend:"
                tail -f logs/frontend.log
            else
                error "Arquivo de log do frontend não encontrado!"
            fi
            ;;
        *)
            error "Uso: $0 logs [backend|frontend]"
            ;;
    esac
}

# Função para mostrar ajuda
show_help() {
    echo "🚀 AMZ Web Tools - Script Local"
    echo ""
    echo "Uso: $0 [comando]"
    echo ""
    echo "Comandos:"
    echo "  start     - Inicia backend e frontend"
    echo "  stop      - Para todos os serviços"
    echo "  restart   - Reinicia todos os serviços"
    echo "  status    - Mostra status dos serviços"
    echo "  logs      - Mostra logs (backend|frontend)"
    echo "  help      - Mostra esta ajuda"
    echo ""
    echo "Exemplos:"
    echo "  $0 start"
    echo "  $0 stop"
    echo "  $0 logs backend"
    echo "  $0 status"
}

# Cria diretório de logs
mkdir -p logs

# Verifica argumentos
case "${1:-start}" in
    "start")
        log "🚀 Iniciando AMZ Web Tools em modo local..."
        start_backend
        start_frontend
        log ""
        log "🎉 AMZ Web Tools iniciado com sucesso!"
        log "📱 Frontend: http://localhost:3000"
        log "🔧 Backend: http://localhost:8080"
        log ""
        log "Para ver logs: $0 logs [backend|frontend]"
        log "Para parar: $0 stop"
        ;;
    "stop")
        stop_all
        ;;
    "restart")
        log "🔄 Reiniciando AMZ Web Tools..."
        stop_all
        sleep 3
        start_backend
        start_frontend
        log "✅ Reiniciado com sucesso!"
        ;;
    "status")
        show_status
        ;;
    "logs")
        show_logs $2
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        error "Comando desconhecido: $1"
        show_help
        exit 1
        ;;
esac






