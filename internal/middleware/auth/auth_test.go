package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VanGoghDev/gophermart/internal/logger"
	"github.com/VanGoghDev/gophermart/internal/middleware/auth"
	sauth "github.com/VanGoghDev/gophermart/internal/services/auth"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

type fakeClaims struct {
	jwt.RegisteredClaims
	UserLogin int
}

func TestNew(t *testing.T) {
	type args struct {
		login                string
		clientSecret         string
		tokenExpires         time.Duration
		brokenUserLoginClaim bool
	}
	type want struct {
		statusCode int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid token returns 200",
			args: args{
				login:        "test",
				clientSecret: "secret",
				tokenExpires: time.Second * 5,
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "invalid secret returns 401",
			args: args{
				login:        "test",
				clientSecret: "secret2",
				tokenExpires: time.Second * 5,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "expired token returns 401",
			args: args{
				login:        "test",
				clientSecret: "secret",
				tokenExpires: time.Millisecond * 5,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "empty authorization header returns 401",
			args: args{
				login:        "test",
				clientSecret: "",
				tokenExpires: time.Second * 5,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "broken claim user login returns 401",
			args: args{
				login:                "test",
				clientSecret:         "secret",
				tokenExpires:         time.Second * 5,
				brokenUserLoginClaim: true,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
	}

	log := logger.New("dev")
	serverSecret := "secret"

	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.New(log, serverSecret))
			r.Get("/orders", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	client := resty.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := sauth.GenerateToken(tt.args.login, tt.args.clientSecret, tt.args.tokenExpires)
			if tt.args.brokenUserLoginClaim {
				tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, fakeClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(tt.args.tokenExpires)),
					},
					UserLogin: 1,
				})
				token, _ = tkn.SignedString([]byte(tt.args.clientSecret))
			}

			assert.Empty(t, err)
			if tt.args.clientSecret == "" {
				token = ""
			}
			resp, _ := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Authorization", token).
				Get(fmt.Sprintf("%s/%s", srv.URL, "api/user/orders"))

			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
		})
	}
}
