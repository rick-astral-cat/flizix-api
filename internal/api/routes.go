package api

import (
	"net/http"

	_ "github.com/rick-astral-cat/flizix-api/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (api *Config) RegisterRoutes(mux *http.ServeMux, env string) {
	mux.HandleFunc("POST /users", api.HandleCreateUser)
	mux.HandleFunc("GET /health", api.HandleHealth)
	mux.HandleFunc("POST /auth/telegram", api.HandleTelegramLogin)

	if env == "development" {
		mux.Handle("GET /swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	}

	// Private Routes
	profileHandler := http.HandlerFunc(api.HandleGetProfile)
	mux.Handle("GET /me", api.JWTMiddleware(profileHandler))
}
