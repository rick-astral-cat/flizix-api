package api

import "net/http"

func (api *Config) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Flizix service OK"))
}
