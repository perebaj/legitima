package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"golang.org/x/exp/slog"
)

// Token is the token decoded from the Authorization header.
type Token struct {
	Email string `json:"email"`
}

// JWTSecretKey TODO: improve this
// JWTSecretKey is the secret key used to sign the JWT.
const JWTSecretKey = "secret"

func generateToken(email string) (string, error) {
	claims := jwt.MapClaims{
		email: email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecretKey))
}

// TokenFromHeader parses the token from the Authorization header and validates it.
func TokenFromHeader(r *http.Request) (*Token, error) {
	tokenHeader := r.Header.Get("Authorization")
	if tokenHeader == "" {
		slog.Debug("no authorization header")
		return nil, errors.New("no authorization header")
	}

	tokenParts := strings.Split(tokenHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		slog.Debug("invalid authorization header")
		return nil, errors.New("invalid authorization header")
	}

	token, err := jwt.Parse(tokenParts[1], func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			slog.Debug("invalid signing method")
			return nil, errors.New("invalid signing method")
		}
		return []byte(JWTSecretKey), nil
	})
	if err != nil {
		slog.Debug("error parsing token", "error", err.Error())
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		slog.Debug("invalid token", "token", token)
		return nil, errors.New("invalid token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		slog.Debug("invalid email claim")
		return nil, errors.New("invalid email claim")
	}
	var t Token
	t.Email = email

	return &t, nil
}
