package handlers

import (
	"context"
	"net/http"
	"rastreamento-pedidos/db"
	"rastreamento-pedidos/models"
)

// chave usada para guardar o cliente no contexto da requisição
type contextKey string
const ClienteKey contextKey = "cliente"

// AuthMiddleware lê o header X-Api-Key, valida no banco e injeta o cliente no contexto.
// Se a chave não existir ou for inválida, rejeita com 401.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Api-Key")
		if apiKey == "" {
			http.Error(w, "X-Api-Key obrigatório", http.StatusUnauthorized)
			return
		}

		var cliente models.Cliente
		err := db.Pool.QueryRow(context.Background(),
			`SELECT id, nome, email, api_key, plano, ativo FROM clientes WHERE api_key = $1`,
			apiKey,
		).Scan(&cliente.ID, &cliente.Nome, &cliente.Email, &cliente.ApiKey, &cliente.Plano, &cliente.Ativo)

		if err != nil {
			http.Error(w, "API Key inválida", http.StatusUnauthorized)
			return
		}

		if !cliente.Ativo {
			http.Error(w, "Conta suspensa — entre em contato com o suporte", http.StatusForbidden)
			return
		}

		// Registra o uso para controle de volume
		go registrarUso(cliente.ID, r.URL.Path)

		// Passa o cliente adiante no contexto — handlers recuperam com ClienteDaRequisicao()
		ctx := context.WithValue(r.Context(), ClienteKey, &cliente)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ClienteDaRequisicao recupera o cliente autenticado de dentro de qualquer handler
func ClienteDaRequisicao(r *http.Request) *models.Cliente {
	cliente, _ := r.Context().Value(ClienteKey).(*models.Cliente)
	return cliente
}

func registrarUso(clienteID int, endpoint string) {
	db.Pool.Exec(context.Background(),
		`INSERT INTO uso_api (cliente_id, endpoint) VALUES ($1, $2)`,
		clienteID, endpoint,
	)
}
