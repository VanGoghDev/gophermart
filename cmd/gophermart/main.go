package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/VanGoghDev/gophermart/internal/config"
	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
	"github.com/VanGoghDev/gophermart/internal/logger"
	"github.com/VanGoghDev/gophermart/internal/router"
	"github.com/VanGoghDev/gophermart/internal/services/accrual"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/orderspool"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/updater"
	"github.com/VanGoghDev/gophermart/internal/storage"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}

const (
	timeoutShutdown       = time.Second * 10
	timeoutServerShutdown = time.Second * 5
)

func run() error {
	var wg sync.WaitGroup

	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	g, ctx := errgroup.WithContext(rootCtx)

	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to init config: %w", err)
	}

	slog := logger.New(cfg.Env)
	slog.DebugContext(ctx, "server started", "address", cfg.Address)

	s, err := storage.New(ctx, slog, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to init storage: %w", err)
	}

	g.Go(func() error {
		wg.Add(1)
		defer wg.Done()

		defer slog.DebugContext(ctx, "closed DB")

		<-ctx.Done()

		s.Close()
		return nil
	})

	rtr := router.New(slog, s, cfg.Secret, cfg.TokenExpires)

	oPool := orderspool.New(slog, s, cfg.AccrualTimeout)
	updtr := updater.New(slog, s, cfg.AccrualAddress)
	accrl := accrual.New(slog, oPool, updtr)

	g.Go(func() error {
		err := accrl.RunService(ctx, g, &wg)
		if err != nil {
			return fmt.Errorf("failed to run accrual service: %w", err)
		}
		return nil
	})

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: rtr,
	}
	g.Go(func() error {
		wg.Add(1)
		err = srv.ListenAndServe()
		if err != nil {
			return fmt.Errorf("failed to run http server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		wg.Add(1)
		defer wg.Done()

		<-ctx.Done()
		slog.InfoContext(ctx, "server has been shutdown")

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			slog.ErrorContext(ctx, "failed to shutdown server: %w", sl.Err(err))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to wait group: %w", err)
	}

	return nil
}
