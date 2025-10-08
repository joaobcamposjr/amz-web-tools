# 🔥 AWS Security Group - Liberar Porta 8080

## ❌ Problema Identificado:
O backend está rodando corretamente em `0.0.0.0:8080`, mas o **AWS Security Group está bloqueando** conexões externas na porta 8080.

## ✅ Solução:

### 1. Acessar AWS Console
- Faça login em: https://console.aws.amazon.com/ec2/
- Navegue para: **EC2 > Instances**
- Selecione a instância: `ip-172-31-3-255`

### 2. Localizar Security Group
- Na aba **"Details"** ou **"Description"**, encontre **"Security groups"**
- Clique no link do Security Group (ex: `sg-xxxxxxxxx`)

### 3. Editar Inbound Rules
- Clique em **"Edit inbound rules"**
- Clique em **"Add rule"**

### 4. Adicionar Regra para Porta 8080
Configure a nova regra:
- **Type**: Custom TCP
- **Protocol**: TCP
- **Port Range**: `8080`
- **Source**: 
  - **Option 1 (Recomendado)**: `0.0.0.0/0` (qualquer IP)
  - **Option 2 (Mais seguro)**: Seu IP específico (ex: `201.x.x.x/32`)
- **Description**: `Backend API - AMZ Web Tools`

### 5. Salvar
- Clique em **"Save rules"**

### 6. Verificar (no servidor)
```bash
# Testar conexão externa
curl -v http://52.206.225.24:8080/health

# Deve retornar:
# HTTP/1.1 200 OK
# {"status":"ok"}
```

## 📋 Regras Necessárias:
Certifique-se que as seguintes portas estão liberadas:

| Port | Type | Source | Description |
|------|------|--------|-------------|
| 22 | SSH | Your IP | SSH Access |
| 3000 | Custom TCP | 0.0.0.0/0 | Frontend |
| 8080 | Custom TCP | 0.0.0.0/0 | Backend API |

## 🔐 Nota de Segurança:
Se quiser restringir acesso, use:
- **Source**: `Your IP/32` (apenas seu IP)
- Ou configure um **Application Load Balancer** com certificado SSL

## 🚀 Após Liberar:
1. Teste: `curl http://52.206.225.24:8080/health`
2. Acesse frontend: `http://52.206.225.24:3000`
3. Faça login: `admin@amztools.com` / `password`


