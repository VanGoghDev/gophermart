package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/VanGoghDev/gophermart/internal/services/auth"
	"github.com/go-chi/chi/middleware"
)

type contextKey string

const UserLoginKey contextKey = "user-login"

func New(log *slog.Logger, secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.auth.New"

			// взять контекст чтобы потом в него записать инфо о юзере
			ctx := r.Context()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// взять хэдер аутентификации
			token := r.Header.Get("Authorization")
			if token != "" {
				authorized, err := auth.IsAuthorized(token, secret)
				if err != nil || !authorized {
					log.ErrorContext(r.Context(), "%s:%w", op, err)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if authorized {
					// достать claims
					login, err := auth.ExtractLoginFromToken(token, secret)
					if err != nil {
						log.ErrorContext(r.Context(), "%s:%w", op, err)
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if login == "" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					// валиден - все ок, передать инфу о пользователе в контекст
					ctx = context.WithValue(ctx, UserLoginKey, login) // проверить что оно работает мб юнит тесты?
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
