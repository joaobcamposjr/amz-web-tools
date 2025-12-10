# 🔧 GUIA: Atualizar SELECT no Servidor

## ⚠️ IMPORTANTE
Execute estas mudanças **DIRETAMENTE NO SERVIDOR**, não no código local.

## 📋 Arquivos a Alterar

### 1. `backend/internal/services/stock.go` - Linha 115

**MUDAR DE:**
```go
WHERE
    e.cod_empresa IN (1,3,17,31,34,35,40,41,43,144,45,47,48,140,147)
    AND e.cod_item = :1
```

**PARA:**
```go
WHERE
    e.cod_empresa IN (17,34,140,144,147)
    AND e.cod_fornecedor IN (7,9,1,13,16,17)
    AND e.cod_item = :1
```

---

### 2. `backend/internal/services/integration_service.go` - 3 lugares

#### A) Linha 1593 (Query Principal)

**MUDAR DE:**
```go
WHERE
    e.cod_empresa IN (1,3,17,31,34,35,40,41,43,144,45,47,48,140,147)
    AND e.cod_item = :1
    AND e.cod_fornecedor = :2
    AND e.cod_empresa = :3
```

**PARA:**
```go
WHERE
    e.cod_empresa IN (17,34,140,144,147)
    AND e.cod_fornecedor IN (7,9,1,13,16,17)
    AND e.cod_item = :1
    AND e.cod_fornecedor = :2
    AND e.cod_empresa = :3
```

⚠️ **NOTA**: Esta query já filtra por `cod_fornecedor = :2` e `cod_empresa = :3` individualmente. A linha `AND e.cod_fornecedor IN (7,9,1,13,16,17)` pode ser redundante se `:2` já restringir o fornecedor. Verifique a lógica.

#### B) Linha 1637 (Fallback Fornecedor 16)

**MUDAR DE:**
```go
WHERE
    e.cod_empresa IN (1,3,17,31,34,35,40,41,43,144,45,47,48,140,147)
    AND e.cod_item = :1
    AND e.cod_fornecedor = 16
    AND e.cod_empresa = :2
```

**PARA:**
```go
WHERE
    e.cod_empresa IN (17,34,140,144,147)
    AND e.cod_item = :1
    AND e.cod_fornecedor = 16
    AND e.cod_empresa = :2
```

#### C) Linha 1668 (Fallback Fornecedor 12)

**MUDAR DE:**
```go
WHERE
    e.cod_empresa IN (1,3,17,31,34,35,40,41,43,144,45,47,48,140,147)
    AND e.cod_item = :1
    AND e.cod_fornecedor = 12
    AND e.cod_empresa = 17
```

**PARA:**
```go
WHERE
    e.cod_empresa IN (17,34,140,144,147)
    AND e.cod_item = :1
    AND e.cod_fornecedor = 12
    AND e.cod_empresa = 17
```

---

### 3. `backend/internal/services/car_plate.go` - Linha 462

**MUDAR DE:**
```go
query := `SELECT COUNT(DISTINCT cod_item) FROM nbs.CRANI_PECAS_ITENS WHERE cod_empresa IN (1,3,17,31,34,35,40,41,43,144,45,47,48,140)`
```

**PARA:**
```go
query := `SELECT COUNT(DISTINCT cod_item) FROM nbs.CRANI_PECAS_ITENS WHERE cod_empresa IN (17,34,140,144,147)`
```

---

## 🔄 Após Fazer as Mudanças

1. **Recompilar o backend:**
   ```bash
   cd /d02/projects/amz-web-tools/backend
   go clean -cache
   go build -o ../bin/backend .
   ```

2. **Parar backend antigo:**
   ```bash
   pkill -f bin/backend
   sleep 2
   ```

3. **Iniciar backend novo:**
   ```bash
   cd /d02/projects/amz-web-tools
   source .env
   export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
   export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}
   
   nohup ./bin/backend >> /d02/logs/backend.log 2>&1 & echo $! > /d02/logs/backend.pid
   ```

4. **Verificar se funcionou:**
   ```bash
   tail -f /d02/logs/backend.log
   # Teste uma busca de estoque pelo frontend
   ```

---

## 📝 Comandos Rápidos no Servidor

```bash
# 1. Editar os arquivos
vi /d02/projects/amz-web-tools/backend/internal/services/stock.go
vi /d02/projects/amz-web-tools/backend/internal/services/integration_service.go
vi /d02/projects/amz-web-tools/backend/internal/services/car_plate.go

# 2. Recompilar e reiniciar
cd /d02/projects/amz-web-tools/backend
go clean -cache && go build -o ../bin/backend .
cd ..
pkill -f bin/backend && sleep 2
source .env
export ORACLE_LIB_DIR=${ORACLE_LIB_DIR:-/opt/oracle/instantclient_21_7}
export LD_LIBRARY_PATH=${ORACLE_LIB_DIR}:${LD_LIBRARY_PATH:-}
nohup ./bin/backend >> /d02/logs/backend.log 2>&1 & echo $! > /d02/logs/backend.pid

# 3. Verificar
tail -20 /d02/logs/backend.log
```





