package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"rastreamento-pedidos/db"
)

func PainelStats(w http.ResponseWriter, r *http.Request) {
	var totalClientes, totalRastreios, totalChamadas int

	db.Pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM clientes`).Scan(&totalClientes)
	db.Pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM rastreios`).Scan(&totalRastreios)
	db.Pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM uso_api`).Scan(&totalChamadas)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"clientes":  totalClientes,
		"rastreios": totalRastreios,
		"chamadas":  totalChamadas,
	})
}

func PainelLogs(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT u.endpoint, u.criado_em, c.nome
		 FROM uso_api u
		 JOIN clientes c ON c.id = u.cliente_id
		 ORDER BY u.criado_em DESC LIMIT 200`,
	)
	if err != nil {
		http.Error(w, "Erro ao buscar logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Log struct {
		Endpoint  string `json:"endpoint"`
		CriadoEm string `json:"criado_em"`
		Cliente   string `json:"cliente"`
	}

	var lista []Log
	for rows.Next() {
		var l Log
		rows.Scan(&l.Endpoint, &l.CriadoEm, &l.Cliente)
		lista = append(lista, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

func PainelRastreios(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT r.codigo, r.status, COALESCE(r.localizacao,''), r.criado_em,
		 c.nome as cliente, t.nome as transportadora
		 FROM rastreios r
		 JOIN clientes c ON c.id = r.cliente_id
		 JOIN transportadoras t ON t.id = r.transportadora_id
		 ORDER BY r.criado_em DESC LIMIT 50`,
	)
	if err != nil {
		http.Error(w, "Erro ao buscar rastreios", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Item struct {
		Codigo         string `json:"codigo"`
		Status         string `json:"status"`
		Localizacao    string `json:"localizacao"`
		CriadoEm      string `json:"criado_em"`
		Cliente        string `json:"cliente"`
		Transportadora string `json:"transportadora"`
	}

	var lista []Item
	for rows.Next() {
		var i Item
		rows.Scan(&i.Codigo, &i.Status, &i.Localizacao, &i.CriadoEm, &i.Cliente, &i.Transportadora)
		lista = append(lista, i)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}
