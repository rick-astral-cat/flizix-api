package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

type AccountHandler struct {
	Queries *db.Queries
}

func NewAccountHandler(queries *db.Queries) *AccountHandler {
	return &AccountHandler{
		Queries: queries,
	}
}

type CreateAccountRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AccountResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func mapAccountToResponse(acc db.Account) AccountResponse {
	return AccountResponse{
		ID:   acc.ID,
		Name: acc.Name,
		Type: acc.Type,
	}
}

// HandleCreateAccount godoc
// @Summary      Create a new bank account
// @Description  Create a bank or cash account for the authenticated user
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        account body CreateAccountRequest true "Account data"
// @Success      201  {object} AccountResponse
// @Failure      400  {string} string "Invalid request"
// @Failure      401  {string} string "Unauthorized"
// @Security     BearerAuth
// @Router       /accounts [post]
func (h AccountHandler) HandleCreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Account name is required")
		return
	}

	acc, err := h.Queries.CreateAccount(r.Context(), db.CreateAccountParams{
		Name:   req.Name,
		Type:   req.Type,
		UserID: sql.NullInt64{Int64: userID, Valid: true},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create account: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, mapAccountToResponse(acc))
}

// HandleListAccounts godoc
// @Summary      List all bank accounts
// @Description  Get all bank and cash accounts for the authenticated user
// @Tags         accounts
// @Produce      json
// @Success      200  {array}   AccountResponse
// @Failure      401  {string}  string "Unauthorized"
// @Security     BearerAuth
// @Router       /accounts [get]
func (h AccountHandler) HandleListAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	accounts, err := h.Queries.ListAccountsByUser(r.Context(), sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not list accounts: "+err.Error())
		return
	}

	response := make([]AccountResponse, 0)
	for _, acc := range accounts {
		response = append(response, mapAccountToResponse(acc))
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HandleDeleteAccount godoc
// @Summary      Delete a bank account
// @Description  Soft delete a specific bank account for the authenticated user
// @Tags         accounts
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {string}  string "Invalid ID"
// @Failure      401  {string}  string "Unauthorized"
// @Security     BearerAuth
// @Router       /accounts/{id} [delete]
func (h AccountHandler) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Account ID")
		return
	}

	err = h.Queries.SoftDeleteAccount(r.Context(), db.SoftDeleteAccountParams{
		ID:     id,
		UserID: sql.NullInt64{Int64: userID, Valid: true},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not soft delete account: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, map[string]string{"message": "account deleted"})
}
