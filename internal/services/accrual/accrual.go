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

	workersCount int32
}

func New(log *slog.Logger, oPool *orderspool.OrdersPool, updtr *updater.Updater, wrkrsCount int32) *AccrualFetcher {
	return &AccrualFetcher{
		log:          log,
		ordrPool:     oPool,
		updater:      updtr,
		workersCount: wrkrsCount,
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
	// Здесь образуется очередь из заказов, которые нужно обновить
	ordersCh := make(chan models.Order, a.workersCount)

	// канал, в который приходит сигнал о том, что нужно подождать с новыми запросами, если пришел ответ со статусом 429
	waitCh := make(chan bool, 1)

	g.Go(func() error {
		err := a.ordrPool.GetOrders(ctx, ordersCh, wg)
		if err != nil {
			return fmt.Errorf("failed to get orders: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		err := a.updater.Update(ctx, ordersCh, waitCh, wg)
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("group finished with error: %w", err)
	}

	return nil
}
