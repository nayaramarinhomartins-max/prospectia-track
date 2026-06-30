package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"rastreamento-pedidos/db"
)

const charsetSenha = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789!@#$%"

func gerarSenhaAleatoria(tamanho int) (string, error) {
	b := make([]byte, tamanho)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i, v := range b {
		b[i] = charsetSenha[int(v)%len(charsetSenha)]
	}
	return string(b), nil
}

type RequisicaoCriarCliente struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Plano string `json:"plano"`
	Senha string `json:"senha"`
}

func CriarCliente(w http.ResponseWriter, r *http.Request) {
	var req RequisicaoCriarCliente
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Nome == "" || req.Email == "" {
		http.Error(w, "Informe nome e email", http.StatusBadRequest)
		return
	}
	if req.Plano == "" {
		req.Plano = "basic"
	}
	if req.Senha == "" {
		senhaGerada, err := gerarSenhaAleatoria(12)
		if err != nil {
			http.Error(w, "Erro ao gerar senha", http.StatusInternalServerError)
			return
		}
		req.Senha = senhaGerada
	}

	var id int
	var apiKey string
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO clientes (nome, email, plano)
		 VALUES ($1, $2, $3)
		 RETURNING id, api_key`,
		req.Nome, req.Email, req.Plano,
	).Scan(&id, &apiKey)
	if err != nil {
		http.Error(w, "Erro ao criar cliente — email já cadastrado?", http.StatusConflict)
		return
	}

	// Define a senha com bcrypt via Go
	if err := DefinirSenhaCliente(id, req.Senha); err != nil {
		http.Error(w, "Erro ao definir senha", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"mensagem": "Cliente criado com sucesso",
		"id":       id,
		"api_key":  apiKey,
		"plano":    req.Plano,
		"senha":    req.Senha,
	})
}

func ResetarSenhaCliente(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    int    `json:"id"`
		Senha string `json:"senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 || req.Senha == "" {
		http.Error(w, "Informe id e senha", http.StatusBadRequest)
		return
	}
	if err := DefinirSenhaCliente(req.ID, req.Senha); err != nil {
		http.Error(w, "Erro ao redefinir senha", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func ListarClientes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT c.id, c.nome, c.email, c.plano, c.ativo, c.criado_em,
		 COUNT(DISTINCT r.id) as total_rastreios,
		 COUNT(DISTINCT u.id) as total_chamadas
		 FROM clientes c
		 LEFT JOIN rastreios r ON r.cliente_id = c.id
		 LEFT JOIN uso_api u ON u.cliente_id = c.id
		 GROUP BY c.id ORDER BY c.criado_em DESC`,
	)
	if err != nil {
		http.Error(w, "Erro ao listar clientes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Item struct {
		ID             int    `json:"id"`
		Nome           string `json:"nome"`
		Email          string `json:"email"`
		Plano          string `json:"plano"`
		Ativo          bool   `json:"ativo"`
		CriadoEm       string `json:"criado_em"`
		TotalRastreios int    `json:"total_rastreios"`
		TotalChamadas  int    `json:"total_chamadas"`
	}

	var lista []Item
	for rows.Next() {
		var c Item
		rows.Scan(&c.ID, &c.Nome, &c.Email, &c.Plano, &c.Ativo, &c.CriadoEm, &c.TotalRastreios, &c.TotalChamadas)
		lista = append(lista, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}
