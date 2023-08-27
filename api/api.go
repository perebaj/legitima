// Package api implements the HTTP API provided by the ingester service.
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
