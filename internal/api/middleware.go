package api

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"

type MiddlewareHandler struct {
	Auth           *AuthHandler
	EnableCORS     bool
	AllowedOrigins []string
}

func NewMiddlewareHandler(auth *AuthHandler, enableCORS bool, allowedOrigins []string) *MiddlewareHandler {
	return &MiddlewareHandler{
		Auth:           auth,
		EnableCORS:     enableCORS,
		AllowedOrigins: allowedOrigins,
	}
}

func (m *MiddlewareHandler) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var tokenString string

		cookie, err := r.Cookie("access_token")
		if err == nil {
			tokenString = cookie.Value
		}

		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		claims, err := m.Auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized, invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *MiddlewareHandler) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.EnableCORS {
			next.ServeHTTP(w, r)
			return
		}
		origin := r.Header.Get("Origin")
		isAllowed := false
		for _, o := range m.AllowedOrigins {
			if o == origin {
				isAllowed = true
				break
			}
		}
		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
