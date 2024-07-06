package getorders

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/middleware/auth"
	"github.com/VanGoghDev/gophermart/internal/storage"
)

type OrderProvider interface {
	GetOrders(ctx context.Context, userLogin string) ([]models.Order, error)
}

func New(log *slog.Logger, s OrderProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.orders.getorders.New"

		w.Header().Set("Content-Type", "application/json")

		userLogin, ok := r.Context().Value(auth.UserLoginKey).(string)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if userLogin == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		orders, err := s.GetOrders(r.Context(), userLogin)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			log.ErrorContext(r.Context(), "%s: %w", op, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(orders)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}
