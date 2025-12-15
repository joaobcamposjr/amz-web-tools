#!/bin/bash
set -euo pipefail

BASE="/d02/projects/amz-web-tools"
LAYOUT_FILE="$BASE/src/app/layout.tsx"
GLOBALS_FILE="$BASE/src/app/globals.css"

echo "🔧 Corrigindo fontes do Google no servidor..."

# Backup dos arquivos originais
echo "💾 Fazendo backup dos arquivos..."
cp "$LAYOUT_FILE" "${LAYOUT_FILE}.backup" 2>/dev/null || true
cp "$GLOBALS_FILE" "${GLOBALS_FILE}.backup" 2>/dev/null || true

# Corrigir layout.tsx - remover Google Fonts
echo "📝 Corrigindo layout.tsx..."
cat > "$LAYOUT_FILE" << 'EOF'
import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'AMZ Web Tools Portal',
  description: 'Portal de autopeças com sistema de login e módulos específicos',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="pt-BR">
      <body className="bg-gray-50 antialiased font-sans">
        {children}
      </body>
    </html>
  )
}
EOF

# Corrigir globals.css - adicionar fontes do sistema
echo "📝 Corrigindo globals.css..."
# Ler o arquivo e substituir a seção do body
sed -i.bak 's/body {/body {\n    font-family: -apple-system, BlinkMacSystemFont, '\''Segoe UI'\'', '\''Roboto'\'', '\''Oxygen'\'', '\''Ubuntu'\'', '\''Cantarell'\'', '\''Fira Sans'\'', '\''Droid Sans'\'', '\''Helvetica Neue'\'', sans-serif;\n    -webkit-font-smoothing: antialiased;\n    -moz-osx-font-smoothing: grayscale;/' "$GLOBALS_FILE" 2>/dev/null || {
  # Se sed falhar, usar método alternativo com perl
  perl -i.bak -pe 's/(body \{)/$1\n    font-family: -apple-system, BlinkMacSystemFont, '\''Segoe UI'\'', '\''Roboto'\'', '\''Oxygen'\'', '\''Ubuntu'\'', '\''Cantarell'\'', '\''Fira Sans'\'', '\''Droid Sans'\'', '\''Helvetica Neue'\'', sans-serif;\n    -webkit-font-smoothing: antialiased;\n    -moz-osx-font-smoothing: grayscale;/' "$GLOBALS_FILE" 2>/dev/null || {
    # Se ambos falharem, usar Python
    python3 << PYEOF
with open('$GLOBALS_FILE', 'r') as f:
    content = f.read()

# Verificar se já tem font-family
if 'font-family:' not in content:
    content = content.replace(
        '  body {',
        '''  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
      'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
      sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;'''
    )

with open('$GLOBALS_FILE', 'w') as f:
    f.write(content)
PYEOF
  }
}

# Limpar backups temporários do sed/perl
rm -f "${GLOBALS_FILE}.bak"

echo "✅ Arquivos corrigidos!"
echo ""
echo "🔄 Limpando cache e fazendo build..."
cd "$BASE"

# Limpar cache
if [ -d .next ]; then
    sudo rm -rf .next 2>/dev/null || rm -rf .next
fi

# Build
export NODE_ENV=production
export NEXT_PUBLIC_API_URL=/api/v1
export PORT=3000
npm run build

echo ""
echo "✅ Correção concluída! Backups salvos em:"
echo "   ${LAYOUT_FILE}.backup"
echo "   ${GLOBALS_FILE}.backup"
