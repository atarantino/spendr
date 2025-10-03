package auth

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)

func RequireAuth(sessionManager *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := sessionManager.GetInt(r.Context(), "userID")
			if userID == 0 {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				// http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) int {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0
	}
	return userID
}
