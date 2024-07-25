package getbalance

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
	"github.com/VanGoghDev/gophermart/internal/middleware/auth"
	"github.com/VanGoghDev/gophermart/internal/storage"
)

type BalanceProvider interface {
	GetBalance(ctx context.Context, userLogin string) (models.Balance, error)
}

func New(log *slog.Logger, s BalanceProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userLogin, err := auth.GetLogin(r)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to fetch user login from context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		balance, err := s.GetBalance(r.Context(), userLogin)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			log.ErrorContext(r.Context(), "failed to get balance from storage: %w", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		err = enc.Encode(balance)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to encode response on getbalance: %w", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
