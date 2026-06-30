package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"rastreamento-pedidos/db"
)

func ClienteStats(w http.ResponseWriter, r *http.Request) {
	id := clienteIDDaRequisicao(r)

	var rastreios, chamadas, webhooks int
	db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM rastreios WHERE cliente_id = $1`, id,
	).Scan(&rastreios)
	db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM uso_api WHERE cliente_id = $1`, id,
	).Scan(&chamadas)
	db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM webhooks WHERE cliente_id = $1`, id,
	).Scan(&webhooks)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"rastreios": rastreios,
		"chamadas":  chamadas,
		"webhooks":  webhooks,
	})
}

func ClienteRastreios(w http.ResponseWriter, r *http.Request) {
	id := clienteIDDaRequisicao(r)

	rows, err := db.Pool.Query(context.Background(),
		`SELECT r.codigo, r.status, COALESCE(r.localizacao,''), t.nome,
		        COALESCE(r.destinatario_nome,''), COALESCE(r.criado_em, NOW())
		 FROM rastreios r
		 LEFT JOIN transportadoras t ON t.id = r.transportadora_id
		 WHERE r.cliente_id = $1
		 ORDER BY r.criado_em DESC LIMIT 100`, id,
	)
	if err != nil {
		http.Error(w, "Erro ao buscar rastreios", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Item struct {
		Codigo       string `json:"codigo"`
		Status       string `json:"status"`
		Localizacao  string `json:"localizacao"`
		Transportadora string `json:"transportadora"`
		Destinatario string `json:"destinatario"`
		CriadoEm    string `json:"criado_em"`
	}

	var lista []Item
	for rows.Next() {
		var it Item
		rows.Scan(&it.Codigo, &it.Status, &it.Localizacao, &it.Transportadora, &it.Destinatario, &it.CriadoEm)
		lista = append(lista, it)
	}
	if lista == nil {
		lista = []Item{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

func ClientePerfil(w http.ResponseWriter, r *http.Request) {
	id := clienteIDDaRequisicao(r)

	var nome, email, plano, apiKey, provider string
	db.Pool.QueryRow(context.Background(),
		`SELECT nome, email, plano, api_key, COALESCE(whatsapp_provider,'evolution')
		 FROM clientes WHERE id = $1`, id,
	).Scan(&nome, &email, &plano, &apiKey, &provider)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"nome":               nome,
		"email":              email,
		"plano":              plano,
		"api_key":            apiKey,
		"whatsapp_provider":  provider,
	})
}

func ClienteSalvarConfig(w http.ResponseWriter, r *http.Request) {
	id := clienteIDDaRequisicao(r)

	var req struct {
		WhatsAppProvider string `json:"whatsapp_provider"`
		WhatsAppKey      string `json:"whatsapp_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	db.Pool.Exec(context.Background(),
		`UPDATE clientes SET whatsapp_provider = $1, whatsapp_key = $2 WHERE id = $3`,
		req.WhatsAppProvider, req.WhatsAppKey, id,
	)

	w.WriteHeader(http.StatusNoContent)
}
