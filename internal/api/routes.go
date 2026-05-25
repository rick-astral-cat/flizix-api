package api

import "net/http"

func (api *Config) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", api.HandleCreateUser)
	mux.HandleFunc("GET /health", api.HandleHealth)

	// Private Routes
	// profileHandler := http.HandlerFunc(api.HandleProfile)
	// mux.Handle("GET /profile", api.JWTMiddleware(profileHandler))
}
