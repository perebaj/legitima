// Package api contains the authentication endpoints.
package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/birdie-ai/golibs/slog"
	"github.com/birdie-ai/legitima/mysql"
	"golang.org/x/oauth2"
)

// Auth endpoints
const (
	loginURL    = "/login"
	callbackURL = "/callback"
)

// Storage interface take care of functionalities needed by the auth endpoints.
type Storage interface {
	SaveUser(gUsr GoogleUser) error
	UserByEmail(*Token) (mysql.User, error)
}

// SetupAuth sets up the authentication endpoints.
func SetupAuth(mux *http.ServeMux, googleOAuthConfig *oauth2.Config, storage Storage) {
	mux.Handle(loginURL, LoginHandler(googleOAuthConfig))
	mux.Handle(callbackURL, CallbackHandler(googleOAuthConfig, storage))
}

// LoginHandler handles the login endpoint.
func LoginHandler(googleOAuthConfig *oauth2.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			login(w, r, googleOAuthConfig)
		default:
			sendErr(r.Context(), w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		}
	})
}

// CallbackHandler handles the callback from Google.
func CallbackHandler(googleOAuthConfig *oauth2.Config, storage Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			callback(w, r, googleOAuthConfig, storage)
		default:
			sendErr(r.Context(), w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		}
	})
}

// TODO(JOJO): randomize this
var randState = "random"

func login(w http.ResponseWriter, r *http.Request, googleOAuthConfig *oauth2.Config) {
	ctx := r.Context()
	log := slog.FromCtx(ctx)

	url := googleOAuthConfig.AuthCodeURL(randState, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
	log.Info("login request received")
}

func callback(w http.ResponseWriter, r *http.Request, googleOAuthConfig *oauth2.Config, storage Storage) {
	state := r.FormValue("state")
	if state == "" {
		sendErr(r.Context(), w, errors.New("missing state"), http.StatusBadRequest)
		return
	}
	if state != randState {
		sendErr(r.Context(), w, errors.New("invalid state"), http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		sendErr(r.Context(), w, errors.New("missing code"), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	token, err := googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		slog.Error("error exchanging token", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	client := googleOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		slog.Error("error getting user info", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	d := json.NewDecoder(resp.Body)

	var usr GoogleUser
	err = d.Decode(&usr)
	if err != nil {
		slog.Error("error decoding user info", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("user info", "user_data", usr)

	err = storage.SaveUser(usr)
	if err != nil {
		slog.Error("error saving user", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	tokenString, err := generateToken(usr.Email)
	if err != nil {
		slog.Error("error generating token", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}
	slog.Info("token generated", "token", tokenString)
	_, _ = w.Write([]byte(tokenString))
}

// GoogleUser represents the user data returned by Google.
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	VerifiedEmail bool   `json:"verified_email"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func profile(w http.ResponseWriter, r *http.Request, storage Storage) {
	ctx := r.Context()
	token, err := tokenFromHeader(ctx, r)
	if err != nil {
		slog.Warn("user not authenticated", "error", err.Error())
		sendErr(ctx, w, err, http.StatusUnauthorized)
		return
	}

	usr, err := storage.UserByEmail(token)
	if err != nil {
		slog.Error("error getting user", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	usrByte, err := json.Marshal(usr)
	if err != nil {
		slog.Error("error marshaling user", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	sendJSON(ctx, w, http.StatusOK, usrByte)
}
