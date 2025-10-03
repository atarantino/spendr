package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	_ "github.com/joho/godotenv/autoload"

	"spendr/internal/auth"
	"spendr/internal/database"
	"spendr/internal/plaid"
)

type Server struct {
	port int

	db             database.Service
	sessionManager *scs.SessionManager
	authService    *auth.Service
	plaidService   *plaid.Service
}

func NewServer() *http.Server {
	db := database.New()
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(db.GetPool())
	sessionManager.Lifetime = 24 * time.Hour

	port, _ := strconv.Atoi(os.Getenv("PORT"))

	NewServer := &Server{
		port: port,

		db:             database.New(),
		sessionManager: sessionManager,
		authService:    auth.NewService(db.GetQueries()),
		plaidService:   plaid.NewService(),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
