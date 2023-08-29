package api

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/birdie-ai/golibs/slog"
)

const profileURL = "/profile"

// SetupProfile sets up the profile page.
func SetupProfile(mux *http.ServeMux, storage Storage) {
	mux.Handle(profileURL, ProfileHandler(storage))
}

// ProfileHandler handles the profile page.
func ProfileHandler(storage Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			profile(w, r, storage)
		default:
			sendErr(r.Context(), w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		}
	})
}

func profile(w http.ResponseWriter, r *http.Request, storage Storage) {
	authCookie, err := r.Cookie("Authorization")
	if err != nil {
		slog.Error("failed to get token from cookie", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	receivedToken := authCookie.Value
	r.Header.Set("Authorization", receivedToken)
	token, err := TokenFromHeader(r)
	if err != nil {
		slog.Error("invalid token", "error", err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	usr, err := storage.UserByEmail(token.Email)
	if err != nil {
		slog.Error("failed to get user", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		slog.Error("failed to parse template", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, usr)
	if err != nil {
		slog.Error("failed to execute template", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
