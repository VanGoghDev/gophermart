package getorders

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

type OrderProvider interface {
	GetOrders(ctx context.Context, userLogin string) ([]models.Order, error)
}

func New(log *slog.Logger, s OrderProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userLogin, err := auth.GetLogin(r)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to get userLogin from context: %w")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		orders, err := s.GetOrders(r.Context(), userLogin)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			log.ErrorContext(r.Context(), "failed to get orders from storage: %w", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		err = enc.Encode(orders)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to encode orders json: %w")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
