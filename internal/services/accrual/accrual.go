package accrual

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/dispatcher"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/orderspool"
	"github.com/VanGoghDev/gophermart/internal/storage"
	"golang.org/x/sync/errgroup"
)

type AccrualFetcher struct {
	log *slog.Logger

	ordrPool   *orderspool.OrdersPool
	dispatcher *dispatcher.Dispatcher

	workersCount int32
}

func New(log *slog.Logger, oPool *orderspool.OrdersPool, s *storage.Storage, a string, wrkrsCount int32) *AccrualFetcher {
	d := dispatcher.New(log, s, a, wrkrsCount)
	return &AccrualFetcher{
		log:          log,
		ordrPool:     oPool,
		dispatcher:   d,
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

	g.Go(func() error {
		err := a.ordrPool.GetOrders(ctx, ordersCh, wg)
		if err != nil {
			return fmt.Errorf("failed to get orders: %w", err)
		}
		return nil
	})

	wg.Add(1)
	g.Go(func() error {
		err := a.dispatcher.Run(ctx, wg, g, ordersCh)
		if err != nil {
			return fmt.Errorf("%w: failed to run dispatcher", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("group finished with error: %w", err)
	}

	return nil
}
