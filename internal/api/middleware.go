package api

import (
	"context"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "userID"

func (api *Config) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value

		userID, err := api.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized, invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
