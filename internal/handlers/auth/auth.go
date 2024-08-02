package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
	"gopkg.in/go-playground/validator.v9"
)

type Request struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func ValidateUserRequest(ctx context.Context, log *slog.Logger, r *http.Request) (statusCode int, req *Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return http.StatusBadRequest, req
	}
	req = &Request{}
	dec := json.NewDecoder(r.Body)

	err := dec.Decode(req)
	if err != nil {
		log.ErrorContext(ctx, "%w: failed to unmarshal json", sl.Err(err))
		return http.StatusBadRequest, req
	}

	if err := validator.New().Struct(req); err != nil {
		return http.StatusBadRequest, req
	}

	return http.StatusOK, req
}
