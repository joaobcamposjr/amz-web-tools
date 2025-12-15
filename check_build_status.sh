#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
BUILD_LOG="$LOGS/frontend_build.log"

echo "🔍 Verificando status do build..."

# Verificar se há processo de build rodando
BUILD_PIDS=$(pgrep -f "npm run build" || true)
if [ -n "$BUILD_PIDS" ]; then
    echo "⏳ Build ainda em andamento (PIDs: $BUILD_PIDS)"
    echo ""
    echo "📄 Últimas 20 linhas do log:"
    echo "----------------------------------------"
    tail -20 "$BUILD_LOG" 2>/dev/null || echo "Log ainda não existe"
    echo "----------------------------------------"
    echo ""
    echo "Para acompanhar em tempo real:"
    echo "   tail -f $BUILD_LOG"
else
    echo "✅ Nenhum build em andamento"
    
    # Verificar se a pasta .next existe (build concluído)
    if [ -d "$BASE/.next" ]; then
        echo "✅ Build parece ter concluído (.next existe)"
        echo ""
        echo "📊 Últimas linhas do log:"
        echo "----------------------------------------"
        tail -30 "$BUILD_LOG" 2>/dev/null || echo "Log não encontrado"
        echo "----------------------------------------"
    else
        echo "⚠️  Pasta .next não encontrada - build pode ter falhado ou não iniciado"
        echo ""
        if [ -f "$BUILD_LOG" ]; then
            echo "📄 Últimas linhas do log:"
            echo "----------------------------------------"
            tail -50 "$BUILD_LOG"
            echo "----------------------------------------"
        fi
    fi
fi

