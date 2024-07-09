package postwithdraw_test

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
		contentType              string
		body                     string
		storageUser              models.User
		storageSaveWithdrawalErr error
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		// TODO: Add test cases.
		// 	200 — успешная обработка запроса;
		{
			name: "must return 200 status",
			args: args{
				contentType: "application/json",
				body:        "{\"order\": \"123456\", \"sum\": 256}",
				storageUser: models.User{
					Balance: 500,
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "must return 400 status",
			args: args{
				contentType: "text/plain",
				body:        "{\"order\": \"123456\", \"sum\": 256}",
				storageUser: models.User{
					Balance: 200,
				},
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "must return 402 status",
			args: args{
				contentType: "application/json",
				body:        "{\"order\": \"123456\", \"sum\": 256}",
				storageUser: models.User{
					Balance: 200,
				},
				storageSaveWithdrawalErr: storage.ErrNotEnoughFunds,
			},
			wantStatusCode: http.StatusPaymentRequired,
		},
		{
			name: "must return 422 status",
			args: args{
				contentType: "application/json",
				body:        "{\"order\": \"123456\", \"sum\": 256}",
				storageUser: models.User{
					Balance: 400,
				},
				storageSaveWithdrawalErr: storage.ErrNotFound,
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "must return 500 status (save error)",
			args: args{
				contentType: "application/json",
				body:        "{\"order\": \"123456\", \"sum\": 256}",
				storageUser: models.User{
					Balance: 400,
				},
				storageSaveWithdrawalErr: errors.New("storage error"),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.New("dev")
			cfg, err := config.New()
			assert.Empty(t, err)

			token, err := auth.GenerateToken("login", cfg.Secret, cfg.TokenExpires)
			assert.Empty(t, err)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockStorage(ctrl)

			m.EXPECT().SaveWithdrawal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.args.storageSaveWithdrawalErr).AnyTimes()

			r := router.New(log, m, cfg.Secret, cfg.TokenExpires)
			srv := httptest.NewServer(r)
			defer srv.Close()

			client := resty.New()

			resp, err := client.R().
				SetHeader("Content-Type", tt.args.contentType).
				SetHeader("Authorization", token).
				SetBody(tt.args.body).
				Post(fmt.Sprintf("%s/%s", srv.URL, "api/user/balance/withdraw"))

			assert.Empty(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
		})
	}
}
