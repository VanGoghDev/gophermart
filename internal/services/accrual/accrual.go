package accrual

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/orderspool"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/updater"
	"golang.org/x/sync/errgroup"
)

type AccrualFetcher struct {
	log *slog.Logger

	ordrPool *orderspool.OrdersPool
	updater  *updater.Updater
}

func New(log *slog.Logger, oPool *orderspool.OrdersPool, updtr *updater.Updater) *AccrualFetcher {
	return &AccrualFetcher{
		log:      log,
		ordrPool: oPool,
		updater:  updtr,
	}
}

func (a *AccrualFetcher) RunService(ctx context.Context, g *errgroup.Group, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	if err := a.Run(ctx, g, wg); err != nil {
		return fmt.Errorf("failed to run app: %w", err)
	}

	return nil
}

func (a *AccrualFetcher) Run(ctx context.Context, g *errgroup.Group, wg *sync.WaitGroup) error {
	const op = "services.accrual.Serve"
	rateLimit := 5

	// Здесь образуется очередь из заказов, которые нужно обновить
	ordersCh := make(chan models.Order, rateLimit)

	g.Go(func() error {
		err := a.ordrPool.GetOrders(ctx, ordersCh, wg)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})

	g.Go(func() error {
		err := a.updater.Update(ctx, ordersCh, wg)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
