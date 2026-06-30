# Deploy — ProspectIA Track

## Servidor

| Campo | Valor |
|---|---|
| Provedor | Hetzner |
| IP | 168.119.154.223 |
| OS | Ubuntu 24.04 |
| Usuário | root |
| Acesso SSH | `ssh root@168.119.154.223` (senha no chaves_mestre.env) |
| Acesso PuTTY | Host: `168.119.154.223` · Usuário: `root` |

## URLs em produção

| URL | O que é |
|---|---|
| https://track.prospectia.space/ | Painel admin |
| https://track.prospectia.space/cliente | Portal do cliente |
| https://track.prospectia.space/rastreio | Página pública de rastreio |

## Arquivos no servidor

```
/opt/prospectia-track/
├── prospectia-track   # binário Go
├── .env               # variáveis de ambiente
└── static/            # HTML do painel, portal e página pública
    ├── index.html
    ├── cliente.html
    └── rastreio.html
```

## Serviço systemd

```bash
# Ver status
systemctl status prospectia-track

# Ver logs em tempo real
journalctl -u prospectia-track -f

# Reiniciar
systemctl restart prospectia-track

# Parar
systemctl stop prospectia-track
```

## Como fazer redeploy

Toda vez que alterar o código, rodar estes comandos no Windows:

```powershell
# 1. Compilar para Linux
cd "C:\Users\User\Desktop\PROGRAMAS NAYARA CRIOU\PORTFOLIO DESAFIOS DEV\01 - Go - API de Rastreamento de Pedidos"
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o prospectia-track .

# 2. Enviar binário atualizado
& "C:\Program Files\PuTTY\pscp.exe" -pw "SENHA" prospectia-track root@168.119.154.223:/opt/prospectia-track/

# 3. Enviar static (se mudou o HTML)
& "C:\Program Files\PuTTY\pscp.exe" -pw "SENHA" -r static root@168.119.154.223:/opt/prospectia-track/

# 4. Reiniciar no servidor
& "C:\Program Files\PuTTY\plink.exe" -ssh root@168.119.154.223 -pw "SENHA" "systemctl restart prospectia-track"
```

## Nginx

Configuração em `/etc/nginx/sites-available/prospectia-track`  
Porta interna do app: **8085**  
Nginx escuta na 80 (HTTP → redireciona para HTTPS) e 443 (HTTPS)

## SSL

Certificado Let's Encrypt via Certbot.  
Expira em: **28/09/2026** (renova automaticamente).

```bash
# Forçar renovação manual se necessário
certbot renew --force-renewal
```

## DNS

Registrado na Hostinger:

| Tipo | Nome | Valor |
|---|---|---|
| A | track | 168.119.154.223 |
