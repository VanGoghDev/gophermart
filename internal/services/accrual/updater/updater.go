package updater

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/client"
)

type storage interface {
	UpdateStatusAndBalance(ctx context.Context, order models.Accrual) error
}

// Updater запрашивает у стороннего сервиса статус обработки заказа
// и сохраняет обновленный статус в бд, вместе с начисленными бонусами.
type Updater struct {
	log *slog.Logger
	s   storage

	client *client.Client

	retryTimeout time.Duration
}

func New(log *slog.Logger, storage storage, accrlHost string, retryTimeout time.Duration) *Updater {
	clnt := client.New(http.Client{}, accrlHost)

	return &Updater{
		log,
		storage,
		clnt,
		retryTimeout,
	}
}

func (u *Updater) Update(
	ctx context.Context,
	ordersCh <-chan models.Order,
	waitCh chan bool,
	wg *sync.WaitGroup,
) (err error) {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case <-waitCh:
			time.Sleep(u.retryTimeout)
		case order := <-ordersCh:
			// запросим статус обработки у стороннего сервиса

			accrl, err := u.client.GetAccrual(ctx, order.Number)
			if err != nil {
				if errors.Is(err, client.ErrToManyRequests) {
					waitCh <- true
				}
				u.log.ErrorContext(ctx, "cannot access external accrual service", sl.Err(err))
			}

			// обновим в бд
			err = u.s.UpdateStatusAndBalance(ctx, accrl)
			if err != nil {
				u.log.ErrorContext(ctx, "failed to update status and balance", sl.Err(err))
			}
		}
	}
}
