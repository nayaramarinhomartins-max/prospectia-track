package main

import (
	"log"
	"net/http"
	"os"
	"rastreamento-pedidos/db"
	"rastreamento-pedidos/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	db.Connect()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)
	r.Use(rateLimiter)

	// Painel web estático
	r.Handle("/", http.FileServer(http.Dir("static")))
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Auth admin
	r.Post("/auth/login", handlers.AdminLogin)
	r.Post("/auth/logout", handlers.AdminLogout)

	// Rotas públicas da API
	r.Group(func(r chi.Router) {
		r.Get("/transportadora", handlers.ListarTransportadoras)
		r.Get("/consulta/{codigo}", handlers.ConsultaPublica)
	})

	// Página pública de rastreio (comprador)
	r.Get("/rastreio", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/rastreio.html")
	})

	// Portal do cliente
	r.Get("/cliente", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/cliente.html")
	})
	r.Post("/cliente/login", handlers.ClienteLogin)
	r.Post("/cliente/logout", handlers.ClienteLogout)
	r.Group(func(r chi.Router) {
		r.Use(handlers.ClienteAuthMiddleware)
		r.Get("/cliente/stats", handlers.ClienteStats)
		r.Get("/cliente/rastreios", handlers.ClienteRastreios)
		r.Get("/cliente/perfil", handlers.ClientePerfil)
		r.Post("/cliente/config", handlers.ClienteSalvarConfig)
		r.Post("/cliente/senha", handlers.ClienteTrocarSenha)
	})

	// Rotas protegidas por API Key (clientes)
	r.Group(func(r chi.Router) {
		r.Use(handlers.AuthMiddleware)
		r.Post("/rastrear", handlers.Rastrear)
		r.Get("/historico/{codigo}", handlers.Historico)
		r.Post("/webhook/registrar", handlers.RegistrarWebhook)
		r.Post("/transportadora", handlers.CadastrarTransportadora)
	})

	// Rotas protegidas por token admin (painel)
	r.Group(func(r chi.Router) {
		r.Use(handlers.AdminAuthMiddleware)
		r.Post("/admin/cliente", handlers.CriarCliente)
		r.Get("/admin/clientes", handlers.ListarClientes)
		r.Get("/painel/stats", handlers.PainelStats)
		r.Get("/painel/rastreios", handlers.PainelRastreios)
		r.Get("/painel/logs", handlers.PainelLogs)
		r.Get("/admin/whatsapp/qr", handlers.WhatsAppQR)
		r.Get("/admin/whatsapp/status", handlers.WhatsAppStatus)
		r.Post("/admin/integracao", handlers.SalvarIntegracao)
		r.Get("/admin/integracoes", handlers.ListarIntegracoes)
		r.Post("/admin/integracao/toggle", handlers.ToggleIntegracao)
		r.Post("/admin/cliente/senha", handlers.ResetarSenhaCliente)
	})

	porta := os.Getenv("PORT")
	log.Printf("ProspectIA Track rodando em http://localhost:%s", porta)
	log.Fatal(http.ListenAndServe(":"+porta, r))
}

var limiter = rate.NewLimiter(10, 20)

func rateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Muitas requisições", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Api-Key, X-Admin-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
