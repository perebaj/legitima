package api_test

import (
	"net/http/httptest"
	"testing"

	"github.com/birdie-ai/legitima/api"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type mockStorage interface {
	SaveUser(gUsr api.GoogleUser) error
}

func TestAuth_Callback_EmptyCode(t *testing.T) {
	mStorage := new(mockStorage)
	googleOAuthConfig := oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
	}

	h := api.CallbackHandler(&googleOAuthConfig, *mStorage)
	req := httptest.NewRequest("GET", "/callback", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
