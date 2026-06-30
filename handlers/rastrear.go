package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"rastreamento-pedidos/db"
	"rastreamento-pedidos/services"
)

type RequisicaoRastrear struct {
	Codigo              string `json:"codigo"`
	TransportadoraID    int    `json:"transportadora_id"`
	DestinatarioNome    string `json:"destinatario_nome"`
	DestinatarioEmail   string `json:"destinatario_email"`
	DestinatarioWhatsApp string `json:"destinatario_whatsapp"`
}

func Rastrear(w http.ResponseWriter, r *http.Request) {
	cliente := ClienteDaRequisicao(r)

	var req RequisicaoRastrear
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Codigo == "" || req.TransportadoraID == 0 {
		http.Error(w, "Informe codigo e transportadora_id", http.StatusBadRequest)
		return
	}

	var config services.ConfigTransportadora
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, nome, tipo, COALESCE(endpoint_url,''), COALESCE(auth_tipo,'none'), COALESCE(auth_valor,''), COALESCE(json_path,'')
		 FROM transportadoras WHERE id = $1 AND ativo = true`,
		req.TransportadoraID,
	).Scan(&config.ID, &config.Nome, &config.Tipo, &config.EndpointURL, &config.AuthTipo, &config.AuthValor, &config.JsonPath)

	if err != nil {
		http.Error(w, "Transportadora não encontrada ou inativa", http.StatusNotFound)
		return
	}

	transportadora, err := services.CriarTransportadora(config)
	if err != nil {
		http.Error(w, "Erro ao configurar transportadora: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Tenta até 3x com espera crescente se a transportadora cair: 1s → 2s → 4s
	resultado, err := services.ComRetry(services.PadraoRetry, config.Nome, func() (*services.ResultadoRastreio, error) {
		return transportadora.ConsultarRastreio(req.Codigo)
	})
	if err != nil {
		http.Error(w, "Erro ao rastrear após 3 tentativas: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Busca último status deste cliente para detectar mudança
	var statusAnterior string
	db.Pool.QueryRow(context.Background(),
		`SELECT status FROM rastreios WHERE codigo = $1 AND cliente_id = $2 ORDER BY criado_em DESC LIMIT 1`,
		req.Codigo, cliente.ID,
	).Scan(&statusAnterior)

	// Salva histórico com dados do destinatário
	_, err = db.Pool.Exec(context.Background(),
		`INSERT INTO rastreios (codigo, status, descricao, localizacao, transportadora_id, cliente_id,
		  destinatario_nome, destinatario_email, destinatario_whatsapp)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		resultado.Codigo, resultado.Status, resultado.Descricao, resultado.Localizacao,
		req.TransportadoraID, cliente.ID,
		req.DestinatarioNome, req.DestinatarioEmail, req.DestinatarioWhatsApp,
	)
	if err != nil {
		log.Printf("Erro ao salvar rastreio: %v", err)
	}

	// Se status mudou, dispara notificações em background
	if statusAnterior != "" && statusAnterior != resultado.Status {
		go dispararWebhooks(req.Codigo, resultado.Status, cliente.ID)
		go dispararNotificacoesV2(req, resultado, cliente.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultado)
}

// dispararNotificacoesV2 usa o dispatcher que respeita o provider configurado por cliente
func dispararNotificacoesV2(req RequisicaoRastrear, resultado *services.ResultadoRastreio, clienteID int) {
	nome := req.DestinatarioNome
	if nome == "" {
		nome = "Cliente"
	}

	// Busca configuração de WhatsApp do cliente
	var cfg services.ConfigNotificacao
	db.Pool.QueryRow(context.Background(),
		`SELECT COALESCE(whatsapp_provider,'evolution'), COALESCE(whatsapp_key,''), COALESCE(whatsapp_instance,'')
		 FROM clientes WHERE id = $1`, clienteID,
	).Scan(&cfg.WhatsAppProvider, &cfg.WhatsAppKey, &cfg.WhatsAppInstance)

	services.EnviarNotificacao(cfg,
		req.DestinatarioWhatsApp,
		req.DestinatarioEmail,
		nome,
		resultado.Codigo,
		resultado.Status,
		resultado.Localizacao,
	)
}

func dispararWebhooks(codigo, novoStatus string, clienteID int) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT url FROM webhooks WHERE codigo = $1 AND cliente_id = $2`, codigo, clienteID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	payload, _ := json.Marshal(map[string]string{"codigo": codigo, "status": novoStatus})

	for rows.Next() {
		var url string
		rows.Scan(&url)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("Webhook falhou para %s: %v", url, err)
			continue
		}
		resp.Body.Close()
	}
}
