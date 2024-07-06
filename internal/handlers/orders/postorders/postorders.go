package postorders

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/middleware/auth"
	"github.com/VanGoghDev/gophermart/internal/storage"
)

type OrderProvider interface {
	GetOrder(ctx context.Context, number string) (models.Order, error)
}

type OrdersSaver interface {
	SaveOrder(ctx context.Context, number string, userLogin string, status models.OrderStatus) error
}

func New(log *slog.Logger, s OrdersSaver, sp OrderProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.orders.postorders.New"

		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userLogin, ok := r.Context().Value(auth.UserLoginKey).(string)
		if !ok {
			log.ErrorContext(r.Context(), "failed to fetch user login from context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if userLogin == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		bNum, err := io.ReadAll(r.Body)
		if err != nil {
			log.ErrorContext(r.Context(), "%s: %w", op, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// проверка на валидность номера заказа

		// 422 — неверный формат номера заказа;

		// тут пока такая заглушка. Переделать на нормальную
		// валидацию через алгоритм Луна. Юнит тест тоже.
		if string(bNum) == "a" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		order, err := sp.GetOrder(r.Context(), string(bNum))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if order.UserLogin != userLogin && order.UserLogin != "" {
			w.WriteHeader(http.StatusConflict)
			return
		}

		err = s.SaveOrder(r.Context(), string(bNum), userLogin, models.New)
		if err != nil {
			if errors.Is(err, storage.ErrAlreadyExists) {
				log.InfoContext(r.Context(), "user already has this order")
				w.WriteHeader(http.StatusOK)
				return
			}
			log.ErrorContext(r.Context(), "%s: %w", op, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusAccepted)
	}
}