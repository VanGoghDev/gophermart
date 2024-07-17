package orderspool

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
)

type storage interface {
	GetOrdersByStatus(ctx context.Context, statuses ...models.OrderStatus) ([]models.Order, error)
}

type OrdersPool struct {
	log     *slog.Logger
	s       storage
	timeOut time.Duration
}

func New(log *slog.Logger, s storage, t time.Duration) *OrdersPool {
	return &OrdersPool{
		log:     log,
		s:       s,
		timeOut: t,
	}
}

func (o *OrdersPool) GetOrders(
	ctx context.Context,
	ordersCh chan<- models.Order,
	wg *sync.WaitGroup,
) error {
	const op = "services.orderspool.FetchOrders"
	log := o.log.With("op", op)
	wg.Add(1)
	defer wg.Done()

	for {
		orders, err := o.s.GetOrdersByStatus(ctx, models.New, models.Processing, models.Registered)
		if err != nil {
			log.WarnContext(ctx, "", sl.Err(err))
			close(ordersCh)
			return fmt.Errorf("%s: %w", op, err)
		}

		for _, v := range orders {
			time.Sleep(o.timeOut)
			ordersCh <- v
		}
	}
}
