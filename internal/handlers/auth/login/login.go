package login

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	hauth "github.com/VanGoghDev/gophermart/internal/handlers/auth"
	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
	"github.com/VanGoghDev/gophermart/internal/services/auth"
	"github.com/VanGoghDev/gophermart/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type UserProvider interface {
	GetUser(ctx context.Context, login string) (models.User, error)
}

func New(log *slog.Logger, s UserProvider, secret string, tokenExpires time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, req := hauth.ValidateUserRequest(r.Context(), log, r)
		if code >= http.StatusBadRequest {
			w.WriteHeader(code)
		}

		// 401 неверная пара логин/пароль.
		user, err := s.GetUser(r.Context(), req.Login)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				w.WriteHeader(http.StatusUnauthorized)
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
