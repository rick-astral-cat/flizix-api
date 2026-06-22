package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

type CardHandler struct {
	Queries *db.Queries
}

func NewCardHandler(queries *db.Queries) *CardHandler {
	return &CardHandler{
		Queries: queries,
	}
}

type CreateCardRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	CreditLimit *int64 `json:"credit_limit"`
	CutoffDate  *int64 `json:"cutoff_date"`
	AccountID   *int64 `json:"account_id"`
}

type CardResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	CreditLimit *int64 `json:"credit_limit,omitempty"`
	CutoffDate  *int64 `json:"cutoff_date,omitempty"`
	AccountID   *int64 `json:"account_id"`
}

func mapCardToResponse(card db.Card) CardResponse {
	var creditLimit *int64
	if card.CreditLimit.Valid {
		val := card.CreditLimit.Int64
		creditLimit = &val
	}

	var accountID *int64
	if card.AccountID.Valid {
		val := card.AccountID.Int64
		accountID = &val
	}

	var cutoffDate *int64
	if card.CutoffDate.Valid {
		val := card.CutoffDate.Int64
		cutoffDate = &val
	}

	return CardResponse{
		ID:          card.ID,
		Name:        card.Name,
		Type:        card.Type,
		CreditLimit: creditLimit,
		CutoffDate:  cutoffDate,
		AccountID:   accountID,
	}
}

// HandleCreateCard godoc
// @Summary      Create a new card
// @Description  Create a credit or debit card for the authenticated user
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        card body CreateCardRequest true "Card data"
// @Success      201  {object} CardResponse
// @Failure      400  {string} string "Invalid request"
// @Security     BearerAuth
// @Router       /cards [post]
func (h *CardHandler) HandleCreateCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "no UserID in request context")
		return
	}

	var req CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if req.Type != "credit" && req.Type != "debit" {
		respondWithError(w, http.StatusBadRequest, "Type must be 'credit' or 'debit'")
		return
	}

	var creditLimit sql.NullInt64
	var cutoffDate sql.NullInt64
	var accountID sql.NullInt64

	if req.Type == "credit" {
		if req.CreditLimit == nil || *req.CreditLimit <= 0 {
			respondWithError(w, http.StatusBadRequest, "Credit limit must be a positive number for credit cards")
			return
		}
		if req.CutoffDate == nil {
			respondWithError(w, http.StatusBadRequest, "Cutoff date is required for credit cards")
			return
		}
		if *req.CutoffDate < 1 || *req.CutoffDate > 31 {
			respondWithError(w, http.StatusBadRequest, "Cutoff date must be a day of the month between 1 and 31")
			return
		}

		creditLimit = sql.NullInt64{Valid: true, Int64: *req.CreditLimit}
		cutoffDate = sql.NullInt64{Valid: true, Int64: *req.CutoffDate}
		accountID = sql.NullInt64{Valid: false}
	} else if req.Type == "debit" {
		if req.AccountID == nil {
			respondWithError(w, http.StatusBadRequest, "Debit account ID is required")
			return
		}

		acc, err := h.Queries.GetAccountByID(r.Context(), db.GetAccountByIDParams{
			ID:     *req.AccountID,
			UserID: sql.NullInt64{Valid: true, Int64: userID},
		})
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusBadRequest, "The associated account does not exist or does not belong to the user")
			return
		} else if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error verifying account: "+err.Error())
			return
		}

		accountID = sql.NullInt64{Valid: true, Int64: acc.ID}
		cutoffDate = sql.NullInt64{Valid: false}
		creditLimit = sql.NullInt64{Valid: false}
	}

	card, err := h.Queries.CreateCard(r.Context(), db.CreateCardParams{
		Name:        req.Name,
		Type:        req.Type,
		CreditLimit: creditLimit,
		CutoffDate:  cutoffDate,
		AccountID:   accountID,
		UserID:      sql.NullInt64{Int64: userID, Valid: true},
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create card: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, mapCardToResponse(card))
}

// HandleListCards godoc
// @Summary      List all cards
// @Description  Get all credit and debit cards for the authenticated user
// @Tags         cards
// @Produce      json
// @Success      200  {array}   CardResponse
// @Failure      401  {string}  string "Not authorized"
// @Security     BearerAuth
// @Router       /cards [get]
func (h *CardHandler) HandleListCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "No UserID in request context")
		return
	}

	cards, err := h.Queries.ListCardsByUser(r.Context(), sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not list cards: "+err.Error())
		return
	}

	response := make([]CardResponse, 0)
	for _, card := range cards {
		response = append(response, mapCardToResponse(card))
	}

	respondWithJSON(w, http.StatusOK, response)
}
