package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"rastreamento-pedidos/db"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Sessões de clientes separadas das sessões admin
var (
	sessoesCliente   = map[string]sessaoCliente{}
	muSessoesCliente sync.RWMutex
)

type sessaoCliente struct {
	clienteID int
	expira    time.Time
}

type contextKeyCliente struct{}

func ClienteLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Senha == "" {
		http.Error(w, "Informe email e senha", http.StatusBadRequest)
		return
	}

	var id int
	var senhaHash string
	var senhaAlterada bool
	var nome, plano, apiKey string
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, nome, plano, api_key, COALESCE(senha,''), COALESCE(senha_alterada, false)
		 FROM clientes WHERE email = $1 AND ativo = true`,
		req.Email,
	).Scan(&id, &nome, &plano, &apiKey, &senhaHash, &senhaAlterada)

	if err != nil || senhaHash == "" {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(senhaHash), []byte(req.Senha)); err != nil {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// Gera token de sessão
	b := make([]byte, 24)
	rand.Read(b)
	token := hex.EncodeToString(b)

	muSessoesCliente.Lock()
	sessoesCliente[token] = sessaoCliente{clienteID: id, expira: time.Now().Add(8 * time.Hour)}
	muSessoesCliente.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token":           token,
		"nome":            nome,
		"plano":           plano,
		"api_key":         apiKey,
		"senha_alterada":  senhaAlterada,
	})
}

func ClienteLogout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Cliente-Token")
	muSessoesCliente.Lock()
	delete(sessoesCliente, token)
	muSessoesCliente.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func ClienteAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Cliente-Token")
		muSessoesCliente.RLock()
		sess, ok := sessoesCliente[token]
		muSessoesCliente.RUnlock()

		if !ok || time.Now().After(sess.expira) {
			http.Error(w, "Não autorizado", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyCliente{}, sess.clienteID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func clienteIDDaRequisicao(r *http.Request) int {
	id, _ := r.Context().Value(contextKeyCliente{}).(int)
	return id
}

func ClienteTrocarSenha(w http.ResponseWriter, r *http.Request) {
	clienteID := clienteIDDaRequisicao(r)

	var req struct {
		SenhaAtual string `json:"senha_atual"`
		SenhaNova  string `json:"senha_nova"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SenhaNova == "" {
		http.Error(w, "Informe a nova senha", http.StatusBadRequest)
		return
	}

	if len(req.SenhaNova) < 6 {
		http.Error(w, "Senha deve ter no mínimo 6 caracteres", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.SenhaNova), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erro ao processar senha", http.StatusInternalServerError)
		return
	}

	db.Pool.Exec(context.Background(),
		`UPDATE clientes SET senha = $1, senha_alterada = true WHERE id = $2`,
		string(hash), clienteID,
	)

	w.WriteHeader(http.StatusNoContent)
}

// DefinirSenhaCliente — chamado pelo admin ao criar/resetar senha
func DefinirSenhaCliente(clienteID int, senha string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Pool.Exec(context.Background(),
		`UPDATE clientes SET senha = $1, senha_alterada = false WHERE id = $2`,
		string(hash), clienteID,
	)
	return err
}

