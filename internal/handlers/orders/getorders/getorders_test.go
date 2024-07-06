package getorders_test

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
)

func TestNew(t *testing.T) {
	type args struct {
		login           string
		contentType     string
		storageGetOrder []models.Order
		storageGetErr   error
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
			name: "must return 204 status",
			args: args{
				login:           "test",
				contentType:     "text/plain",
				storageGetOrder: make([]models.Order, 0),
				storageGetErr:   nil,
			},
			want: want{
				http.StatusOK,
			},
		},
		{
			name: "must return 204 status",
			args: args{
				login:           "test",
				contentType:     "text/plain",
				storageGetOrder: make([]models.Order, 0),
				storageGetErr:   storage.ErrNotFound,
			},
			want: want{
				http.StatusNoContent,
			},
		},
		{
			name: "must return 401 status",
			args: args{
				login:           "",
				contentType:     "text/plain",
				storageGetOrder: make([]models.Order, 0),
				storageGetErr:   nil,
			},
			want: want{
				http.StatusUnauthorized,
			},
		},
		{
			name: "must return 500 status",
			args: args{
				login:           "test",
				contentType:     "text/plain",
				storageGetOrder: make([]models.Order, 0),
				storageGetErr:   errors.New("storage error"),
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

			m.EXPECT().GetOrders(gomock.Any(), gomock.Any()).
				Return(tt.args.storageGetOrder, tt.args.storageGetErr).AnyTimes()

			r := router.New(log, m, cfg.Secret, cfg.TokenExpires)
			srv := httptest.NewServer(r)
			defer srv.Close()

			client := resty.New()

			resp, err := client.R().
				SetHeader("Content-Type", tt.args.contentType).
				SetHeader("Authorization", token).
				Get(fmt.Sprintf("%s/%s", srv.URL, "api/user/orders"))

			assert.Empty(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
		})
	}
}
