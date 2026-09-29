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

func TestHandleCreateCard_CreditSuccess(t *testing.T) {
	mock := &mockQuerier{
		createCard: func(_ context.Context, arg db.CreateCardParams) (db.Card, error) {
			return db.Card{
				ID:          1,
				Name:        arg.Name,
				Type:        arg.Type,
				CreditLimit: arg.CreditLimit,
				CutoffDate:  arg.CutoffDate,
			}, nil
		},
	}
	h := NewCardHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateCard))

	body := `{"name":"Visa","type":"credit","credit_limit":50000,"cutoff_date":15}`
	r := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var resp CardResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Name != "Visa" {
		t.Errorf("expected name Visa, got %s", resp.Name)
	}
}

func TestHandleCreateCard_DebitSuccess(t *testing.T) {
	mock := &mockQuerier{
		getAccountByID: func(_ context.Context, arg db.GetAccountByIDParams) (db.Account, error) {
			return db.Account{ID: arg.ID, Name: "BBVA"}, nil
		},
		createCard: func(_ context.Context, arg db.CreateCardParams) (db.Card, error) {
			return db.Card{ID: 2, Name: arg.Name, Type: arg.Type, AccountID: arg.AccountID}, nil
		},
	}
	h := NewCardHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateCard))

	body := `{"name":"BBVA Debit","type":"debit","account_id":1}`
	r := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestHandleCreateCard_InvalidType(t *testing.T) {
	h := NewCardHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateCard))

	body := `{"name":"Card","type":"prepaid"}`
	r := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreateCard_CreditMissingLimit(t *testing.T) {
	h := NewCardHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateCard))

	body := `{"name":"Visa","type":"credit","cutoff_date":15}`
	r := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreateCard_DebitMissingAccount(t *testing.T) {
	h := NewCardHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateCard))

	body := `{"name":"BBVA Debit","type":"debit"}`
	r := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreateCard_DebitAccountNotFound(t *testing.T) {
	mock := &mockQuerier{
		getAccountByID: func(_ context.Context, arg db.GetAccountByIDParams) (db.Account, error) {
			return db.Account{}, sql.ErrNoRows
		},
	}
	h := NewCardHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateCard))

	body := `{"name":"BBVA Debit","type":"debit","account_id":99}`
	r := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleListCards_Success(t *testing.T) {
	mock := &mockQuerier{
		listCardsByUser: func(_ context.Context, userID sql.NullInt64) ([]db.Card, error) {
			return []db.Card{
				{ID: 1, Name: "Visa", Type: "credit"},
				{ID: 2, Name: "BBVA Debit", Type: "debit"},
			}, nil
		},
	}
	h := NewCardHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleListCards))

	r := httptest.NewRequest(http.MethodGet, "/cards", nil)
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp []CardResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 cards, got %d", len(resp))
	}
}

func TestHandleCreateCard_CreditInvalidCutoffDate(t *testing.T) {
	h := NewCardHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateCard))

	// cutoff_date 32 is outside the valid 1-31 range
	body := `{"name":"Visa","type":"credit","credit_limit":50000,"cutoff_date":32}`
	r := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleListCards_Empty(t *testing.T) {
	mock := &mockQuerier{
		listCardsByUser: func(_ context.Context, userID sql.NullInt64) ([]db.Card, error) {
			return []db.Card{}, nil
		},
	}
	h := NewCardHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleListCards))

	r := httptest.NewRequest(http.MethodGet, "/cards", nil)
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp []CardResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d cards", len(resp))
	}
}
