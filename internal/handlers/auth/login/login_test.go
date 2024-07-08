package login_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VanGoghDev/gophermart/internal/config"
	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/logger"
	"github.com/VanGoghDev/gophermart/internal/mocks"
	"github.com/VanGoghDev/gophermart/internal/router"
	"github.com/VanGoghDev/gophermart/internal/services/auth"
	"github.com/VanGoghDev/gophermart/internal/storage"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestNew(t *testing.T) {
	type args struct {
		login       string
		password    string
		contentType string
		body        string
		storageUser models.User
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
			name: "must return 200 status (invalid content type)",
			args: args{
				login:       "test",
				password:    "123",
				contentType: "application/json",
				body:        "{\"login\": \"test\", \"password\":\"123\"}",
				storageUser: models.User{
					Login:    "test",
					PassHash: make([]byte, 0),
				},
				storageErr: nil,
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
			name: "must return 401 status (invalid password)",
			args: args{
				login:       "test",
				password:    "1234",
				contentType: "application/json",
				body:        "{\"login\": \"test\", \"password\":\"123\"}",
				storageUser: models.User{
					Login:    "test",
					PassHash: make([]byte, 0),
				},
			},
			want: want{
				http.StatusUnauthorized,
			},
		},
		{
			name: "must return 500 status",
			args: args{
				login:       "test",
				contentType: "application/json",
				body:        "{\"login\": \"test\", \"password\":\"123\"}",
				storageErr:  storage.ErrNotFound,
			},
			want: want{
				http.StatusUnauthorized,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.New("dev")
			cfg, err := config.New()
			assert.Empty(t, err)

			token, err := auth.GenerateToken(tt.args.login, cfg.Secret, cfg.TokenExpires)
			assert.Empty(t, err)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockStorage(ctrl)

			if len(tt.args.password) != 0 {
				passHash, _ := bcrypt.GenerateFromPassword([]byte(tt.args.password), bcrypt.DefaultCost)
				tt.args.storageUser.PassHash = passHash
			}

			m.EXPECT().GetUser(gomock.Any(), gomock.Any()).
				Return(tt.args.storageUser, tt.args.storageErr).AnyTimes()

			r := router.New(log, m, cfg.Secret, cfg.TokenExpires)
			srv := httptest.NewServer(r)
			defer srv.Close()

			client := resty.New()

			resp, err := client.R().
				SetHeader("Content-Type", tt.args.contentType).
				SetHeader("Authorization", token).
				SetBody(tt.args.body).
				Post(fmt.Sprintf("%s/%s", srv.URL, "api/user/login"))

			assert.Empty(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			if resp.StatusCode() == http.StatusOK {
				assert.NotEmpty(t, resp.Header().Get("Authorization"))
			}
		})
	}
}
