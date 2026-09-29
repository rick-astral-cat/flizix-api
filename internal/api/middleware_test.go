package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestMiddleware() *MiddlewareHandler {
	return &MiddlewareHandler{
		Auth: newTestAuth(),
	}
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestJWTMiddleware_NoToken(t *testing.T) {
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(dummyHandler))

	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(dummyHandler))

	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid.token.here"})
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTMiddleware_ValidCookie(t *testing.T) {
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(dummyHandler))

	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	r = authenticatedRequest(r, 1)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestJWTMiddleware_ValidBearerHeader(t *testing.T) {
	mid := newTestMiddleware()
	handler := mid.JWTMiddleware(http.HandlerFunc(dummyHandler))

	auth := newTestAuth()
	token, _ := auth.GenerateToken(1)

	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rr := executeRequest(handler, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
