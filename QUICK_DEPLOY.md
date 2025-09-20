# 🚀 Deploy Rápido - AMZ Web Tools

## ⚡ Deploy em 5 Passos

### 1. Preparar Servidor
```bash
# Upload e executar setup
scp setup-server.sh user@server:/tmp/
ssh user@server
chmod +x /tmp/setup-server.sh
sudo /tmp/setup-server.sh

# Logout e login novamente
exit
ssh user@server
```

### 2. Upload da Aplicação
```bash
# Criar diretório
sudo mkdir -p /opt/amz-web-tools
sudo chown $USER:$USER /opt/amz-web-tools

# Upload (do seu Mac)
scp -r /Users/user/GItProjects/amz-web-tools/* user@server:/opt/amz-web-tools/
```

### 3. Configurar Ambiente
```bash
cd /opt/amz-web-tools

# Copiar e editar configurações
cp env.production .env
nano .env  # Editar com suas configurações
```

### 4. Download Oracle Client
```bash
# Download manual do Oracle Instant Client
# https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html
# Arquivo: oracle-instantclient-basic-linux.x64-21.13.0.0.0dbru.zip

# Upload para o servidor
scp oracle-instantclient-basic-linux.x64-21.13.0.0.0dbru.zip user@server:/opt/amz-web-tools/oracle-client/
```

### 5. Deploy
```bash
cd /opt/amz-web-tools
chmod +x deploy.sh
./deploy.sh start
```

## 🔧 Configurações Importantes

### Arquivo `.env`
```bash
# Banco de dados (já configurado)
DB_HOST=54.204.42.134
DB_PASSWORD=321@Mudar@7089341@

# Oracle (já configurado)
ORACLE_HOST=10.13.2.159
ORACLE_PASSWORD=@Joao1225

# JWT (GERAR NOVA CHAVE!)
JWT_SECRET=sua_chave_secreta_super_forte_aqui

# Telegram (configurar)
TELEGRAM_BOT_TOKEN=seu_token_aqui
TELEGRAM_CHAT_ID=seu_chat_id_aqui
```

## 🌐 Acesso

- **Frontend**: `http://seu-servidor/`
- **API**: `http://seu-servidor/api/v1/`
- **Health**: `http://seu-servidor/health`

## 📊 Comandos Úteis

```bash
# Ver status
./deploy.sh status

# Ver logs
./deploy.sh logs

# Reiniciar
./deploy.sh restart

# Parar
./deploy.sh stop
```

## 🔍 Troubleshooting

### Oracle não conecta
```bash
# Verificar logs
docker-compose -f docker-compose.prod.yml logs backend | grep -i oracle
```

### Porta ocupada
```bash
# Verificar portas
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :8080
```

### Rebuild completo
```bash
docker-compose -f docker-compose.prod.yml down -v
./deploy.sh start
```

---

**🎉 Pronto! Sua aplicação está rodando em produção!**
