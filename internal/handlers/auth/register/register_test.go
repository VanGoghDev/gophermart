package register_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VanGoghDev/gophermart/internal/config"
	"github.com/VanGoghDev/gophermart/internal/logger"
	"github.com/VanGoghDev/gophermart/internal/mocks"
	"github.com/VanGoghDev/gophermart/internal/router"
	"github.com/VanGoghDev/gophermart/internal/storage"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		login       string
		contentType string
		body        string
		storageErr  error
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
			name: "must return 200 status",
			args: args{
				login:       "test",
				contentType: "application/json",
				body:        "{\"login\": \"123\", \"password\":\"123\"}",
			},
			want: want{
				http.StatusOK,
			},
		},
		{
			name: "must return 400 status (invalid content type)",
			args: args{
				login:       "test",
				contentType: "text/plain",
				storageErr:  errors.New("storage error"),
			},
			want: want{
				http.StatusBadRequest,
			},
		},
		{
			name: "must return 400 status (empty body)",
			args: args{
				login:       "test",
				contentType: "application/json",
				storageErr:  errors.New("storage error"),
			},
			want: want{
				http.StatusBadRequest,
			},
		},
		{
			name: "must return 400 status (invalid body)",
			args: args{
				login:       "test",
				contentType: "application/json",
				body:        "{\"login\": \"\", \"password\":\"\"}",
			},
			want: want{
				http.StatusBadRequest,
			},
		},
		{
			name: "must return 400 status (invalid login)",
			args: args{
				login:       "test",
				contentType: "application/json",
				body:        "{\"login\": \"\", \"password\":\"123\"}",
			},
			want: want{
				http.StatusBadRequest,
			},
		},
		{
			name: "must return 400 status (invalid password)",
			args: args{
				login:       "test",
				contentType: "application/json",
				body:        "{\"login\": \"test\", \"password\":\"\"}",
			},
			want: want{
				http.StatusBadRequest,
			},
		},
		{
			name: "must return 409 status",
			args: args{
				login:       "test",
				contentType: "application/json",
				body:        "{\"login\": \"test\", \"password\":\"123\"}",
				storageErr:  storage.ErrAlreadyExists,
			},
			want: want{
				http.StatusConflict,
			},
		},
		{
			name: "must return 500 status",
			args: args{
				login:       "test",
				contentType: "application/json",
				body:        "{\"login\": \"test\", \"password\":\"123\"}",
				storageErr:  errors.New("storage error"),
			},
			want: want{
				http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.New("dev")
			cfg, err := config.New()
			assert.Empty(t, err)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockStorage(ctrl)

			m.EXPECT().RegisterUser(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.args.login, tt.args.storageErr).AnyTimes()

			r := router.New(log, m, cfg.Secret, cfg.TokenExpires)
			srv := httptest.NewServer(r)
			defer srv.Close()

			client := resty.New()

			resp, err := client.R().
				SetHeader("Content-Type", tt.args.contentType).
				SetBody(tt.args.body).
				Post(fmt.Sprintf("%s/%s", srv.URL, "api/user/register"))

			assert.Empty(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
		})
	}
}
