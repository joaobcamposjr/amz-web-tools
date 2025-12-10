# 🔄 Como Recompilar Após Alterar Código Go

## ⚠️ IMPORTANTE
**Sempre que você alterar um arquivo `.go`, você DEVE:**
1. Recompilar o binário (`go build`)
2. Parar o backend antigo
3. Reiniciar o backend com o novo binário

## 📋 Processo Completo

### 1. Após alterar um arquivo `.go` (ex: `stock.go`)

```bash
cd /d02/projects/amz-web-tools

# 1. Parar o backend atual
pkill -f bin/backend || pkill -f backend || true
sleep 2

# 2. Recompilar o backend
cd backend
go build -o ../bin/backend .

# 3. Verificar se compilou com sucesso
if [ -f "../bin/backend" ]; then
    echo "✅ Compilação OK"
else
    echo "❌ Erro na compilação!"
    exit 1
fi

# 4. Voltar para raiz e carregar .env
cd ..
source .env

# 5. Configurar Oracle Client
export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}

# 6. Iniciar backend com novo binário
nohup ./bin/backend >> /d02/logs/backend.log 2>&1 & echo $! > /d02/logs/backend.pid

# 7. Verificar se iniciou
sleep 3
if ps -p $(cat /d02/logs/backend.pid) > /dev/null 2>&1; then
    echo "✅ Backend iniciado (PID: $(cat /d02/logs/backend.pid))"
else
    echo "❌ Backend não iniciou! Verifique os logs:"
    tail -20 /d02/logs/backend.log
fi
```

---

## 🚀 Script Rápido (Copiar e Colar)

```bash
cd /d02/projects/amz-web-tools && \
pkill -f bin/backend 2>/dev/null || pkill -f backend 2>/dev/null || true && \
sleep 2 && \
cd backend && \
go build -o ../bin/backend . && \
cd .. && \
source .env && \
export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7} && \
export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-} && \
nohup ./bin/backend >> /d02/logs/backend.log 2>&1 & echo $! > /d02/logs/backend.pid && \
echo "✅ Backend recompilado e reiniciado! (PID: $(cat /d02/logs/backend.pid))"
```

---

## 📝 Explicação dos Comandos

### `go build`
- **O que faz**: Compila o código Go e gera o binário executável
- **Quando usar**: Sempre que alterar qualquer arquivo `.go`
- **Não precisa de**: `go mod tidy` (a menos que tenha adicionado novas dependências)

### `go mod tidy`
- **O que faz**: Limpa e organiza as dependências do `go.mod`
- **Quando usar**: Apenas se você adicionou ou removeu pacotes externos
- **Não precisa de**: Apenas para alterar código existente

### `go clean -cache`
- **O que faz**: Limpa o cache de compilação do Go
- **Quando usar**: Se você tiver problemas de compilação ou cache antigo
- **Opcional**: Geralmente não é necessário

---

## ⚡ Diferenças

### Apenas Alterou Código (não adicionou dependências):
```bash
cd backend
go build -o ../bin/backend .
```
✅ **Suficiente!** Não precisa de `go mod tidy` ou `go clean`

### Adicionou Novas Dependências (import novo):
```bash
cd backend
go mod tidy        # Atualiza go.mod e go.sum
go build -o ../bin/backend .
```

### Problemas de Compilação ou Cache:
```bash
cd backend
go clean -cache    # Limpa cache
go mod tidy        # Organiza dependências
go build -o ../bin/backend .
```

---

## 🔍 Verificar se Usou o Código Novo

Após recompilar e reiniciar, verifique os logs:

```bash
# Ver última inicialização
tail -30 /d02/logs/backend.log

# Testar funcionalidade
# Ex: Fazer uma busca de estoque e ver se os filtros estão aplicados
```

Se ainda estiver mostrando resultados antigos, o backend pode estar usando o binário antigo. Verifique:

```bash
# Ver quando o binário foi compilado
ls -lh /d02/projects/amz-web-tools/bin/backend

# Ver se o processo está usando o binário correto
ps aux | grep backend | grep -v grep
```

---

## ❌ Erro Comum

**❌ ERRADO:**
```bash
# Apenas reiniciar sem recompilar
pkill -f backend
./bin/backend  # Usa binário antigo!
```

**✅ CORRETO:**
```bash
# Recompilar E depois reiniciar
pkill -f backend
cd backend && go build -o ../bin/backend .
cd .. && ./bin/backend  # Usa binário novo!
```

---

## 🎯 Resumo

**SIM, você precisa recompilar!** O Go é uma linguagem compilada, então:
- Alterações no código `.go` → `go build` → Novo binário
- Sem recompilar → Backend usa binário antigo → Mudanças não aparecem





