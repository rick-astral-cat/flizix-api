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

func TestHandleListAccountTypes_Success(t *testing.T) {
	mock := &mockQuerier{
		listAccountTypesByUser: func(_ context.Context, userID sql.NullInt64) ([]db.AccountType, error) {
			return []db.AccountType{
				{ID: 1, Name: "account_type.checking", IsSystem: 1},
				{ID: 2, Name: "account_type.cash", IsSystem: 1},
				{ID: 3, Name: "Savings", IsSystem: 0},
			}, nil
		},
	}
	h := NewAccountTypeHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleListAccountTypesByUser))

	r := httptest.NewRequest(http.MethodGet, "/account-types", nil)
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp []AccountTypeResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(resp) != 3 {
		t.Errorf("expected 3 account types, got %d", len(resp))
	}
}

func TestHandleListAccountTypes_Empty(t *testing.T) {
	mock := &mockQuerier{
		listAccountTypesByUser: func(_ context.Context, userID sql.NullInt64) ([]db.AccountType, error) {
			return []db.AccountType{}, nil
		},
	}
	h := NewAccountTypeHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleListAccountTypesByUser))

	r := httptest.NewRequest(http.MethodGet, "/account-types", nil)
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp []AccountTypeResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d items", len(resp))
	}
}

func TestHandleCreateAccountType_Success(t *testing.T) {
	mock := &mockQuerier{
		createAccountType: func(_ context.Context, arg db.CreateAccountTypeParams) (db.AccountType, error) {
			return db.AccountType{ID: 1, Name: arg.Name, IsSystem: 0}, nil
		},
	}
	h := NewAccountTypeHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateAccountType))

	body := `{"name":"Savings"}`
	r := httptest.NewRequest(http.MethodPost, "/account-types", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var resp AccountTypeResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Name != "Savings" {
		t.Errorf("expected name 'Savings', got '%s'", resp.Name)
	}
	if resp.IsSystem {
		t.Error("user-created type should not be marked as system")
	}
}

func TestHandleCreateAccountType_MissingName(t *testing.T) {
	h := NewAccountTypeHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateAccountType))

	body := `{}`
	r := httptest.NewRequest(http.MethodPost, "/account-types", strings.NewReader(body))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleCreateAccountType_InvalidJSON(t *testing.T) {
	h := NewAccountTypeHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleCreateAccountType))

	r := httptest.NewRequest(http.MethodPost, "/account-types", strings.NewReader("not-json"))
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleDeleteAccountType_Success(t *testing.T) {
	mock := &mockQuerier{
		softDeleteAccountTypeByUser: func(_ context.Context, arg db.SoftDeleteAccountTypeByUserParams) error {
			return nil
		},
	}
	h := NewAccountTypeHandler(mock)
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleSoftDeleteAccountType))

	r := httptest.NewRequest(http.MethodDelete, "/account-types/3", nil)
	r.SetPathValue("id", "3")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestHandleDeleteAccountType_InvalidID(t *testing.T) {
	h := NewAccountTypeHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleSoftDeleteAccountType))

	r := httptest.NewRequest(http.MethodDelete, "/account-types/abc", nil)
	r.SetPathValue("id", "abc")
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestMapAccountTypeToResponse_SystemFlag(t *testing.T) {
	// Verify that IsSystem=1 maps to true and IsSystem=0 maps to false.
	system := mapAccountTypeToResponse(db.AccountType{ID: 1, Name: "checking", IsSystem: 1})
	if !system.IsSystem {
		t.Error("expected IsSystem true for system type")
	}

	custom := mapAccountTypeToResponse(db.AccountType{ID: 2, Name: "Savings", IsSystem: 0})
	if custom.IsSystem {
		t.Error("expected IsSystem false for custom type")
	}
}
