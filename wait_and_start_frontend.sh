#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LOGS="/d02/logs"
BUILD_LOG="$LOGS/frontend_build.log"

echo "⏳ Aguardando conclusão do build..."
echo "   (Você pode cancelar com Ctrl+C a qualquer momento)"
echo ""

MAX_WAIT=1200  # 20 minutos
ELAPSED=0
CHECK_INTERVAL=10

while [ $ELAPSED -lt $MAX_WAIT ]; do
    # Verificar se build ainda está rodando
    if ! pgrep -f "npm run build" >/dev/null 2>&1; then
        # Build terminou, verificar se foi bem-sucedido
        if [ -d "$BASE/.next" ] && [ -f "$BASE/.next/BUILD_ID" ]; then
            echo ""
            echo "✅ Build concluído com sucesso!"
            echo ""
            echo "🚀 Iniciando frontend..."
            
            cd "$BASE"
            export NODE_ENV=production
            export NEXT_PUBLIC_API_URL=/api/v1
            export PORT=3000
            
            # Verificar se frontend já está rodando
            if lsof -ti:3000 >/dev/null 2>&1; then
                echo "⚠️  Porta 3000 já está em uso"
                exit 1
            fi
            
            nohup npm start >> "$LOGS/frontend.log" 2>&1 &
            FRONTEND_PID=$!
            echo $FRONTEND_PID > "$LOGS/frontend.pid"
            
            sleep 5
            
            if kill -0 $FRONTEND_PID 2>/dev/null && lsof -ti:3000 >/dev/null 2>&1; then
                echo "✅ Frontend iniciado com sucesso (PID: $FRONTEND_PID)"
                echo ""
                echo "📊 Status Final:"
                echo "   Backend:  $(lsof -ti:8080 >/dev/null 2>&1 && echo '✅ Rodando' || echo '❌ Parado')"
                echo "   Frontend: ✅ Rodando"
                echo ""
                echo "🌐 Acesse: http://52.206.225.24:3000"
                exit 0
            else
                echo "❌ Frontend falhou ao iniciar"
                echo "   Verifique os logs: tail -f $LOGS/frontend.log"
                exit 1
            fi
        else
            echo ""
            echo "❌ Build falhou ou está incompleto"
            echo ""
            echo "📄 Últimas linhas do log:"
            tail -50 "$BUILD_LOG"
            exit 1
        fi
    fi
    
    # Mostrar progresso a cada minuto
    if [ $((ELAPSED % 60)) -eq 0 ] && [ $ELAPSED -gt 0 ]; then
        echo "   ⏱️  Build ainda em andamento... ${ELAPSED}s decorridos"
        # Mostrar última linha do log
        tail -1 "$BUILD_LOG" 2>/dev/null | sed 's/^/   📝 /' || true
    fi
    
    sleep $CHECK_INTERVAL
    ELAPSED=$((ELAPSED + CHECK_INTERVAL))
done

echo ""
echo "❌ Timeout: Build excedeu 20 minutos"
echo "   Verifique os logs: tail -f $BUILD_LOG"
exit 1

