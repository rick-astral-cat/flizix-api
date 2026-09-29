package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

// --- Helpers ---

// basePurchase returns a db.Purchase with sensible defaults for use in mock returns.
func basePurchase() db.Purchase {
	return db.Purchase{
		ID:       1,
		Purchase: "Groceries",
		Date:     "2026-09-23",
		Amount:   500,
		CardID:   sql.NullInt64{Valid: true, Int64: 1},
	}
}

// newPurchaseHandler creates a PurchaseHandler with the given mock and a test middleware.
func newPurchaseHandler(mock *mockQuerier) (*PurchaseHandler, *MiddlewareHandler) {
	return NewPurchaseHandler(mock), newTestMiddleware()
}

// --- HandleCreatePurchase ---

func TestHandleCreatePurchase_WithCardSuccess(t *testing.T) {
	mock := &mockQuerier{
		getCardByID: func(_ context.Context, arg db.GetCardByIDParams) (db.Card, error) {
			return db.Card{ID: arg.ID}, nil
		},
		createPurchase: func(_ context.Context, arg db.CreatePurchaseParams) (db.Purchase, error) {
			return db.Purchase{
				ID:       1,
				Purchase: arg.Purchase,
				Date:     arg.Date,
				Amount:   arg.Amount,
				CardID:   arg.CardID,
			}, nil
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":500,"card_id":1}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var resp PurchaseResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Purchase != "Groceries" {
		t.Errorf("expected purchase 'Groceries', got '%s'", resp.Purchase)
	}
	if resp.Amount != 500 {
		t.Errorf("expected amount 500, got %d", resp.Amount)
	}
}

func TestHandleCreatePurchase_WithAccountSuccess(t *testing.T) {
	mock := &mockQuerier{
		getAccountByID: func(_ context.Context, arg db.GetAccountByIDParams) (db.Account, error) {
			return db.Account{ID: arg.ID}, nil
		},
		createPurchase: func(_ context.Context, arg db.CreatePurchaseParams) (db.Purchase, error) {
			return db.Purchase{
				ID:        1,
				Purchase:  arg.Purchase,
				Date:      arg.Date,
				Amount:    arg.Amount,
				AccountID: arg.AccountID,
			}, nil
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Transfer","date":"2026-09-23","amount":1000,"account_id":2}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_MissingPurchaseName(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"date":"2026-09-23","amount":500,"card_id":1}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_MissingDate(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Groceries","amount":500,"card_id":1}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_InvalidDateFormat(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Groceries","date":"23-09-2026","amount":500,"card_id":1}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_ZeroAmount(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":0,"card_id":1}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_NegativeAmount(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":-100,"card_id":1}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_BothPaymentSources(t *testing.T) {
	// Providing both card_id and account_id at the same time is invalid.
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":500,"card_id":1,"account_id":2}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_NoPaymentSource(t *testing.T) {
	// Providing neither card_id nor account_id is invalid.
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":500}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_CardNotFound(t *testing.T) {
	mock := &mockQuerier{
		getCardByID: func(_ context.Context, arg db.GetCardByIDParams) (db.Card, error) {
			return db.Card{}, sql.ErrNoRows
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":500,"card_id":99}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_AccountNotFound(t *testing.T) {
	mock := &mockQuerier{
		getAccountByID: func(_ context.Context, arg db.GetAccountByIDParams) (db.Account, error) {
			return db.Account{}, sql.ErrNoRows
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	body := `{"purchase":"Transfer","date":"2026-09-23","amount":500,"account_id":99}`
	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreatePurchase_InvalidJSON(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreatePurchase))

	r := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader("not-json"))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// --- HandleListPurchases ---

func TestHandleListPurchases_Success(t *testing.T) {
	mock := &mockQuerier{
		listPurchasesByUserId: func(_ context.Context, userID sql.NullInt64) ([]db.Purchase, error) {
			return []db.Purchase{
				basePurchase(),
				{ID: 2, Purchase: "Restaurant", Date: "2026-09-22", Amount: 300, CardID: sql.NullInt64{Valid: true, Int64: 1}},
			}, nil
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleListPurchases))

	r := httptest.NewRequest(http.MethodGet, "/purchases", nil)
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp []PurchaseResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 purchases, got %d", len(resp))
	}
}

func TestHandleListPurchases_Empty(t *testing.T) {
	// A user with no purchases should still get 200 with an empty array, not 404.
	mock := &mockQuerier{
		listPurchasesByUserId: func(_ context.Context, userID sql.NullInt64) ([]db.Purchase, error) {
			return []db.Purchase{}, nil
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleListPurchases))

	r := httptest.NewRequest(http.MethodGet, "/purchases", nil)
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp []PurchaseResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d items", len(resp))
	}
}

// --- HandleGetPurchase ---

func TestHandleGetPurchase_Success(t *testing.T) {
	mock := &mockQuerier{
		getPurchaseById: func(_ context.Context, arg db.GetPurchaseByIdParams) (db.Purchase, error) {
			return basePurchase(), nil
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleGetPurchase))

	r := httptest.NewRequest(http.MethodGet, "/purchases/1", nil)
	r.SetPathValue("id", "1")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp PurchaseResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.ID != 1 {
		t.Errorf("expected ID 1, got %d", resp.ID)
	}
}

func TestHandleGetPurchase_NotFound(t *testing.T) {
	mock := &mockQuerier{
		getPurchaseById: func(_ context.Context, arg db.GetPurchaseByIdParams) (db.Purchase, error) {
			return db.Purchase{}, sql.ErrNoRows
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleGetPurchase))

	r := httptest.NewRequest(http.MethodGet, "/purchases/99", nil)
	r.SetPathValue("id", "99")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleGetPurchase_InvalidID(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleGetPurchase))

	r := httptest.NewRequest(http.MethodGet, "/purchases/abc", nil)
	r.SetPathValue("id", "abc")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// --- HandleUpdatePurchase ---

func TestHandleUpdatePurchase_Success(t *testing.T) {
	mock := &mockQuerier{
		getCardByID: func(_ context.Context, arg db.GetCardByIDParams) (db.Card, error) {
			return db.Card{ID: arg.ID}, nil
		},
		updatePurchase: func(_ context.Context, arg db.UpdatePurchaseParams) (db.Purchase, error) {
			return db.Purchase{
				ID:         arg.ID,
				Purchase:   arg.Purchase,
				Date:       arg.Date,
				Amount:     arg.Amount,
				AmountPaid: arg.AmountPaid,
				CardID:     arg.CardID,
			}, nil
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleUpdatePurchase))

	body := `{"purchase":"Groceries updated","date":"2026-09-23","amount":600,"amount_paid":200,"card_id":1}`
	r := httptest.NewRequest(http.MethodPut, "/purchases/1", strings.NewReader(body))
	r.SetPathValue("id", "1")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp PurchaseResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.AmountPaid != 200 {
		t.Errorf("expected amount_paid 200, got %d", resp.AmountPaid)
	}
}

func TestHandleUpdatePurchase_AmountPaidExceedsAmount(t *testing.T) {
	// amount_paid > amount should be rejected before hitting the DB.
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleUpdatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":100,"amount_paid":200,"card_id":1}`
	r := httptest.NewRequest(http.MethodPut, "/purchases/1", strings.NewReader(body))
	r.SetPathValue("id", "1")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleUpdatePurchase_NegativeAmountPaid(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleUpdatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":500,"amount_paid":-50,"card_id":1}`
	r := httptest.NewRequest(http.MethodPut, "/purchases/1", strings.NewReader(body))
	r.SetPathValue("id", "1")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleUpdatePurchase_InvalidID(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleUpdatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":500,"card_id":1}`
	r := httptest.NewRequest(http.MethodPut, "/purchases/abc", strings.NewReader(body))
	r.SetPathValue("id", "abc")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleUpdatePurchase_BothPaymentSources(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleUpdatePurchase))

	body := `{"purchase":"Groceries","date":"2026-09-23","amount":500,"card_id":1,"account_id":2}`
	r := httptest.NewRequest(http.MethodPut, "/purchases/1", strings.NewReader(body))
	r.SetPathValue("id", "1")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// --- HandleDeletePurchase ---

func TestHandleDeletePurchase_Success(t *testing.T) {
	mock := &mockQuerier{
		softDeletePurchase: func(_ context.Context, arg db.SoftDeletePurchaseParams) error {
			return nil
		},
	}
	h, mid := newPurchaseHandler(mock)
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleDeletePurchase))

	r := httptest.NewRequest(http.MethodDelete, "/purchases/1", nil)
	r.SetPathValue("id", "1")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestHandleDeletePurchase_InvalidID(t *testing.T) {
	h, mid := newPurchaseHandler(&mockQuerier{})
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleDeletePurchase))

	r := httptest.NewRequest(http.MethodDelete, "/purchases/abc", nil)
	r.SetPathValue("id", "abc")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
