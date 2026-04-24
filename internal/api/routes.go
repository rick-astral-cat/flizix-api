package api

import "net/http"

func (api *Config) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", api.HandleCreateUser)
}
