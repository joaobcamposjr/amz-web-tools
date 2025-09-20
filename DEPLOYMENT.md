# 🚀 AMZ Web Tools - Deployment Guide

Este guia te ajudará a fazer o deploy da aplicação AMZ Web Tools em um servidor Linux de produção.

## 📋 Pré-requisitos

- Servidor Linux (Ubuntu 20.04+ ou CentOS 8+)
- Acesso root ou sudo
- Domínio configurado (opcional, mas recomendado)
- Portas 80, 443 e 8080 liberadas no firewall

## 🔧 Setup do Servidor

### 1. Configurar Servidor

Execute o script de setup no servidor:

```bash
# Fazer upload do script para o servidor
scp setup-server.sh user@your-server:/tmp/

# Conectar ao servidor
ssh user@your-server

# Executar setup
chmod +x /tmp/setup-server.sh
sudo /tmp/setup-server.sh
```

### 2. Logout e Login

Após a instalação do Docker, faça logout e login novamente para que as mudanças do grupo docker tenham efeito.

## 📁 Upload da Aplicação

### 1. Fazer Upload dos Arquivos

```bash
# Criar diretório da aplicação
sudo mkdir -p /opt/amz-web-tools
sudo chown $USER:$USER /opt/amz-web-tools

# Upload via SCP (do seu Mac)
scp -r /Users/user/GItProjects/amz-web-tools/* user@your-server:/opt/amz-web-tools/
```

### 2. Configurar Variáveis de Ambiente

```bash
# Copiar arquivo de exemplo
cp env.production .env

# Editar configurações
nano .env
```

**Configurações importantes:**
- `DB_HOST`: IP do servidor SQL Server
- `DB_PASSWORD`: Senha do banco de dados
- `ORACLE_HOST`: IP do servidor Oracle
- `ORACLE_PASSWORD`: Senha do Oracle
- `JWT_SECRET`: Chave secreta forte (gere uma nova!)
- `TELEGRAM_BOT_TOKEN`: Token do bot do Telegram
- `TELEGRAM_CHAT_ID`: ID do chat do Telegram

## 🐳 Oracle Instant Client

### 1. Download Manual

```bash
# Criar diretório
mkdir -p oracle-client

# Download do site oficial Oracle
# https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html
# Arquivo: oracle-instantclient-basic-linux.x64-21.13.0.0.0dbru.zip

# Upload via SCP
scp oracle-instantclient-basic-linux.x64-21.13.0.0.0dbru.zip user@your-server:/opt/amz-web-tools/oracle-client/
```

### 2. Configurar Volume Docker

O Docker vai extrair automaticamente o arquivo durante o build.

## 🌐 Configurar Domínio (Opcional)

### 1. DNS

Configure os registros DNS:
- `A` record: `yourdomain.com` → IP do servidor
- `CNAME` record: `www.yourdomain.com` → `yourdomain.com`

### 2. SSL Certificates

```bash
# Instalar Certbot
sudo apt-get install certbot

# Obter certificado
sudo certbot certonly --standalone -d yourdomain.com

# Copiar certificados
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem /opt/amz-web-tools/nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem /opt/amz-web-tools/nginx/ssl/key.pem
sudo chown $USER:$USER /opt/amz-web-tools/nginx/ssl/*.pem
```

### 3. Atualizar Nginx Config

Edite `nginx/nginx.conf` e descomente a seção HTTPS.

## 🚀 Deploy

### 1. Deploy Inicial

```bash
cd /opt/amz-web-tools

# Verificar se tudo está pronto
./deploy.sh

# Iniciar serviços
./deploy.sh start
```

### 2. Verificar Status

```bash
# Ver status dos containers
./deploy.sh status

# Ver logs
./deploy.sh logs

# Testar aplicação
curl http://your-server/health
```

## 📊 Monitoramento

### 1. Logs

```bash
# Logs em tempo real
./deploy.sh logs

# Logs específicos
docker-compose -f docker-compose.prod.yml logs backend
docker-compose -f docker-compose.prod.yml logs frontend
```

### 2. Status dos Serviços

```bash
# Status dos containers
docker-compose -f docker-compose.prod.yml ps

# Uso de recursos
docker stats
```

### 3. Health Checks

- Frontend: `http://your-server/`
- Backend: `http://your-server/health`
- API: `http://your-server/api/v1/`

## 🔄 Comandos de Deploy

```bash
# Iniciar serviços
./deploy.sh start

# Parar serviços
./deploy.sh stop

# Reiniciar serviços
./deploy.sh restart

# Ver logs
./deploy.sh logs

# Ver status
./deploy.sh status
```

## 🔧 Troubleshooting

### 1. Problemas de Conexão Oracle

```bash
# Verificar se Oracle client está montado
docker volume ls
docker volume inspect amz-web-tools_oracle_client

# Verificar logs do backend
docker-compose -f docker-compose.prod.yml logs backend | grep -i oracle
```

### 2. Problemas de Porta

```bash
# Verificar portas em uso
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :8080

# Parar serviços que podem estar usando as portas
sudo systemctl stop apache2  # Se estiver rodando
sudo systemctl stop nginx    # Se estiver rodando
```

### 3. Problemas de Permissão

```bash
# Corrigir permissões
sudo chown -R $USER:$USER /opt/amz-web-tools
chmod +x /opt/amz-web-tools/deploy.sh
```

### 4. Rebuild Completo

```bash
# Parar e remover tudo
docker-compose -f docker-compose.prod.yml down -v

# Rebuild completo
./deploy.sh start
```

## 📝 Atualizações

### 1. Atualizar Código

```bash
# Fazer backup
cp -r /opt/amz-web-tools /opt/amz-web-tools.backup

# Upload nova versão
scp -r /path/to/new/code/* user@your-server:/opt/amz-web-tools/

# Reiniciar serviços
./deploy.sh restart
```

### 2. Rollback

```bash
# Restaurar backup
rm -rf /opt/amz-web-tools
mv /opt/amz-web-tools.backup /opt/amz-web-tools

# Reiniciar
./deploy.sh restart
```

## 🔒 Segurança

### 1. Firewall

```bash
# Verificar regras
sudo ufw status

# Permitir apenas portas necessárias
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
```

### 2. SSL/TLS

- Use sempre HTTPS em produção
- Renove certificados automaticamente com cron
- Configure HSTS headers no Nginx

### 3. Backup

```bash
# Backup do banco de dados (SQL Server)
# Configure backup automático no SQL Server

# Backup da aplicação
tar -czf amz-web-tools-backup-$(date +%Y%m%d).tar.gz /opt/amz-web-tools
```

## 📞 Suporte

Em caso de problemas:

1. Verifique os logs: `./deploy.sh logs`
2. Verifique o status: `./deploy.sh status`
3. Teste conectividade: `curl http://localhost/health`
4. Verifique configurações: `cat .env`

---

**🎉 Parabéns! Sua aplicação AMZ Web Tools está rodando em produção!**
