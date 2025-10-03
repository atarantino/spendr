package handlers

import (
	"net/http"

	"spendr/cmd/web"
	"spendr/internal/auth"

	"github.com/a-h/templ"
	"github.com/alexedwards/scs/v2"
)

type AuthHandler struct {
	authService    *auth.Service
	sessionManager *scs.SessionManager
}

func NewAuthHandler(authService *auth.Service, sessionManager *scs.SessionManager) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		sessionManager: sessionManager,
	}
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	templ.Handler(web.LoginPage()).ServeHTTP(w, r)
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	templ.Handler(web.RegisterPage()).ServeHTTP(w, r)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.authService.Login(r.Context(), email, password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid credentials"))
		return
	}

	h.sessionManager.Put(r.Context(), "userID", int(user.ID))
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.authService.Register(r.Context(), email, password, name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	h.sessionManager.Put(r.Context(), "userID", int(user.ID))
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.sessionManager.Destroy(r.Context())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
