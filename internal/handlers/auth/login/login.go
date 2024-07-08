package login

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
	"github.com/VanGoghDev/gophermart/internal/services/auth"
	"github.com/VanGoghDev/gophermart/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

type Request struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserProvider interface {
	GetUser(ctx context.Context, login string) (models.User, error)
}

func New(log *slog.Logger, s UserProvider, secret string, tokenExpires time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.login.New"
		log = log.With("op", op)

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		req := &Request{}

		dec := json.NewDecoder(r.Body)

		err := dec.Decode(req)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to decode request body", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := validator.New().Struct(req); err != nil {
			log.ErrorContext(r.Context(), "", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// 401 неверная пара логин/пароль.
		user, err := s.GetUser(r.Context(), req.Login)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				w.WriteHeader(http.StatusUnauthorized) // нужна ли эта обработка?
				return
			}
			log.ErrorContext(r.Context(), "failed to get user", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(req.Password)); err != nil {
			log.InfoContext(r.Context(), "invalid credentials", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// выписать токен
		token, err := auth.GrantToken(user.Login, secret, tokenExpires)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to generate auth token", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Authorization", token)
	}
}
