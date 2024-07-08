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
	SaveWithdrawal(ctx context.Context, userLogin string, orderNum string, sum int64) error
}

type Request struct {
	OrderNum string `json:"order"`
	Sum      int64  `json:"sum"`
}

func New(log *slog.Logger, s WithdrawalSaver, su UserProvider, so OrderProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.balance.postwithdraw.New"
		log = log.With("op", op)

		userLogin, ok := r.Context().Value(auth.UserLoginKey).(string)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		req := &Request{}
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := su.GetUser(r.Context(), userLogin)
		if err != nil {
			log.ErrorContext(r.Context(), "", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.Balance < req.Sum {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

		_, err = so.GetOrder(r.Context(), req.OrderNum)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			log.ErrorContext(r.Context(), "", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = s.SaveWithdrawal(r.Context(), userLogin, req.OrderNum, req.Sum)
		if err != nil {
			log.ErrorContext(r.Context(), "", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
