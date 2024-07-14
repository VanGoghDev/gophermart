package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/VanGoghDev/gophermart/internal/config"
	"github.com/VanGoghDev/gophermart/internal/logger"
	"github.com/VanGoghDev/gophermart/internal/router"
	"github.com/VanGoghDev/gophermart/internal/services/accrual"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/orderspool"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/updater"
	"github.com/VanGoghDev/gophermart/internal/storage"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(context.Background()); err != nil {
		const op = "main"
		log.Fatalf("%s: %v", op, err)
	}
}

func run(ctx context.Context) error {
	const op = "main.run"

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	slog := logger.New(cfg.Env)
	slog.DebugContext(ctx, "server started", "address", cfg.Address)

	s, err := storage.New(ctx, slog, cfg.DSN)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.RunMigrations(cfg.DSN)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rtr := router.New(slog, s, cfg.Secret, cfg.TokenExpires)

	oPool := orderspool.New(slog, s)
	updtr := updater.New(slog, s, cfg.AccrualAddress)
	accrl := accrual.New(slog, oPool, updtr)

	// waitgroup?
	var wg sync.WaitGroup

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		err := accrl.RunService(ctx, g, &wg)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})

	err = http.ListenAndServe(cfg.Address, rtr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	wg.Wait()
	return nil
}
