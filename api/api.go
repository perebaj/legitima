// Package api contains the API useful functions and types.
package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/birdie-ai/golibs/slog"
)

// ErrorResponse is the response sent to the client in case of error.
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func sendErr(ctx context.Context, res http.ResponseWriter, err error, statusCode int) {
	log := slog.FromCtx(ctx)
	switch {
	case statusCode >= 500:
		log.Error("server side error", "error", err, "status", statusCode)
	case statusCode >= 400:
		log.Warn("client side error", "error", err, "status", statusCode)
	}

	res.WriteHeader(statusCode)

	encoder := json.NewEncoder(res)
	var response ErrorResponse
	response.Error.Message = err.Error()
	err = encoder.Encode(response)
	if err != nil {
		log.Error("failed to send json error message to client", "error", err)
	}
}

func sendJSON(ctx context.Context, w http.ResponseWriter, statusCode int, body interface{}) {
	const jsonContentType = "application/json; charset=utf-8"

	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.FromCtx(ctx).Error("Unable to encode body as JSON", "error", err)
	}
}
