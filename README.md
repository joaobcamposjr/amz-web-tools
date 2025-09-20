# AMZ Web Tools Portal

Portal de autopeças com sistema de login e módulos específicos para diferentes funcionalidades.

## 🏗️ Arquitetura

- **Backend:** Go (Gin Framework) + SQL Server
- **Frontend:** Next.js + React + TypeScript + Tailwind CSS
- **Banco de Dados:** SQL Server
- **Autenticação:** JWT

## 📋 Módulos

### 🔐 Autenticação
- Login/Registro de usuários
- Controle de acesso por roles (admin, manager, user)
- JWT tokens com refresh

### 👤 Perfil
- Edição de dados pessoais
- Alteração de senha
- Visualização de permissões

### 🚗 Car Plate
- Consulta de placas com cache
- Integração com API externa
- Cache no banco de dados

### 🔗 Integration
- Execução de processos externos
- Status em tempo real
- Log de execuções

### 📄 Import XML
- Upload de arquivos XML
- Processamento assíncrono
- Relatórios de importação

### 📊 DePara
- CRUD completo de produtos
- Busca por código
- Histórico de alterações

### 📦 Stock
- Consulta de estoque por marca + SKU
- Informações de disponibilidade
- Relatórios

## 🚀 Como Executar Localmente

### Pré-requisitos
- Go 1.21+
- Node.js 18+
- SQL Server

### 1. Configurar Variáveis de Ambiente
```bash
cp env.example .env
# Editar .env com suas credenciais
```

### 2. Backend (Go)
```bash
cd backend
go mod tidy
go run main.go
```

### 3. Frontend (Next.js)
```bash
npm install
npm run dev
```

## 🗄️ Estrutura do Banco

O sistema criará automaticamente as seguintes tabelas:
- `users` - Usuários do sistema
- `user_sessions` - Sessões ativas
- `plate_cache` - Cache de consultas de placas
- `depara_products` - Produtos DePara
- `stock_items` - Itens de estoque
- `integration_logs` - Logs de integração
- `import_logs` - Logs de importação XML

## 🔧 Configuração

### Variáveis de Ambiente

```env
# Database
DB_HOST=54.204.42.134
DB_PORT=1433
DB_USER=sa
DB_PASSWORD=321@Mudar@7089341@
DB_NAME=portal

# Server
SERVER_PORT=8080
JWT_SECRET=your-jwt-secret-key-here
JWT_EXPIRE_HOURS=24

# API
PLATE_API_URL=https://api.exemplo.com/placa
PLATE_API_KEY=your-api-key-here
```

## 📱 Endpoints da API

### Autenticação
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/register` - Registro

### Perfil (Protegido)
- `GET /api/v1/profile` - Obter perfil
- `PUT /api/v1/profile` - Atualizar perfil
- `PUT /api/v1/profile/password` - Alterar senha

### Car Plate (Protegido)
- `GET /api/v1/car-plate/:plate` - Consultar placa

### Integration (Protegido)
- `POST /api/v1/integration/execute` - Executar integração
- `GET /api/v1/integration/status/:id` - Status da integração

### DePara (Protegido)
- `GET /api/v1/depara` - Listar produtos
- `POST /api/v1/depara` - Criar produto
- `GET /api/v1/depara/:id` - Obter produto
- `PUT /api/v1/depara/:id` - Atualizar produto
- `DELETE /api/v1/depara/:id` - Deletar produto

### Stock (Protegido)
- `GET /api/v1/stock?brand=X&sku=Y` - Consultar estoque

## 🚀 Deploy

### Desenvolvimento
- Branch: `develop`
- Porta: 3001 (frontend), 8080 (backend)

### Produção
- Branch: `main`
- Porta: 3000 (frontend), 8080 (backend)

## 📝 TODO

- [ ] Implementar serviços de Car Plate com cache
- [ ] Implementar sistema de Integration
- [ ] Implementar upload e processamento de XML
- [ ] Implementar CRUD completo do DePara
- [ ] Implementar consulta de estoque
- [ ] Configurar CI/CD com GitHub Actions
- [ ] Implementar testes automatizados
- [ ] Documentação completa da API
