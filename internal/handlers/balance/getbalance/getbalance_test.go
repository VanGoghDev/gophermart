package getbalance_test

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
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		login       string
		contentType string
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
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "must return 500 status",
			args: args{
				login:       "test",
				contentType: "application/json",
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

			token, err := auth.GenerateToken(tt.args.login, cfg.Secret, cfg.TokenExpires)
			assert.Empty(t, err)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockStorage(ctrl)

			m.EXPECT().GetBalance(gomock.Any(), gomock.Any()).
				Return(models.Balance{}, tt.args.storageErr).AnyTimes()

			r := router.New(log, m, cfg.Secret, cfg.TokenExpires)
			srv := httptest.NewServer(r)
			defer srv.Close()

			client := resty.New()

			resp, err := client.R().
				SetHeader("Content-Type", tt.args.contentType).
				SetHeader("Authorization", token).
				Get(fmt.Sprintf("%s/%s", srv.URL, "api/user/balance"))

			assert.Empty(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
		})
	}
}
