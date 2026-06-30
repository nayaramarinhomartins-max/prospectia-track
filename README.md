# ProspectIA Track

API de rastreamento de pedidos multi-transportadora com painel administrativo, portal do cliente e notificações automáticas via WhatsApp e e-mail.

Projeto 01 do portfólio de desafios dev — linguagem **Go (Golang)**.

---

## Stack

| Camada | Tecnologia |
|---|---|
| Linguagem | Go 1.22 |
| HTTP Router | Chi v5 |
| Banco de dados | PostgreSQL (Neon) |
| Driver de banco | pgxpool (pgx v5) |
| Autenticação | API Key + sessões em memória + bcrypt |
| E-mail | Resend (SDK v3) |
| WhatsApp | Evolution API / 360Dialog (WhatsApp Business) |
| Frontend | HTML + CSS + JS puro (sem framework) |
| Deploy | Binário único — sem dependências externas |

---

## Arquitetura

```
ProspectIA Track
├── Multi-tenant         → cada cliente tem API Key própria, dados isolados
├── Multi-transportadora → Correios, Jadlog e qualquer API via conector custom
├── Retry automático     → exponential backoff: 1s → 2s → 4s (máx 3 tentativas)
├── Notificações         → email (Resend) + WhatsApp (Evolution ou 360Dialog)
├── Webhooks             → disparo automático ao detectar mudança de status
└── Rate limiting        → 10 req/s com burst de 20
```

---

## Estrutura de pastas

```
.
├── main.go                     # Entry point — rotas e middlewares
├── .env                        # Credenciais (nunca commitar)
├── go.mod / go.sum
│
├── db/
│   └── db.go                   # Pool de conexões PostgreSQL
│
├── models/
│   └── cliente.go              # Struct do cliente
│
├── services/
│   ├── transportadora.go       # Interface + factory de transportadoras
│   ├── correios.go             # Integração Correios (via Linketrack)
│   ├── jadlog.go               # Integração Jadlog
│   ├── custom.go               # Conector genérico para qualquer API
│   ├── retry.go                # Exponential backoff retry
│   ├── email.go                # Notificações por e-mail (Resend)
│   ├── whatsapp.go             # Notificações WhatsApp (Evolution API)
│   ├── dialog360.go            # Notificações WhatsApp (360Dialog — oficial)
│   └── notificador.go          # Dispatcher: decide qual provider usar por cliente
│
├── handlers/
│   ├── auth.go                 # Middleware de API Key (clientes)
│   ├── authadmin.go            # Auth do painel admin (sessão + token)
│   ├── authcliente.go          # Auth do portal do cliente (sessão + bcrypt)
│   ├── admin.go                # CRUD de clientes + reset de senha
│   ├── rastrear.go             # POST /rastrear — consulta + notificação
│   ├── historico.go            # GET /historico/{codigo}
│   ├── webhook.go              # POST /webhook/registrar
│   ├── transportadora.go       # CRUD de transportadoras
│   ├── painel.go               # Stats, rastreios e logs do painel admin
│   ├── portalcliente.go        # Dashboard, rastreios e config do cliente
│   ├── consulta.go             # GET /consulta/{codigo} — página pública
│   ├── integracao.go           # CRUD de integrações ERP (Bling, Olist...)
│   └── whatsapp.go             # QR Code e status da Evolution API
│
└── static/
    ├── index.html              # Painel administrativo (você)
    ├── cliente.html            # Portal do cliente (vendedor)
    └── rastreio.html           # Página pública de rastreio (comprador)
```

---

## Variáveis de ambiente (.env)

```env
DATABASE_URL=postgresql://user:password@host:5432/dbname
PORT=8080
ADMIN_EMAIL=seu-email@exemplo.com
ADMIN_PASSWORD=sua-senha-aqui
APP_NAME=ProspectIA Track
RESEND_API_KEY=re_sua_chave_aqui
EVOLUTION_URL=https://sua-instancia-evolution.com
EVOLUTION_KEY=sua_chave_aqui
EVOLUTION_INSTANCE=nome_da_instancia
```

---

## Como rodar localmente

```bash
# 1. Clone o projeto
cd "01 - Go - API de Rastreamento de Pedidos"

# 2. Instale as dependências
go mod tidy

# 3. Configure o .env com suas credenciais

# 4. Rode o servidor
go run main.go

# Acesse:
# http://localhost:8080/          → Painel admin
# http://localhost:8080/cliente   → Portal do cliente
# http://localhost:8080/rastreio  → Página pública de rastreio
```

---

## Deploy (servidor Linux)

```bash
# Compila para Linux a partir do Windows
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o prospectia-track main.go

# Envia para o servidor
scp prospectia-track .env static/ root@168.119.154.223:/opt/prospectia-track/

# No servidor — cria serviço systemd
systemctl start prospectia-track
```

---

## API — Endpoints

### Autenticação
Todas as rotas de API exigem o header `X-Api-Key` com a chave do cliente.

```
POST /auth/login          → Login admin (retorna X-Admin-Token)
POST /auth/logout         → Logout admin
POST /cliente/login       → Login do portal do cliente
```

### Rastreamento
```
POST /rastrear            → Consulta e salva rastreio, dispara notificações
GET  /historico/{codigo}  → Histórico de um código (filtrado por cliente)
GET  /consulta/{codigo}   → Consulta pública (sem auth) — usada na página do comprador
```

### Webhooks
```
POST /webhook/registrar   → Registra URL para receber atualizações de status
```

### Transportadoras
```
GET  /transportadora      → Lista transportadoras ativas
POST /transportadora      → Cadastra nova transportadora (incluindo custom)
```

### Painel Admin (requer X-Admin-Token)
```
POST /admin/cliente           → Cria cliente + define senha
GET  /admin/clientes          → Lista clientes com métricas
POST /admin/cliente/senha     → Reseta senha de um cliente
GET  /painel/stats            → Contadores gerais
GET  /painel/rastreios        → Últimos 50 rastreios
GET  /painel/logs             → Últimas 200 chamadas à API
GET  /admin/whatsapp/qr       → QR Code para conectar WhatsApp (Evolution)
GET  /admin/whatsapp/status   → Status da instância WhatsApp
POST /admin/integracao        → Conecta plataforma ERP (Bling, Olist...)
GET  /admin/integracoes       → Lista integrações ativas
POST /admin/integracao/toggle → Ativa/pausa uma integração
```

### Portal do Cliente (requer X-Cliente-Token)
```
GET  /cliente/stats    → Contadores da conta
GET  /cliente/rastreios → Rastreios do cliente
GET  /cliente/perfil   → Perfil + API Key
POST /cliente/config   → Salva provider WhatsApp e chave
POST /cliente/senha    → Troca senha
```

---

## Transportadoras suportadas

| Transportadora | Tipo | Observação |
|---|---|---|
| Correios | `correios` | Via Linketrack (requer credenciais) |
| Jadlog | `jadlog` | Autenticação Bearer Token |
| Qualquer API | `custom` | URL configurável com `{codigo}`, suporte a none/bearer/api_key |

---

## Notificações

O sistema detecta mudança de status automaticamente a cada consulta. Quando o status muda:

1. Dispara **webhooks** cadastrados (goroutine paralela)
2. Envia **e-mail** via Resend
3. Envia **WhatsApp** via Evolution API ou 360Dialog — conforme config do cliente

### Providers de WhatsApp por cliente

| Provider | Tipo | Risco de ban |
|---|---|---|
| Evolution API | Não-oficial | Médio |
| 360Dialog | WhatsApp Business API oficial | Zero |
| Desativado | — | — |

---

## Banco de dados — Tabelas principais

```sql
clientes          → multi-tenant: API Key, plano, provider WhatsApp, senha portal
transportadoras   → config de cada transportadora (tipo, URL, auth, json_path)
rastreios         → histórico de consultas por cliente
webhooks          → URLs registradas por código + cliente
uso_api           → log de todas as chamadas (endpoint, cliente, data)
integracoes       → conexões com ERPs (Bling, Olist, Tiny, etc.)
```

---

## Conceitos de Go aplicados neste projeto

| Conceito | Onde aparece |
|---|---|
| **Interface** | `Transportadora` — qualquer carrier implementa o mesmo contrato |
| **Factory pattern** | `CriarTransportadora()` retorna a implementação correta |
| **Goroutines** | Webhooks e notificações disparam em paralelo sem bloquear a resposta |
| **pgxpool** | Pool de conexões — suporta 500+ requisições simultâneas |
| **context** | Propaga cliente autenticado pelo middleware até o handler |
| **bcrypt** | Hash seguro de senhas dos clientes do portal |
| **Exponential backoff** | Retry automático com espera crescente quando transportadora falha |
| **Compilação cruzada** | `GOOS=linux go build` gera binário Linux no Windows |

---

## Páginas

| URL | Público | Descrição |
|---|---|---|
| `/` | ❌ Admin | Painel de gestão completo |
| `/cliente` | ❌ Cliente | Portal do vendedor — dados isolados por conta |
| `/rastreio` | ✅ Todos | Consulta de status pelo comprador final |
| `/rastreio?codigo=XX` | ✅ Todos | Link direto (enviado no WhatsApp/email) |

---

*Projeto desenvolvido como portfólio — ProspectIA Track · 2026*
