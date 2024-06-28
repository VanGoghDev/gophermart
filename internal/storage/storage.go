package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, storagePath string) (*Storage, error) {
	const op = "storage.New"

	pool, err := pgxpool.New(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{
		db: pool,
	}, nil
}
