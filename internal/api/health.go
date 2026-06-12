package api

import "net/http"

// HandleHealth godoc
// @Summary	Check health system
// @Description	Returns simple message to validate server is running
// @Tags system
// @Produce plain
// @Success	200  {string}  string  "Flizix service OK"
// @Router	/health [get]
func HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Flizix service OK")); err != nil {
		_ = err
	}
}
