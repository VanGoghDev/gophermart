package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
)

type Storage interface {
}

func New(storage Storage, log *slog.Logger) chi.Router {
	r := chi.NewRouter()

	// todo auth middleware

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {})
		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {})

		r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {})
		r.Get("/orders", func(w http.ResponseWriter, r *http.Request) {})

		r.Get("/balance", func(w http.ResponseWriter, r *http.Request) {})

		r.Post("/withdraw", func(w http.ResponseWriter, r *http.Request) {})
		r.Get("/withdrawals", func(w http.ResponseWriter, r *http.Request) {})
	})
	return r
}
