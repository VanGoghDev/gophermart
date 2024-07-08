package router

import (
	"context"
	"log/slog"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/handlers/auth/login"
	"github.com/VanGoghDev/gophermart/internal/handlers/auth/register"
	"github.com/VanGoghDev/gophermart/internal/handlers/balance/getbalance"
	"github.com/VanGoghDev/gophermart/internal/handlers/balance/getwithdrawals"
	"github.com/VanGoghDev/gophermart/internal/handlers/balance/postwithdraw"
	"github.com/VanGoghDev/gophermart/internal/handlers/orders/getorders"
	"github.com/VanGoghDev/gophermart/internal/handlers/orders/postorders"
	"github.com/VanGoghDev/gophermart/internal/middleware/auth"
	"github.com/go-chi/chi"
)

type Storage interface {
	RegisterUser(ctx context.Context, login string, password string) (string, error)
	GetUser(ctx context.Context, userLogin string) (models.User, error)

	GetOrder(ctx context.Context, number string) (models.Order, error)
	GetOrders(ctx context.Context, userLogin string) ([]models.Order, error)
	SaveOrder(ctx context.Context, number string, userLogin string, status models.OrderStatus) error

	GetBalance(ctx context.Context, userLogin string) (models.Balance, error)

	GetWithdrawals(ctx context.Context, userLogin string) ([]models.Withdrawal, error)
	SaveWithdrawal(ctx context.Context, userLogin string, orderNum string, sum int64) error
}

func New(log *slog.Logger, storage Storage, tokenSecret string, tokenExpires time.Duration) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", register.New(log, storage, tokenSecret, tokenExpires))
		r.Post("/login", login.New(log, storage, tokenSecret, tokenExpires))

		r.Group(func(r chi.Router) {
			r.Use(auth.New(log, tokenSecret))
			r.Post("/orders", postorders.New(log, storage, storage))
			r.Get("/orders", getorders.New(log, storage))

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", getbalance.New(log, storage))

				r.Post("/withdraw", postwithdraw.New(log, storage, storage, storage))
			})
			r.Get("/withdrawals", getwithdrawals.New(log, storage))
		})
	})
	return r
}
