package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := executeRequest(http.HandlerFunc(HandleHealth), r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if body != "Flizix service OK" {
		t.Errorf("unexpected body: %q", body)
	}
}
