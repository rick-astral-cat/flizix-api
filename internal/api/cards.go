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
	CreditLimit int64  `json:"credit_limit"`
	CutoffDate  string `json:"cutoff_date"`
}

type CardResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	CreditLimit int64  `json:"credit_limit,omitempty"`
	CutoffDate  string `json:"cutoff_date"`
}

func mapCardToResponse(card db.Card) CardResponse {
	return CardResponse{
		ID:          card.ID,
		Name:        card.Name,
		Type:        card.Type,
		CreditLimit: card.CreditLimit.Int64,
		CutoffDate:  card.CutoffDate,
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
	}

	if req.Name == "" || (req.Type != "credit" && req.Type != "debit") {
		respondWithError(w, http.StatusBadRequest, "Name is required and Type must be credit or debit")
		return
	}

	card, err := h.Queries.CreateCard(r.Context(), db.CreateCardParams{
		Name: req.Name,
		Type: req.Type,
		CreditLimit: sql.NullInt64{
			Int64: req.CreditLimit,
			Valid: req.Type == "credit",
		},
		CutoffDate: req.CutoffDate,
		UserID:     sql.NullInt64{Int64: userID, Valid: true},
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
// @Router       /cards [get
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
