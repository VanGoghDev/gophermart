package router

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/handlers/auth/login"
	"github.com/VanGoghDev/gophermart/internal/handlers/auth/register"
	"github.com/VanGoghDev/gophermart/internal/handlers/orders/getorders"
	"github.com/VanGoghDev/gophermart/internal/middleware/auth"
	"github.com/go-chi/chi"
)

type Storage interface {
	RegisterUser(ctx context.Context, login string, password string) (string, error)
	GetUser(ctx context.Context, userLogin string, password string) (models.User, error)
}

func New(log *slog.Logger, storage Storage, tokenSecret string, tokenExpires time.Duration) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", register.New(log, storage, tokenSecret, tokenExpires))
		r.Post("/login", login.New(log, storage, tokenSecret, tokenExpires))

		r.Group(func(r chi.Router) {
			r.Use(auth.New(log, tokenSecret))
			r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {})
			r.Get("/orders", getorders.New())

			r.Get("/balance", func(w http.ResponseWriter, r *http.Request) {})

			r.Post("/withdraw", func(w http.ResponseWriter, r *http.Request) {})
			r.Get("/withdrawals", func(w http.ResponseWriter, r *http.Request) {})
		})
	})
	return r
}
