package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/birdie-ai/legitima/api"
)

// TestToken is responsible for testing the token generation and validation in the same flow.
func TestToken(t *testing.T) {
	email := "jj@gmail.com"
	token, err := api.GenerateToken(email)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	r := httptest.NewRequest(http.MethodGet, "/profile", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	tokenFromHeader, err := api.TokenFromHeader(r)
	if err != nil {
		t.Fatalf("failed to get token from header: %v", err)
	}
	if tokenFromHeader.Email != email {
		t.Fatalf("expected email %s, got %s", email, tokenFromHeader.Email)
	}
}
