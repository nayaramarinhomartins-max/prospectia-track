package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"rastreamento-pedidos/db"
	"rastreamento-pedidos/models"

	"github.com/go-chi/chi/v5"
)

func Historico(w http.ResponseWriter, r *http.Request) {
	cliente := ClienteDaRequisicao(r)
	codigo := chi.URLParam(r, "codigo")

	if codigo == "" {
		http.Error(w, "Código de rastreio obrigatório", http.StatusBadRequest)
		return
	}

	// Cliente só vê o histórico dos próprios rastreios
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, codigo, status, COALESCE(descricao,''), COALESCE(localizacao,''),
		 COALESCE(data_evento, NOW()), COALESCE(criado_em, NOW())
		 FROM rastreios
		 WHERE codigo = $1 AND cliente_id = $2
		 ORDER BY criado_em DESC`,
		codigo, cliente.ID,
	)
	if err != nil {
		http.Error(w, "Erro ao buscar histórico", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var historico []models.Rastreio
	for rows.Next() {
		var item models.Rastreio
		rows.Scan(&item.ID, &item.Codigo, &item.Status, &item.Descricao, &item.Localizacao, &item.DataEvento, &item.CriadoEm)
		historico = append(historico, item)
	}

	if len(historico) == 0 {
		http.Error(w, "Nenhum histórico encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(historico)
}
