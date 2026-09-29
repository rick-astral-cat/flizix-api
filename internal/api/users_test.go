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

func TestHandleCreateUser_Success(t *testing.T) {
	mock := &mockQuerier{
		createUserWithPasskey: func(_ context.Context, arg db.CreateUserWithPasskeyParams) (db.User, error) {
			return db.User{ID: 1, Name: arg.Name, Email: arg.Email}, nil
		},
	}
	h := NewUserHandler(mock)

	body := `{"name":"Rick","email":"rick@example.com","passkey_id":"pk123"}`
	r := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	rr := executeRequest(http.HandlerFunc(h.HandleCreateUser), r)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Name != "Rick" {
		t.Errorf("expected name Rick, got %s", resp.Name)
	}
}

func TestHandleCreateUser_InvalidJSON(t *testing.T) {
	h := NewUserHandler(&mockQuerier{})

	r := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("not-json"))
	rr := executeRequest(http.HandlerFunc(h.HandleCreateUser), r)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleGetProfile_Success(t *testing.T) {
	mock := &mockQuerier{
		getUserById: func(_ context.Context, id int64) (db.User, error) {
			return db.User{ID: id, Name: "Rick", Email: sql.NullString{String: "rick@example.com", Valid: true}}, nil
		},
	}
	h := NewUserHandler(mock)

	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	r = authenticatedRequest(r, 1)

	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleGetProfile))
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.ID != 1 {
		t.Errorf("expected ID 1, got %d", resp.ID)
	}
}

func TestHandleGetProfile_Unauthorized(t *testing.T) {
	h := NewUserHandler(&mockQuerier{})
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(h.HandleGetProfile))

	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
