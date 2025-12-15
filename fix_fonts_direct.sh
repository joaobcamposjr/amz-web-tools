#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LAYOUT_FILE="$BASE/src/app/layout.tsx"
GLOBALS_FILE="$BASE/src/app/globals.css"

echo "🔧 Corrigindo fontes do Google diretamente no servidor..."

# Verificar se os arquivos existem
if [ ! -f "$LAYOUT_FILE" ]; then
    echo "❌ Erro: $LAYOUT_FILE não encontrado"
    exit 1
fi

if [ ! -f "$GLOBALS_FILE" ]; then
    echo "❌ Erro: $GLOBALS_FILE não encontrado"
    exit 1
fi

# Backup dos arquivos originais
echo "💾 Fazendo backup..."
cp "$LAYOUT_FILE" "${LAYOUT_FILE}.backup.$(date +%Y%m%d_%H%M%S)" 2>/dev/null || true
cp "$GLOBALS_FILE" "${GLOBALS_FILE}.backup.$(date +%Y%m%d_%H%M%S)" 2>/dev/null || true

# 1. Remover import do Google Fonts do layout.tsx
echo "📝 Removendo Google Fonts do layout.tsx..."
sed -i.tmp '/import { Inter } from .next\/font\/google/d' "$LAYOUT_FILE"
sed -i.tmp '/^const inter = Inter({ subsets: .latin. })/d' "$LAYOUT_FILE"
sed -i.tmp 's/${inter.className}//g' "$LAYOUT_FILE"
sed -i.tmp 's/className={`/className="/g' "$LAYOUT_FILE"
sed -i.tmp 's/`} bg-gray-50 antialiased`/" bg-gray-50 antialiased font-sans"/g' "$LAYOUT_FILE"
sed -i.tmp 's/className="bg-gray-50 antialiased`/"bg-gray-50 antialiased font-sans"/g' "$LAYOUT_FILE"
rm -f "${LAYOUT_FILE}.tmp"

# Garantir que o body tenha a classe correta
if ! grep -q 'className="bg-gray-50 antialiased font-sans"' "$LAYOUT_FILE"; then
    sed -i.tmp 's/<body className="[^"]*">/<body className="bg-gray-50 antialiased font-sans">/g' "$LAYOUT_FILE"
    rm -f "${LAYOUT_FILE}.tmp"
fi

# 2. Adicionar fontes do sistema no globals.css
echo "📝 Adicionando fontes do sistema no globals.css..."

# Verificar se já tem font-family no body
if ! grep -q "font-family:" "$GLOBALS_FILE"; then
    # Adicionar font-family após @apply bg-gray-50 text-gray-900;
    sed -i.tmp '/@apply bg-gray-50 text-gray-900;/a\
    font-family: -apple-system, BlinkMacSystemFont, '\''Segoe UI'\'', '\''Roboto'\'', '\''Oxygen'\'', '\''Ubuntu'\'', '\''Cantarell'\'', '\''Fira Sans'\'', '\''Droid Sans'\'', '\''Helvetica Neue'\'', sans-serif;\
    -webkit-font-smoothing: antialiased;\
    -moz-osx-font-smoothing: grayscale;' "$GLOBALS_FILE"
    rm -f "${GLOBALS_FILE}.tmp"
else
    echo "   ⚠️  font-family já existe no globals.css, pulando..."
fi

echo "✅ Arquivos modificados com sucesso!"
echo ""
echo "📋 Verificando alterações..."

# Verificar se o Google Fonts foi removido
if grep -q "next/font/google" "$LAYOUT_FILE"; then
    echo "⚠️  Aviso: Ainda há referências ao Google Fonts em layout.tsx"
else
    echo "✅ Google Fonts removido de layout.tsx"
fi

# Verificar se font-family foi adicionado
if grep -q "font-family:" "$GLOBALS_FILE"; then
    echo "✅ Fontes do sistema adicionadas ao globals.css"
else
    echo "⚠️  Aviso: Fontes do sistema não foram adicionadas ao globals.css"
fi

echo ""
echo "🔄 Limpando cache do Next.js..."
cd "$BASE"
if [ -d .next ]; then
    sudo rm -rf .next 2>/dev/null || rm -rf .next
    echo "✅ Cache limpo"
fi

echo ""
echo "✅ Correção aplicada! Para fazer o build, execute:"
echo "   cd $BASE && npm run build"

