package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/VanGoghDev/gophermart/internal/services/auth"
	"github.com/go-chi/chi/middleware"
)

type contextKey int

const (
	KeyUserLogin contextKey = iota
)

func GetLogin(r *http.Request) (login string, err error) {
	userLogin, ok := r.Context().Value(KeyUserLogin).(string)
	if !ok {
		return "", fmt.Errorf("unable to cast given context value to string")
	}
	return userLogin, nil
}

func New(log *slog.Logger, secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// взять контекст чтобы потом в него записать инфо о юзере
			ctx := r.Context()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// взять хэдер аутентификации
			token := r.Header.Get("Authorization")
			if token != "" {
				authorized, err := auth.IsAuthorized(token, secret)
				if err != nil || !authorized {
					log.ErrorContext(r.Context(), "authorization failed: %w", err)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if authorized {
					// достать claims
					login, err := auth.ExtractLoginFromToken(token, secret)
					if err != nil {
						log.ErrorContext(r.Context(), "failed to get claims from token: %w", err)
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if login == "" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					ctx = context.WithValue(ctx, KeyUserLogin, login)
				}
			} else {
				http.Error(w, "Authorization header is empty", http.StatusUnauthorized)
				return
			}

			// вызываем следующий обработчик
			next.ServeHTTP(ww, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
