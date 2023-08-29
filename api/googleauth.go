// Package api contains the authentication endpoints.
package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/birdie-ai/golibs/slog"
	"github.com/birdie-ai/legitima"
	"golang.org/x/oauth2"
)

// Auth endpoints
const (
	loginURL    = "/login"
	callbackURL = "/callback"
)

// Storage interface take care of functionalities needed by the auth endpoints.
type Storage interface {
	SaveUser(gUsr legitima.GoogleUser) error
	UserByEmail(email string) (*legitima.User, error)
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

	var usr legitima.GoogleUser
	err = d.Decode(&usr)
	if err != nil {
		slog.Error("error decoding user info", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	err = storage.SaveUser(usr)
	if err != nil {
		slog.Error("error saving user", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	tokenString, err := GenerateToken(usr.Email)
	if err != nil {
		slog.Error("error generating token", "error", err.Error())
		sendErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    "Bearer " + tokenString,
		HttpOnly: true,
		Path:     profileURL,
		Secure:   true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, profileURL, http.StatusSeeOther)
}

// func profile(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	token, err := tokenFromHeader(ctx, r)
// 	if err != nil {
// 		slog.Warn("user not authenticated", "error", err.Error())
// 		sendErr(ctx, w, err, http.StatusUnauthorized)
// 		return
// 	}
// 	slog.Info("user authenticated", "token", token)
// 	// usr, err := storage.UserByEmail(token)
// 	// if err != nil {
// 	// 	slog.Error("error getting user", "error", err.Error())
// 	// 	sendErr(ctx, w, err, http.StatusInternalServerError)
// 	// 	return
// 	// }

// 	// usrByte, err := json.Marshal(usr)
// 	// if err != nil {
// 	// 	slog.Error("error marshaling user", "error", err.Error())
// 	// 	sendErr(ctx, w, err, http.StatusInternalServerError)
// 	// 	return
// 	// }

// 	sendJSON(ctx, w, http.StatusOK, token)
// }
