# 🚀 AMZ Web Tools - Server Deployment Guide

## 📋 Pré-requisitos do Servidor

- **Sistema**: Amazon Linux 2023
- **Oracle Instant Client**: `/opt/oracle/instantclient_21_7`
- **Variáveis de ambiente**: `/opt/oracle/var.sh`
- **Espaço em disco**: `/d02/` (54GB disponível)

## 🔧 Setup Inicial do Servidor

### 1. Oracle Environment (já configurado)
```bash
export ORACLE_HOME=/opt/oracle/instantclient_21_7
export LD_LIBRARY_PATH="$ORACLE_HOME"
export PATH="$ORACLE_HOME:$PATH"
source /opt/oracle/var.sh
```

### 2. Clone do Projeto
```bash
cd /d02/
git clone <seu-repositorio> amz-web-tools
cd amz-web-tools
```

### 3. Configurar Variáveis de Ambiente
```bash
# Copiar arquivo de servidor (configurado para acesso externo)
cp env.server .env

# Verificar configurações de rede
cat .env | grep -E "(CORS|NEXT_PUBLIC|SERVER_HOST|EXTERNAL)"
```

### 4. Inicializar Sistema
```bash
# Tornar executável
chmod +x init_server.sh stop_server.sh

# Inicializar sistema completo
./init_server.sh
```

## 🎯 Scripts Disponíveis

### `init_server.sh` - Inicialização Completa
- ✅ Configura variáveis de ambiente Oracle
- ✅ Cria diretórios necessários (`/d02/logs`, `/d02/.cache`)
- ✅ Mata todos os processos existentes
- ✅ Limpa cache e build artifacts
- ✅ Corrige permissões
- ✅ Instala dependências (Go + Node.js)
- ✅ Builda aplicações
- ✅ Inicia serviços (Backend + Frontend)
- ✅ Executa health checks
- ✅ Logs detalhados para debug

### `stop_server.sh` - Parar Serviços
- ✅ Para serviços por PID
- ✅ Mata processos por porta
- ✅ Mata processos por nome
- ✅ Limpa PID files

## 📊 Monitoramento

### Logs
```bash
# Backend
tail -f /d02/logs/backend.log

# Frontend  
tail -f /d02/logs/frontend.log

# Logs em tempo real
tail -f /d02/logs/*.log
```

### Health Checks
```bash
# Backend
curl http://52.206.225.24:8080/health

# Frontend
curl http://52.206.225.24:3000

# Status dos processos
lsof -i :8080 -i :3000

# Teste de acesso externo
curl -H "Origin: http://52.206.225.24:3000" http://52.206.225.24:8080/health
```

### Debug Commands
```bash
# Verificar Oracle
echo $ORACLE_HOME
ls -la /opt/oracle/instantclient_21_7/

# Verificar Go
go version
go env GOCACHE GOMODCACHE

# Verificar Node.js
node --version
npm --version

# Verificar processos
ps aux | grep -E '(backend|next)'
```

## 🔄 Comandos de Deploy

### Deploy Inicial
```bash
./init_server.sh
```

### Rebuild Completo
```bash
./stop_server.sh
./init_server.sh
```

### Apenas Restart
```bash
./stop_server.sh
sleep 5
./init_server.sh
```

## 🚨 Troubleshooting

### Problemas Comuns

1. **Porta ocupada**
   ```bash
   lsof -i :8080 -i :3000
   ./stop_server.sh
   ```

2. **Permissões**
   ```bash
   sudo chown -R ec2-user:ec2-user /d02/
   chmod +x init_server.sh stop_server.sh
   ```

3. **Oracle não conecta**
   ```bash
   echo $ORACLE_HOME
   echo $LD_LIBRARY_PATH
   ls -la /opt/oracle/instantclient_21_7/
   ```

4. **Cache corrompido**
   ```bash
   rm -rf /d02/.cache/go-build
   rm -rf /d02/projects/amz-web-tools/.next
   ./init_server.sh
   ```

## 📝 Estrutura de Diretórios

```
/d02/
├── projects/amz-web-tools/     # Código fonte
├── logs/                       # Logs do sistema
├── .cache/go-build/           # Cache do Go
├── go/pkg/mod/                # Módulos do Go
└── pids/                      # PID files dos processos
```

## 🔒 Segurança

- Todos os logs em `/d02/logs/`
- Cache em `/d02/.cache/`
- PID files em `/d02/pids/`
- Permissões corretas (ec2-user:ec2-user)
- Processos isolados por PID

---

**🎉 Sistema pronto para produção no Amazon Linux!**
