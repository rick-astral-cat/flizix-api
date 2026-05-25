package api

import (
	"context"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Temporary function for placeholder
func (api *Config) validateTokenPlaceholder(token string) (int64, error) {
	return 0, nil
}

func (api *Config) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value

		//Temporary placeholder on api, next commit will add JWT logic
		userID, err := api.validateTokenPlaceholder(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized, invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
