package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"
)

// Sessões em memória — token → expira em
var (
	sessoes   = map[string]time.Time{}
	sessoesMu sync.RWMutex
)

func gerarToken() string {
	b := make([]byte, 24)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if body.Email != os.Getenv("ADMIN_EMAIL") || body.Senha != os.Getenv("ADMIN_PASSWORD") {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	token := gerarToken()
	sessoesMu.Lock()
	sessoes[token] = time.Now().Add(8 * time.Hour)
	sessoesMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func AdminLogout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Admin-Token")
	sessoesMu.Lock()
	delete(sessoes, token)
	sessoesMu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// AdminAuthMiddleware protege rotas que só o admin pode acessar
func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Admin-Token")
		if token == "" {
			http.Error(w, "Não autorizado", http.StatusUnauthorized)
			return
		}

		sessoesMu.RLock()
		expira, existe := sessoes[token]
		sessoesMu.RUnlock()

		if !existe || time.Now().After(expira) {
			http.Error(w, "Sessão expirada — faça login novamente", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
