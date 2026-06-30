package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"rastreamento-pedidos/db"
	"time"

	"github.com/go-chi/chi/v5"
)

type EventoRastreio struct {
	Status      string    `json:"status"`
	Descricao   string    `json:"descricao"`
	Localizacao string    `json:"localizacao"`
	Data        time.Time `json:"data"`
}

type ResultadoConsulta struct {
	Codigo          string           `json:"codigo"`
	Status          string           `json:"status"`
	Descricao       string           `json:"descricao"`
	Localizacao     string           `json:"localizacao"`
	Transportadora  string           `json:"transportadora"`
	DestinatarioNome string          `json:"destinatario_nome"`
	Historico       []EventoRastreio `json:"historico"`
	UltimaAtualizacao time.Time      `json:"ultima_atualizacao"`
}

// ConsultaPublica — sem autenticação, o comprador consulta pelo código do pedido
func ConsultaPublica(w http.ResponseWriter, r *http.Request) {
	codigo := chi.URLParam(r, "codigo")
	if codigo == "" {
		http.Error(w, "Informe o código", http.StatusBadRequest)
		return
	}

	// Busca histórico completo do código (todos os eventos)
	rows, err := db.Pool.Query(context.Background(),
		`SELECT r.status, COALESCE(r.descricao,''), COALESCE(r.localizacao,''),
		        COALESCE(r.criado_em, NOW()), COALESCE(t.nome,''), COALESCE(r.destinatario_nome,'')
		 FROM rastreios r
		 LEFT JOIN transportadoras t ON t.id = r.transportadora_id
		 WHERE r.codigo = $1
		 ORDER BY r.criado_em DESC`,
		codigo,
	)
	if err != nil {
		http.Error(w, "Erro ao consultar", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var resultado ResultadoConsulta
	resultado.Codigo = codigo
	resultado.Historico = []EventoRastreio{}

	for rows.Next() {
		var ev EventoRastreio
		var transp, nome string
		rows.Scan(&ev.Status, &ev.Descricao, &ev.Localizacao, &ev.Data, &transp, &nome)

		// Primeiro registro = mais recente = status atual
		if resultado.Status == "" {
			resultado.Status = ev.Status
			resultado.Descricao = ev.Descricao
			resultado.Localizacao = ev.Localizacao
			resultado.Transportadora = transp
			resultado.DestinatarioNome = nome
			resultado.UltimaAtualizacao = ev.Data
		}
		resultado.Historico = append(resultado.Historico, ev)
	}

	if resultado.Status == "" {
		http.Error(w, "Código não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultado)
}
