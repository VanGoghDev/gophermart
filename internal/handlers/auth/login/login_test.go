package login_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VanGoghDev/gophermart/internal/logger"
	"github.com/VanGoghDev/gophermart/internal/router"
	"github.com/VanGoghDev/gophermart/internal/storage"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// тест на 401

// тест на 500

// тест на токен

type args struct {
	body string
}
type want struct {
	statusCode int
	tokenEmpty bool
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty body",
			args: args{
				body: "",
			},
			want: want{
				statusCode: 400,
				tokenEmpty: true,
			},
		},
		{
			name: "invalid body",
			args: args{
				body: "{\"login\": \"\", \"password\":\"\"}",
			},
			want: want{
				statusCode: 400,
				tokenEmpty: true,
			},
		},
		{
			name: "invalid login",
			args: args{
				body: "{\"login\": \"\", \"password\":\"123\"}",
			},
			want: want{
				statusCode: 400,
				tokenEmpty: true,
			},
		},
		{
			name: "invalid password",
			args: args{
				body: "{\"login\": \"test\", \"password\":\"\"}",
			},
			want: want{
				statusCode: 400,
				tokenEmpty: true,
			},
		},
		// {
		// 	name: "valid body",
		// 	args: args{
		// 		body: "{\"login\": \"test\", \"password\":\"123\"}",
		// 	},
		// 	want: want{
		// 		statusCode: 200,
		// 		tokenEmpty: false,
		// 	},
		// },
	}
	log := logger.New("dev")
	secret := "secret"
	tokenExpires := time.Hour * 3

	s, _ := storage.New(context.Background(), "")
	r := router.New(log, s, secret, tokenExpires)
	srv := httptest.NewServer(r)
	defer srv.Close()

	client := resty.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetBody(tt.args.body).
				Post(fmt.Sprintf("%s/%s", srv.URL, "api/user/login"))

			assert.Empty(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())

			// токен не всегда нужно проверять
			if tt.want.tokenEmpty {
				assert.Empty(t, resp.Header().Get("Authorization"))
			} else {
				assert.NotEmpty(t, resp.Header().Get("Authorization"))
			}
		})
	}
}
