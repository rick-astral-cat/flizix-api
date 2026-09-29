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

func TestHandleCreateAccount_Success(t *testing.T) {
	mock := &mockQuerier{
		createAccount: func(_ context.Context, arg db.CreateAccountParams) (db.Account, error) {
			return db.Account{ID: 1, Name: arg.Name, Type: arg.Type}, nil
		},
	}
	h := NewAccountHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateAccount))

	body := `{"name":"BBVA","type":1}`
	r := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var resp AccountResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Name != "BBVA" {
		t.Errorf("expected name BBVA, got %s", resp.Name)
	}
}

func TestHandleCreateAccount_MissingName(t *testing.T) {
	h := NewAccountHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateAccount))

	body := `{"type":1}`
	r := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleListAccounts_Success(t *testing.T) {
	mock := &mockQuerier{
		listAccountsByUser: func(_ context.Context, userID sql.NullInt64) ([]db.Account, error) {
			return []db.Account{
				{ID: 1, Name: "BBVA", Type: sql.NullInt64{Int64: 1, Valid: true}},
				{ID: 2, Name: "Cash", Type: sql.NullInt64{Int64: 2, Valid: true}},
			}, nil
		},
	}
	h := NewAccountHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleListAccounts))

	r := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp []AccountResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(resp))
	}
}

func TestHandleDeleteAccount_Success(t *testing.T) {
	mock := &mockQuerier{
		softDeleteAccount: func(_ context.Context, arg db.SoftDeleteAccountParams) error {
			return nil
		},
	}
	h := NewAccountHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleDeleteAccount))

	r := httptest.NewRequest(http.MethodDelete, "/accounts/1", nil)
	r.SetPathValue("id", "1")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestHandleDeleteAccount_InvalidID(t *testing.T) {
	h := NewAccountHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleDeleteAccount))

	r := httptest.NewRequest(http.MethodDelete, "/accounts/abc", nil)
	r.SetPathValue("id", "abc")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleUpdateAccount_Success(t *testing.T) {
	mock := &mockQuerier{
		updateAccount: func(_ context.Context, arg db.UpdateAccountParams) (db.Account, error) {
			return db.Account{ID: arg.ID, Name: arg.Name, Type: arg.Type}, nil
		},
	}
	h := NewAccountHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleUpdateAccount))

	body := `{"name":"BBVA Updated","type":2}`
	r := httptest.NewRequest(http.MethodPut, "/accounts/1", strings.NewReader(body))
	r.SetPathValue("id", "1")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp AccountResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Name != "BBVA Updated" {
		t.Errorf("expected name 'BBVA Updated', got '%s'", resp.Name)
	}
}

func TestHandleUpdateAccount_InvalidID(t *testing.T) {
	h := NewAccountHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleUpdateAccount))

	body := `{"name":"BBVA","type":1}`
	r := httptest.NewRequest(http.MethodPut, "/accounts/abc", strings.NewReader(body))
	r.SetPathValue("id", "abc")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleUpdateAccount_InvalidJSON(t *testing.T) {
	h := NewAccountHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleUpdateAccount))

	r := httptest.NewRequest(http.MethodPut, "/accounts/1", strings.NewReader("not-json"))
	r.SetPathValue("id", "1")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
