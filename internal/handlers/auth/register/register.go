package register

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	hauth "github.com/VanGoghDev/gophermart/internal/handlers/auth"
	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
	"github.com/VanGoghDev/gophermart/internal/services/auth"
	"github.com/VanGoghDev/gophermart/internal/storage"
)

type Register interface {
	RegisterUser(ctx context.Context, login string, password string) (lgn string, err error)
}

func New(log *slog.Logger, s Register, secret string, tokenExpires time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, req := hauth.ValidateUserRequest(r.Context(), log, r)
		if code >= http.StatusBadRequest {
			w.WriteHeader(code)
		}
		login, err := s.RegisterUser(r.Context(), req.Login, req.Password)
		if err != nil {
			// 409 логин уже занят.
			if errors.Is(err, storage.ErrAlreadyExists) {
				log.ErrorContext(r.Context(), "login already exists", sl.Err(err))
				w.WriteHeader(http.StatusConflict)
				return
			}

			// 500 внутренняя ошибка сервера.
			log.ErrorContext(r.Context(), "failed to register user", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		token, err := auth.GrantToken(login, secret, tokenExpires)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to grant token", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Authorization", token)
	}
}
