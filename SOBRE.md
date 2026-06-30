# Sobre este projeto

> Este arquivo existe para dar contexto a qualquer pessoa ou agente de IA que abrir esta pasta. Aqui está o propósito do projeto, as decisões tomadas e o que foi aprendido — informações que não cabem no README técnico.

---

## Contexto

Este é o **Projeto 01** de um portfólio de 20 projetos em linguagens diferentes, criado por Nayara Martins para ganhar experiência prática em stacks que ela ainda não dominava. Cada projeto é construído do zero, com intenção de produto real, não apenas como exercício de sintaxe.

**Linguagem deste projeto:** Go (Golang) — primeira experiência com a linguagem.

---

## De onde veio a ideia

O ponto de partida foi um desafio técnico simples: construir uma API REST em Go. Mas ao longo do desenvolvimento, a pergunta mudou de *"como fazer funcionar?"* para *"quem pagaria por isso e por quê?"*

O objetivo foi pensar com a cabeça de produto — entender a complexidade real por trás de algo que parece simples do lado de fora.

---

## O problema real

Rastreamento de pedidos existe em qualquer e-commerce. Mas o que está por baixo é mais complicado do que parece:

- Dezenas de transportadoras com APIs completamente diferentes entre si
- Compradores que querem saber do status sem precisar entrar em nenhum site
- Vendedores que já têm um ERP (Bling, Olist, Tiny) e não querem refazer integração do zero
- Notificações que precisam chegar no momento certo — sem ban de WhatsApp e sem custo absurdo
- Clientes que trazem transportadoras desconhecidas que você nunca vai ter tempo de integrar individualmente

---

## O que foi construído

Um **SaaS de rastreamento B2B** onde o cliente (vendedor) se cadastra, recebe uma API Key e conecta o sistema à transportadora que usa — conhecida ou desconhecida. A partir daí, os compradores dele recebem atualizações automáticas por WhatsApp e e-mail sem nenhuma intervenção manual.

O produto tem três camadas de acesso:

- **Painel admin** — visão completa de todos os clientes, logs e transportadoras
- **Portal do cliente** — cada vendedor vê só os próprios dados, configura o próprio WhatsApp
- **Página do comprador** — página pública onde qualquer pessoa rastreia um pedido pelo código

---

## O que este projeto exercita além do código

- Pensar em **multi-tenancy** desde a primeira linha — cada cliente vê só o que é dele
- Projetar **extensibilidade real**: qualquer transportadora se conecta sem alterar o core do sistema
- Entender **por que** certas escolhas técnicas existem — Go vs Node, pgxpool vs ORM, bcrypt vs hash simples
- Construir **camadas independentes**: API → portal do cliente → painel admin → página pública
- Equilibrar o que é viável agora com o que seria necessário para virar produto de verdade

---

## O que falta para virar produto

Este projeto não está completo como produto comercial — e isso é intencional. O foco foi na profundidade técnica, não na completude de negócio.

O que ainda seria necessário:

| O que | Por quê |
|---|---|
| Billing e planos com limites | Para cobrar por volume de rastreios |
| Polling automático de ERP | Para buscar pedidos do Bling sem o cliente fazer nada |
| Testes automatizados | Go tem testing nativo — seria o próximo passo natural |
| Integração real com Correios | A atual é simulada — Linketrack requer credenciais pagas |
| White-label | Página de rastreio com a marca do cliente, não da ProspectIA |

---

## Por que Go?

Go foi escolhido por ser a linguagem do desafio, mas ao longo do projeto ficou claro por que faz sentido para esse tipo de sistema:

- **Goroutines**: webhooks e notificações disparam em paralelo sem bloquear a resposta da API
- **Compilação**: um único binário sem dependências para fazer deploy — sem `npm install`, sem runtime instalado
- **Performance**: suporta centenas de milhares de requisições simultâneas com memória mínima
- **Tipagem forte**: erros aparecem em tempo de compilação, não em produção

A curva de aprendizado é suave para quem já conhece JavaScript — a sintaxe é parecida, o que muda é o modelo mental sobre concorrência e memória.

---

*Projeto 01 — Portfólio de desafios dev · ProspectIA · 2026*
