package getbalance

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/middleware/auth"
)

type BalanceProvider interface {
	GetBalance(ctx context.Context, userLogin string) (models.Balance, error)
}

func New(log *slog.Logger, s BalanceProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.balance.getbalance.New"
		log = log.With("op", op)

		userLogin, ok := r.Context().Value(auth.UserLoginKey).(string)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if userLogin == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		balance, err := s.GetBalance(r.Context(), userLogin)
		if err != nil {
			log.ErrorContext(r.Context(), "%s: %w", op, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		err = enc.Encode(balance)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
