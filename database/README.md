# Database Scripts - AMZ Web Tools Portal

Este diretório contém os scripts SQL para configuração do banco de dados SQL Server.

## 📁 Arquivos

### `create_tables.sql`
Script principal para criação de todas as tabelas, índices e dados iniciais.

**O que faz:**
- Cria todas as tabelas do sistema
- Adiciona índices para performance
- Cria usuário administrador padrão
- Insere dados de exemplo para testes

### `drop_tables.sql`
Script para remoção completa de todas as tabelas (limpeza total).

**⚠️ ATENÇÃO:** Este script remove TODOS os dados!

## 🚀 Como Usar

### 1. Criar Tabelas (Primeira vez)
```sql
-- Execute no SQL Server Management Studio ou Azure Data Studio
-- Conectado ao servidor: 54.204.42.134:1433
-- Database: portal

-- Copie e cole o conteúdo do arquivo create_tables.sql
```

### 2. Remover Tabelas (Se necessário)
```sql
-- Execute o conteúdo do arquivo drop_tables.sql
-- ⚠️ CUIDADO: Isso apaga TODOS os dados!
```

## 📊 Estrutura das Tabelas

### `users`
- **id:** UUID (chave primária)
- **email:** Email único do usuário
- **password_hash:** Senha criptografada (bcrypt)
- **name:** Nome do usuário
- **department:** Departamento
- **role:** Função (admin, manager, user)
- **created_at/updated_at:** Timestamps

### `user_sessions`
- **id:** UUID (chave primária)
- **user_id:** FK para users
- **token:** Token JWT
- **expires_at:** Data de expiração
- **created_at:** Timestamp

### `plate_cache`
- **id:** UUID (chave primária)
- **plate:** Placa do veículo (única)
- **data:** JSON com dados da API
- **created_at:** Data de criação
- **expires_at:** Data de expiração do cache

### `depara_products`
- **id:** UUID (chave primária)
- **product_code:** Código do produto
- **name:** Nome do produto
- **description:** Descrição
- **category:** Categoria
- **created_at/updated_at:** Timestamps

### `stock_items`
- **id:** UUID (chave primária)
- **brand:** Marca
- **sku:** SKU do produto
- **quantity:** Quantidade em estoque
- **location:** Localização no estoque
- **updated_at:** Última atualização

### `integration_logs`
- **id:** UUID (chave primária)
- **user_id:** FK para users
- **process_type:** Tipo de processo
- **status:** Status da execução
- **data:** Dados do processo
- **created_at:** Timestamp

### `import_logs`
- **id:** UUID (chave primária)
- **user_id:** FK para users
- **file_name:** Nome do arquivo
- **status:** Status da importação
- **processed_records:** Registros processados
- **total_records:** Total de registros
- **error_message:** Mensagem de erro (se houver)
- **created_at/completed_at:** Timestamps

## 👤 Usuário Administrador Padrão

Após executar o script de criação:

- **Email:** `admin@amztools.com`
- **Senha:** `password`
- **Role:** `admin`

**⚠️ IMPORTANTE:** Altere a senha padrão após o primeiro login!

## 🔍 Dados de Exemplo

O script cria dados de exemplo para testes:

### DePara Products
- SAMPLE001 - Filtro de Óleo
- SAMPLE002 - Pastilha de Freio  
- SAMPLE003 - Vela de Ignição

### Stock Items
- MANN FILTRO001 - 50 unidades
- BOSCH PASTILHA001 - 25 unidades
- NGK VELA001 - 100 unidades

## 🛠️ Comandos Úteis

### Verificar se as tabelas foram criadas:
```sql
SELECT TABLE_NAME 
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_TYPE = 'BASE TABLE'
ORDER BY TABLE_NAME;
```

### Verificar usuários criados:
```sql
SELECT id, email, name, role, created_at 
FROM users;
```

### Limpar cache de placas expirado:
```sql
DELETE FROM plate_cache 
WHERE expires_at < GETDATE();
```

### Verificar índices criados:
```sql
SELECT i.name AS IndexName, t.name AS TableName
FROM sys.indexes i
INNER JOIN sys.tables t ON i.object_id = t.object_id
WHERE i.name LIKE 'IX_%'
ORDER BY t.name, i.name;
```

