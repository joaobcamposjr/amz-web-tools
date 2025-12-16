# 🔍 Diagnóstico Final - AMZ Web Tools

## ✅ Status dos Serviços

### Frontend (Porta 3000)
- **Status**: ✅ Funcionando localmente
- **Escutando em**: `*:3000` (0.0.0.0:3000) - CORRETO
- **Problema**: ⚠️ Não acessível externamente (Security Group AWS)

### Backend (Porta 8080)
- **Status**: ❌ Não está rodando
- **Motivo**: Falha ao conectar no banco de dados
- **Erro**: `dial tcp 54.204.42.134:1433: i/o timeout`

## 🔧 Correções Aplicadas

1. ✅ Removido Google Fonts do layout.tsx
2. ✅ Corrigido package.json (removido comandos maliciosos)
3. ✅ Ajustado next.config.js
4. ✅ Build do frontend concluído com sucesso
5. ✅ Frontend iniciado e funcionando
6. ⚠️ Backend não inicia por causa de conexão ao banco

## 🚨 Problemas Identificados

### 1. Backend não inicia
**Causa**: Backend tenta conectar no banco e falha com timeout
**Solução**: Modificado para iniciar mesmo sem banco (mas rotas que precisam de banco vão falhar)

### 2. Acesso externo bloqueado
**Causa**: Security Group da AWS não tem portas 3000 e 8080 liberadas
**Solução**: Adicionar regras no Security Group

## 📋 Ações Necessárias

### 1. Liberar Portas no Security Group da AWS

No console AWS:
1. Vá em EC2 > Security Groups
2. Encontre o Security Group da instância `52.206.225.24`
3. Adicione regras Inbound:
   - **Tipo**: Custom TCP
   - **Porta**: 3000
   - **Source**: 0.0.0.0/0
   - **Descrição**: Frontend AMZ Web Tools

   - **Tipo**: Custom TCP  
   - **Porta**: 8080
   - **Source**: 0.0.0.0/0
   - **Descrição**: Backend AMZ Web Tools

### 2. Verificar Conectividade com Banco de Dados

O banco `54.204.42.134:1433` não está acessível. Verificar:
- Firewall do SQL Server permite conexões do IP da instância
- Security Group do RDS (se for RDS) permite conexões
- SQL Server está configurado para aceitar conexões remotas

### 3. Reiniciar Backend

Após corrigir banco ou liberar Security Group, executar no servidor:

```bash
cd /d02/projects/amz-web-tools
./kill_all.sh
./fix_all_and_start.sh
```

## 🌐 URLs de Acesso

- **Frontend**: http://52.206.225.24:3000 (precisa liberar Security Group)
- **Backend API**: http://52.206.225.24:8080/api/v1/health (precisa liberar Security Group)

## 📊 Comandos Úteis

```bash
# Ver status dos serviços
lsof -ti:3000 && echo "Frontend OK" || echo "Frontend DOWN"
lsof -ti:8080 && echo "Backend OK" || echo "Backend DOWN"

# Ver logs
tail -f /d02/logs/frontend.log
tail -f /d02/logs/backend.log

# Reiniciar tudo
cd /d02/projects/amz-web-tools
./kill_all.sh
./fix_all_and_start.sh
```

