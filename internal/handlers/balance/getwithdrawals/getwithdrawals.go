package getwithdrawals

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

type WithdrawalsProvider interface {
	GetWithdrawals(ctx context.Context, userLogin string) ([]models.Withdrawal, error)
}

func New(log *slog.Logger, s WithdrawalsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userLogin, ok := r.Context().Value(auth.UserLoginKey).(string)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		withdrawals, err := s.GetWithdrawals(r.Context(), userLogin)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			log.ErrorContext(r.Context(), "", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err = enc.Encode(withdrawals)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to encode withdrawals json", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
