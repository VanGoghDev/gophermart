package register

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"gopkg.in/go-playground/validator.v9"

	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
	"github.com/VanGoghDev/gophermart/internal/services/auth"
	"github.com/VanGoghDev/gophermart/internal/storage"
)

type Request struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Register interface {
	RegisterUser(ctx context.Context, login string, password string) (lgn string, err error)
}

func New(log *slog.Logger, s Register, secret string, tokenExpires time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.register.New"
		log = log.With("op", op)

		req := &Request{}

		// 400 неверный формат запроса.
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(req); err != nil {
			log.WarnContext(r.Context(), "", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := validator.New().Struct(req); err != nil {
			log.ErrorContext(r.Context(), "", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		login, err := s.RegisterUser(r.Context(), req.Login, req.Password)
		if err != nil {
			// 409 логин уже занят.
			if errors.Is(err, storage.ErrLoginAlreadyExists) {
				log.ErrorContext(r.Context(), "", sl.Err(err))
				w.WriteHeader(http.StatusConflict)
				return
			}

			// 500 внутренняя ошибка сервера.
			log.ErrorContext(r.Context(), "", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		token, err := auth.GrantToken(login, secret, tokenExpires)
		if err != nil {
			log.ErrorContext(r.Context(), "", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// в случае успешной регистрации записать токен в хэдер.
		w.Header().Set("Authorization", token)
	}
}
