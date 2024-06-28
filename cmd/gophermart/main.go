package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/VanGoghDev/gophermart/internal/config"
	"github.com/VanGoghDev/gophermart/internal/logger"
	"github.com/VanGoghDev/gophermart/internal/router"
	"github.com/VanGoghDev/gophermart/internal/services/accrual"
	"github.com/VanGoghDev/gophermart/internal/storage"
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

	s, err := storage.New(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rtr := router.New(s, slog)

	accrl := accrual.New()

	// waitgroup?
	go accrl.Serve()

	err = http.ListenAndServe(cfg.Address, rtr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
