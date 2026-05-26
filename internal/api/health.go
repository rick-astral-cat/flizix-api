package api

import "net/http"

// HandleHealth godoc
// @Summary	Check health system
// @Description	Returns simple message to validate server is running
// @Tags system
// @Produce plain
// @Success	200  {string}  string  "Flizix service OK"
// @Router	/health [get]
func (api *Config) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Flizix service OK"))
}
