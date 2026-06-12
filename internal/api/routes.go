package api

import (
	"net/http"

	_ "github.com/rick-astral-cat/flizix-api/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func RegisterRoutes(mux *http.ServeMux, env string, userH *UserHandler, authH *AuthHandler, midH *MiddlewareHandler) {
	mux.HandleFunc("POST /users", userH.HandleCreateUser)
	mux.HandleFunc("GET /health", HandleHealth)
	mux.HandleFunc("POST /auth/telegram", authH.HandleTelegramLogin)
	mux.HandleFunc("POST /auth/logout", authH.HandleLogout)

	if env == "development" {
		mux.Handle("GET /swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	}

	// Private Routes
	profileHandler := http.HandlerFunc(userH.HandleGetProfile)
	mux.Handle("GET /me", midH.JWTMiddleware(profileHandler))
}
