package getwithdrawals_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		storageErr error
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
	}{
		{
			name: "must return 200 status",
			args: args{
				storageErr: nil,
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "must return 204 status",
			args: args{
				storageErr: storage.ErrNotFound,
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name: "must return 500 status",
			args: args{
				storageErr: errors.New("storage error"),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.New("dev")
			cfg, err := config.New()
			assert.Empty(t, err)

			token, err := auth.GenerateToken("test", cfg.Secret, cfg.TokenExpires)
			assert.Empty(t, err)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockStorage(ctrl)

			sWithdrawals := make([]models.Withdrawal, 1)
			sWithdrawals[0] = models.Withdrawal{
				OrderNumber: "12345",
				Sum:         500,
				ProcessedAt: time.Now(),
			}
			m.EXPECT().GetWithdrawals(gomock.Any(), gomock.Any()).
				Return(sWithdrawals, tt.args.storageErr).AnyTimes()

			r := router.New(log, m, cfg.Secret, cfg.TokenExpires)
			srv := httptest.NewServer(r)
			defer srv.Close()

			client := resty.New()

			resp, err := client.R().
				SetHeader("Authorization", token).
				Get(fmt.Sprintf("%s/%s", srv.URL, "api/user/withdrawals"))

			assert.Empty(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			if tt.wantStatusCode == http.StatusOK {
				assert.NotEmpty(t, resp.Body())
			}
		})
	}
}
