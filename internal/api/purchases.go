package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

type PurchaseHandler struct {
	Queries db.Querier
}

func NewPurchaseHandler(queries db.Querier) *PurchaseHandler {
	return &PurchaseHandler{
		Queries: queries,
	}
}

type CreatePurchaseRequest struct {
	Purchase           string  `json:"purchase"`
	Date               string  `json:"date"`
	Amount             int64   `json:"amount"`
	CarryOverNextMonth *int64  `json:"carry_over_next_month"`
	Comment            *string `json:"comment"`
	CardID             *int64  `json:"card_id"`
	AccountID          *int64  `json:"account_id"`
	ProjectID          *int64  `json:"project_id"`
}

type UpdatePurchaseRequest struct {
	Purchase           string  `json:"purchase"`
	Date               string  `json:"date"`
	Amount             int64   `json:"amount"`
	AmountPaid         int64   `json:"amount_paid"`
	CarryOverNextMonth *int64  `json:"carry_over_next_month"`
	Comment            *string `json:"comment"`
	CardID             *int64  `json:"card_id"`
	AccountID          *int64  `json:"account_id"`
	ProjectID          *int64  `json:"project_id"`
}

type PurchaseResponse struct {
	ID                 int64   `json:"id"`
	Purchase           string  `json:"purchase"`
	Date               string  `json:"date"`
	Amount             int64   `json:"amount"`
	AmountPaid         int64   `json:"amount_paid"`
	Remaining          int64   `json:"remaining"`
	Paid               bool    `json:"paid"`
	CarryOverNextMonth bool    `json:"carry_over_next_month"`
	Comment            *string `json:"comment,omitempty"`
	CardID             *int64  `json:"card_id,omitempty"`
	AccountID          *int64  `json:"account_id,omitempty"`
	ProjectID          *int64  `json:"project_id,omitempty"`
}

func mapPurchaseToResponse(p db.Purchase) PurchaseResponse {
	var comment *string
	if p.Comment.Valid {
		comment = &p.Comment.String
	}

	var cardID *int64
	if p.CardID.Valid {
		v := p.CardID.Int64
		cardID = &v
	}

	var accountID *int64
	if p.AccountID.Valid {
		v := p.AccountID.Int64
		accountID = &v
	}

	var projectID *int64
	if p.ProjectID.Valid {
		v := p.ProjectID.Int64
		projectID = &v
	}

	return PurchaseResponse{
		ID:                 p.ID,
		Purchase:           p.Purchase,
		Date:               p.Date,
		Amount:             p.Amount,
		AmountPaid:         p.AmountPaid,
		Remaining:          p.Remaining.Int64,
		Paid:               p.Paid.Int64 == 1,
		CarryOverNextMonth: p.CarryOverNextMonth.Int64 == 1,
		Comment:            comment,
		CardID:             cardID,
		AccountID:          accountID,
		ProjectID:          projectID,
	}
}

func (h *PurchaseHandler) validatePaymentSource(
	w http.ResponseWriter,
	r *http.Request,
	userID int64,
	cardID *int64,
	accountID *int64,
) (sql.NullInt64, sql.NullInt64, bool) {
	bothProvided := cardID != nil && accountID != nil
	noneProvided := cardID == nil && accountID == nil

	if bothProvided || noneProvided {
		respondWithError(w, http.StatusBadRequest, "Exactly one of card_id or account_id must be provided")
		return sql.NullInt64{}, sql.NullInt64{}, false
	}

	if cardID != nil {
		_, err := h.Queries.GetCardByID(r.Context(), db.GetCardByIDParams{
			ID:     *cardID,
			UserID: sql.NullInt64{Valid: true, Int64: userID},
		})
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusBadRequest, "Card does not exist or does not belong to the user")
			return sql.NullInt64{}, sql.NullInt64{}, false
		} else if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error verifying card: "+err.Error())
			return sql.NullInt64{}, sql.NullInt64{}, false
		}
		return sql.NullInt64{Valid: true, Int64: *cardID}, sql.NullInt64{Valid: false}, true
	}

	_, err := h.Queries.GetAccountByID(r.Context(), db.GetAccountByIDParams{
		ID:     *accountID,
		UserID: sql.NullInt64{Valid: true, Int64: userID},
	})
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusBadRequest, "Account does not exist or does not belong to the user")
		return sql.NullInt64{}, sql.NullInt64{}, false
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error verifying account: "+err.Error())
		return sql.NullInt64{}, sql.NullInt64{}, false
	}
	return sql.NullInt64{Valid: false}, sql.NullInt64{Valid: true, Int64: *accountID}, true
}

func validatePurchaseFields(w http.ResponseWriter, purchase, date string, amount int64) bool {
	if purchase == "" {
		respondWithError(w, http.StatusBadRequest, "purchase is required")
		return false
	}
	if date == "" {
		respondWithError(w, http.StatusBadRequest, "date is required")
		return false
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		respondWithError(w, http.StatusBadRequest, "date must be in YYYY-MM-DD format")
		return false
	}
	if amount <= 0 {
		respondWithError(w, http.StatusBadRequest, "amount must be a positive number")
		return false
	}
	return true
}

type purchaseNullables struct {
	comment   sql.NullString
	carryOver sql.NullInt64
	projectID sql.NullInt64
}

func buildPurchaseNullables(comment *string, carryOver *int64, projectID *int64) purchaseNullables {
	var n purchaseNullables

	if comment != nil {
		n.comment = sql.NullString{Valid: true, String: *comment}
	}
	if carryOver != nil {
		n.carryOver = sql.NullInt64{Valid: true, Int64: *carryOver}
	}
	if projectID != nil {
		n.projectID = sql.NullInt64{Valid: true, Int64: *projectID}
	}

	return n
}

// HandleCreatePurchase godoc
// @Summary      Create a new purchase
// @Description  Register a purchase for the authenticated user. Exactly one of card_id or account_id must be provided.
// @Tags         purchases
// @Accept       json
// @Produce      json
// @Param        purchase body CreatePurchaseRequest true "Purchase data"
// @Success      201  {object}  PurchaseResponse
// @Failure      400  {string}  string "Invalid request"
// @Failure      401  {string}  string "Unauthorized"
// @Failure      500  {string}  string "Internal server error"
// @Security     BearerAuth
// @Router       /purchases [post]
func (h *PurchaseHandler) HandleCreatePurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	var req CreatePurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Required field validations
	if !validatePurchaseFields(w, req.Purchase, req.Date, req.Amount) {
		return
	}

	resolvedCardID, resolvedAccountID, ok := h.validatePaymentSource(w, r, userID, req.CardID, req.AccountID)
	if !ok {
		return
	}

	nullables := buildPurchaseNullables(req.Comment, req.CarryOverNextMonth, req.ProjectID)

	purchase, err := h.Queries.CreatePurchase(r.Context(), db.CreatePurchaseParams{
		Purchase:           req.Purchase,
		Date:               req.Date,
		Amount:             req.Amount,
		CarryOverNextMonth: nullables.carryOver,
		Comment:            nullables.comment,
		CardID:             resolvedCardID,
		AccountID:          resolvedAccountID,
		ProjectID:          nullables.projectID,
		UserID:             sql.NullInt64{Valid: true, Int64: userID},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create purchase: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, mapPurchaseToResponse(purchase))
}

// HandleListPurchases godoc
// @Summary      List all purchases
// @Description  Get all purchases for the authenticated user, ordered by date descending
// @Tags         purchases
// @Produce      json
// @Success      200  {array}   PurchaseResponse
// @Failure      401  {string}  string "Unauthorized"
// @Failure      500  {string}  string "Internal server error"
// @Security     BearerAuth
// @Router       /purchases [get]
func (h *PurchaseHandler) HandleListPurchases(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	purchases, err := h.Queries.ListPurchasesByUserId(r.Context(), sql.NullInt64{Valid: true, Int64: userID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not list purchases: "+err.Error())
		return
	}

	response := make([]PurchaseResponse, 0, len(purchases))
	for _, p := range purchases {
		response = append(response, mapPurchaseToResponse(p))
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HandleGetPurchase godoc
// @Summary      Get a purchase by ID
// @Description  Retrieve a single purchase by its ID for the authenticated user
// @Tags         purchases
// @Produce      json
// @Param        id   path      int  true  "Purchase ID"
// @Success      200  {object}  PurchaseResponse
// @Failure      400  {string}  string "Invalid ID"
// @Failure      401  {string}  string "Unauthorized"
// @Failure      404  {string}  string "Purchase not found"
// @Security     BearerAuth
// @Router       /purchases/{id} [get]
func (h *PurchaseHandler) HandleGetPurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid purchase ID")
		return
	}

	purchase, err := h.Queries.GetPurchaseById(r.Context(), db.GetPurchaseByIdParams{
		ID:     id,
		UserID: sql.NullInt64{Valid: true, Int64: userID},
	})
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "Purchase not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get purchase: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, mapPurchaseToResponse(purchase))
}

// HandleUpdatePurchase godoc
// @Summary      Update a purchase
// @Description  Update an existing purchase by its ID for the authenticated user. Exactly one of card_id or account_id must be provided.
// @Tags         purchases
// @Accept       json
// @Produce      json
// @Param        id       path      int                   true  "Purchase ID"
// @Param        purchase body      UpdatePurchaseRequest true  "Updated purchase data"
// @Success      200  {object}  PurchaseResponse
// @Failure      400  {string}  string "Invalid request"
// @Failure      401  {string}  string "Unauthorized"
// @Failure      500  {string}  string "Internal server error"
// @Security     BearerAuth
// @Router       /purchases/{id} [put]
func (h *PurchaseHandler) HandleUpdatePurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid purchase ID")
		return
	}

	var req UpdatePurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if !validatePurchaseFields(w, req.Purchase, req.Date, req.Amount) {
		return
	}
	if req.AmountPaid < 0 {
		respondWithError(w, http.StatusBadRequest, "amount_paid cannot be negative")
		return
	}
	if req.AmountPaid > req.Amount {
		respondWithError(w, http.StatusBadRequest, "amount_paid cannot exceed amount")
		return
	}

	resolvedCardID, resolvedAccountID, ok := h.validatePaymentSource(w, r, userID, req.CardID, req.AccountID)
	if !ok {
		return
	}

	nullables := buildPurchaseNullables(req.Comment, req.CarryOverNextMonth, req.ProjectID)

	purchase, err := h.Queries.UpdatePurchase(r.Context(), db.UpdatePurchaseParams{
		ID:                 id,
		Purchase:           req.Purchase,
		Date:               req.Date,
		Amount:             req.Amount,
		AmountPaid:         req.AmountPaid,
		CarryOverNextMonth: nullables.carryOver,
		Comment:            nullables.comment,
		CardID:             resolvedCardID,
		AccountID:          resolvedAccountID,
		ProjectID:          nullables.projectID,
		UserID:             sql.NullInt64{Valid: true, Int64: userID},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update purchase: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, mapPurchaseToResponse(purchase))
}

// HandleDeletePurchase godoc
// @Summary      Delete a purchase
// @Description  Soft delete a purchase by its ID for the authenticated user
// @Tags         purchases
// @Param        id   path      int  true  "Purchase ID"
// @Success      204  "No Content"
// @Failure      400  {string}  string "Invalid ID"
// @Failure      401  {string}  string "Unauthorized"
// @Failure      500  {string}  string "Internal server error"
// @Security     BearerAuth
// @Router       /purchases/{id} [delete]
func (h *PurchaseHandler) HandleDeletePurchase(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIdFromContext(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid purchase ID")
		return
	}

	err = h.Queries.SoftDeletePurchase(r.Context(), db.SoftDeletePurchaseParams{
		ID:     id,
		UserID: sql.NullInt64{Valid: true, Int64: userID},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete purchase: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
