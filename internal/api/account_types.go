package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

type AccountTypeHandler struct {
	Queries *db.Queries
}

func NewAccountTypeHandler(queries *db.Queries) *AccountTypeHandler {
	return &AccountTypeHandler{
		Queries: queries,
	}
}

type CreateAccountTypeRequest struct {
	Name string `json:"name"`
}

type AccountTypeResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsSystem bool   `json:"is_system"`
}

func mapAccountTypeToResponse(accountType db.AccountType) AccountTypeResponse {
	IsSystem := false
	IsSystem = accountType.IsSystem == 1
	return AccountTypeResponse{
		ID:       accountType.ID,
		Name:     accountType.Name,
		IsSystem: IsSystem,
	}
}

// HandleListAccountTypesByUser godoc
// @Summary      List account types
// @Description  Get all system default account types and custom ones for the authenticated user
// @Tags         account-types
// @Produce      json
// @Success      200  {array}   AccountTypeResponse
// @Failure      401  {string}  string "Unauthorized"
// @Security     BearerAuth
// @Router       /account-types [get]
func (h AccountTypeHandler) HandleListAccountTypesByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	accountTypes, err := h.Queries.ListAccountTypesByUser(r.Context(), sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not list account types: "+err.Error())
		return
	}

	response := make([]AccountTypeResponse, 0, len(accountTypes))
	for _, t := range accountTypes {
		response = append(response, mapAccountTypeToResponse(t))
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HandleCreateAccountType godoc
// @Summary      Create a custom account type
// @Description  Create a new personalized account type for the authenticated user
// @Tags         account-types
// @Accept       json
// @Produce      json
// @Param        account_type  body      CreateAccountTypeRequest  true  "Account type data"
// @Success      201   {object}  AccountTypeResponse
// @Failure      400   {string}  string "Invalid request"
// @Failure      401   {string}  string "Unauthorized"
// @Security     BearerAuth
// @Router       /account-types [post]
func (h AccountTypeHandler) HandleCreateAccountType(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	var req CreateAccountTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	accountType, err := h.Queries.CreateAccountType(r.Context(), db.CreateAccountTypeParams{
		Name:     req.Name,
		UserID:   sql.NullInt64{Int64: userID, Valid: true},
		IsSystem: 0,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create account type: "+err.Error())
		return
	}

	response := mapAccountTypeToResponse(accountType)
	respondWithJSON(w, http.StatusCreated, response)
}

// HandleSoftDeleteAccountType godoc
// @Summary      Delete a custom account type
// @Description  Soft delete a user's custom account type by its ID (System types cannot be deleted)
// @Tags         account-types
// @Param        id    path      int  true  "Account Type ID"
// @Success      204   {object}  map[string]string "Account type deleted successfully"
// @Failure      400   {string}  string "Invalid ID"
// @Failure      401   {string}  string "Unauthorized"
// @Failure      500   {string}  string "Internal server error"
// @Security     BearerAuth
// @Router       /account-types/{id} [delete]
func (h AccountTypeHandler) HandleSoftDeleteAccountType(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Account Type ID")
		return
	}

	err = h.Queries.SoftDeleteAccountTypeByUser(r.Context(), db.SoftDeleteAccountTypeByUserParams{
		ID:     id,
		UserID: sql.NullInt64{Int64: userID, Valid: true},
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete account type: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, map[string]string{"result": "Account type deleted successfully"})
}
