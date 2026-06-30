package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"rastreamento-pedidos/db"
)

type RequisicaoIntegracao struct {
	ClienteID  int    `json:"cliente_id"`
	Plataforma string `json:"plataforma"`
	APIKey     string `json:"api_key"`
}

func SalvarIntegracao(w http.ResponseWriter, r *http.Request) {
	var req RequisicaoIntegracao
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ClienteID == 0 || req.Plataforma == "" || req.APIKey == "" {
		http.Error(w, "Informe cliente_id, plataforma e api_key", http.StatusBadRequest)
		return
	}

	_, err := db.Pool.Exec(context.Background(),
		`INSERT INTO integracoes (cliente_id, plataforma, api_key, ativo)
		 VALUES ($1, $2, $3, true)
		 ON CONFLICT (cliente_id, plataforma)
		 DO UPDATE SET api_key = EXCLUDED.api_key, ativo = true`,
		req.ClienteID, req.Plataforma, req.APIKey,
	)
	if err != nil {
		http.Error(w, "Erro ao salvar integração: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func ListarIntegracoes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT i.id, i.cliente_id, c.nome, i.plataforma, i.ativo, i.ultimo_sync
		 FROM integracoes i
		 JOIN clientes c ON c.id = i.cliente_id
		 ORDER BY i.criado_em DESC`,
	)
	if err != nil {
		http.Error(w, "Erro ao listar integrações", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Item struct {
		ID         int     `json:"id"`
		ClienteID  int     `json:"cliente_id"`
		Cliente    string  `json:"cliente"`
		Plataforma string  `json:"plataforma"`
		Ativo      bool    `json:"ativo"`
		UltimoSync *string `json:"ultimo_sync"`
	}

	var lista []Item
	for rows.Next() {
		var it Item
		rows.Scan(&it.ID, &it.ClienteID, &it.Cliente, &it.Plataforma, &it.Ativo, &it.UltimoSync)
		lista = append(lista, it)
	}
	if lista == nil {
		lista = []Item{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

func ToggleIntegracao(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    int  `json:"id"`
		Ativo bool `json:"ativo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	db.Pool.Exec(context.Background(),
		`UPDATE integracoes SET ativo = $1 WHERE id = $2`, req.Ativo, req.ID,
	)
	w.WriteHeader(http.StatusNoContent)
}
