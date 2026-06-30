package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"rastreamento-pedidos/db"
)

type RequisicaoTransportadora struct {
	Nome        string `json:"nome"`
	Tipo        string `json:"tipo"`         // "correios", "jadlog", "custom"
	EndpointURL string `json:"endpoint_url"` // obrigatório se tipo = "custom"
	AuthTipo    string `json:"auth_tipo"`    // "none", "bearer", "api_key"
	AuthValor   string `json:"auth_valor"`   // o token
	JsonPath    string `json:"json_path"`    // ex: "data.status"
}

func CadastrarTransportadora(w http.ResponseWriter, r *http.Request) {
	var req RequisicaoTransportadora
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Nome == "" || req.Tipo == "" {
		http.Error(w, "Informe nome e tipo da transportadora", http.StatusBadRequest)
		return
	}

	if req.Tipo == "custom" && req.EndpointURL == "" {
		http.Error(w, "Transportadoras customizadas precisam de endpoint_url", http.StatusBadRequest)
		return
	}

	var id int
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO transportadoras (nome, tipo, endpoint_url, auth_tipo, auth_valor, json_path)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		req.Nome, req.Tipo, req.EndpointURL, req.AuthTipo, req.AuthValor, req.JsonPath,
	).Scan(&id)

	if err != nil {
		http.Error(w, "Erro ao salvar transportadora", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"mensagem": "Transportadora cadastrada com sucesso",
		"id":       id,
		"nome":     req.Nome,
		"tipo":     req.Tipo,
	})
}

func ListarTransportadoras(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, nome, tipo, COALESCE(endpoint_url,''), COALESCE(auth_tipo,'none'), COALESCE(ativo, true) FROM transportadoras ORDER BY nome`,
	)
	if err != nil {
		http.Error(w, "Erro ao listar transportadoras", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Item struct {
		ID          int    `json:"id"`
		Nome        string `json:"nome"`
		Tipo        string `json:"tipo"`
		EndpointURL string `json:"endpoint_url"`
		AuthTipo    string `json:"auth_tipo"`
		Ativo       bool   `json:"ativo"`
	}

	var lista []Item
	for rows.Next() {
		var t Item
		rows.Scan(&t.ID, &t.Nome, &t.Tipo, &t.EndpointURL, &t.AuthTipo, &t.Ativo)
		lista = append(lista, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}
