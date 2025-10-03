package server

import (
	"encoding/json"
	"log"
	"net/http"

	"spendr/cmd/web"
	"spendr/internal/auth"
	"spendr/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(s.sessionManager.LoadAndSave)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(s.authService, s.sessionManager)
	healthHandler := handlers.NewHealthHandler(s.db)
	wsHandler := handlers.NewWebSocketHandler()
	dashboardHandler := handlers.NewDashboardHandler(s.db)
	plaidHandler := handlers.NewPlaidHandler(s.plaidService, s.db)
	transactionHandler := handlers.NewTransactionHandler(s.db)
	walletsHandler := handlers.NewWalletsHandler(s.db)

	// Public routes
	r.Get("/", s.HelloWorldHandler)
	r.Get("/health", healthHandler.Health)
	r.Get("/websocket", wsHandler.WebSocket)

	// Auth routes
	r.Get("/login", authHandler.LoginPage)
	r.Post("/login", authHandler.Login)
	r.Get("/register", authHandler.RegisterPage)
	r.Post("/register", authHandler.Register)
	r.Post("/logout", authHandler.Logout)

	// Static assets
	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/assets/*", fileServer)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(s.sessionManager))
		r.Get("/dashboard", dashboardHandler.Dashboard)
		r.Get("/wallets", walletsHandler.WalletsPage)

		// Plaid API routes
		r.Post("/api/plaid/link/token", plaidHandler.CreateLinkToken)
		r.Post("/api/plaid/link/exchange", plaidHandler.ExchangePublicToken)
		r.Post("/api/plaid/sync", plaidHandler.SyncTransactions)
		r.Get("/api/plaid/accounts", plaidHandler.GetAccounts)

		// Transaction API routes
		r.Get("/api/transactions", transactionHandler.GetTransactions)
		r.Get("/api/transactions/uncategorized", transactionHandler.GetUncategorizedTransactions)
		r.Post("/api/transactions/{id}/categorize", transactionHandler.CategorizeTransaction)
		r.Delete("/api/transactions/{id}/categorize/{walletID}", transactionHandler.UncategorizeTransaction)
		r.Get("/api/wallets/{walletID}/transactions/shared", transactionHandler.GetSharedTransactions)

		// Wallet API routes
		r.Post("/api/wallets", walletsHandler.CreateWallet)
		r.Post("/api/wallets/{walletID}/members", walletsHandler.AddMember)
		r.Delete("/api/wallets/{walletID}/members/{memberID}", walletsHandler.RemoveMember)
	})

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
