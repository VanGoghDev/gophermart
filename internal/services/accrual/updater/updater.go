package updater

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

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
}

func New(log *slog.Logger, storage storage, accrlHost string) *Updater {
	clnt := client.New(http.Client{}, accrlHost)

	return &Updater{
		log,
		storage,
		clnt,
	}
}

func (u *Updater) Update(
	ctx context.Context,
	ordersCh <-chan models.Order,
	wg *sync.WaitGroup,
) (err error) {
	wg.Add(1)
	defer wg.Done()

	for order := range ordersCh {
		// запросим статус обработки у стороннего сервиса
		accrl, err := u.client.GetAccrual(ctx, order.Number)
		if err != nil {
			u.log.ErrorContext(ctx, "cannot access external accrual service", sl.Err(err))
			continue
		}

		// обновим в бд
		err = u.s.UpdateStatusAndBalance(ctx, accrl)
		if err != nil {
			u.log.ErrorContext(ctx, "failed to update status and balance", sl.Err(err))
			continue
		}
	}

	return nil
}
