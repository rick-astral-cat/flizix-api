package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

type Config struct {
	Queries *db.Queries
}

type CreateUserRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	PasskeyID string `json:"passkey_id"`
}

func (api *Config) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	passkey := sql.NullString{
		String: req.PasskeyID,
		Valid:  req.PasskeyID != "",
	}

	email := sql.NullString{
		String: req.Email,
		Valid:  true,
	}

	user, err := api.Queries.CreateUserWithPasskey(r.Context(), db.CreateUserWithPasskeyParams{
		Name:      req.Name,
		Email:     email,
		PasskeyID: passkey,
	})

	if err != nil {
		http.Error(w, "Error at creating user : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
