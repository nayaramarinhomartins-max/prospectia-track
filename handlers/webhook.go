package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"rastreamento-pedidos/db"
)

type RequisicaoWebhook struct {
	Codigo string `json:"codigo"`
	URL    string `json:"url"`
}

func RegistrarWebhook(w http.ResponseWriter, r *http.Request) {
	cliente := ClienteDaRequisicao(r)

	var req RequisicaoWebhook
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Codigo == "" || req.URL == "" {
		http.Error(w, "Informe codigo e url", http.StatusBadRequest)
		return
	}

	_, err := db.Pool.Exec(context.Background(),
		`INSERT INTO webhooks (codigo, url, cliente_id) VALUES ($1, $2, $3)
		 ON CONFLICT (codigo, url) DO NOTHING`,
		req.Codigo, req.URL, cliente.ID,
	)
	if err != nil {
		http.Error(w, "Erro ao registrar webhook", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"mensagem": "Webhook registrado",
		"codigo":   req.Codigo,
		"url":      req.URL,
	})
}
