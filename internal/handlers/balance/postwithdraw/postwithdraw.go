package postwithdraw

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

type UserProvider interface {
	GetUser(ctx context.Context, login string) (models.User, error)
}

type OrderProvider interface {
	GetOrder(ctx context.Context, number string) (models.Order, error)
}

type WithdrawalSaver interface {
	SaveWithdrawal(ctx context.Context, userLogin string, orderNum string, sum float64) error
}

type Request struct {
	OrderNum string  `json:"order"`
	Sum      float64 `json:"sum"`
}

func New(log *slog.Logger, s WithdrawalSaver, su UserProvider, so OrderProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userLogin, err := auth.GetLogin(r)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to fetch user login from context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		req := &Request{}
		dec := json.NewDecoder(r.Body)
		err = dec.Decode(req)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to decode withrawal json: %w")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = s.SaveWithdrawal(r.Context(), userLogin, req.OrderNum, req.Sum)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				w.WriteHeader(http.StatusOK)
				return
			}
			if errors.Is(err, storage.ErrNotEnoughFunds) {
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			log.ErrorContext(r.Context(), "failed to save withdrawal", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
