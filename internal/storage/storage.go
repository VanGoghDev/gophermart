package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound      = errors.New("records not found")
	ErrAlreadyExists = errors.New("record alreay exists")
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

func (s *Storage) RegisterUser(ctx context.Context, login string, password string) (lgn string, err error) {
	const op = "storage.RegisterNewUser"

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.Exec(ctx, "INSERT INTO users(login, pass_hash) VALUES($1, $2)",
		login, passHash)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return "", ErrAlreadyExists
			}
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return login, nil
}

func (s *Storage) GetUser(ctx context.Context, login string, password string) (user models.User, err error) {
	const op = "storage.GetUser"
	row := s.db.QueryRow(ctx, "SELECT login, pass_hash FROM users")
	err = row.Scan(&user.Login, &user.PassHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (s *Storage) GetOrder(ctx context.Context, number string) (order models.Order, err error) {
	const op = "storage.GetOrder"
	row := s.db.QueryRow(ctx, "SELECT number, status, accrual, uploaded_at FROM orders WHERE number = $1", number)
	err = row.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, ErrNotFound
		}
		return models.Order{}, fmt.Errorf("%s: %w", op, err)
	}

	return order, nil
}

func (s *Storage) GetOrders(ctx context.Context, userLogin string) (orders []models.Order, err error) {
	const op = "storage.GetOrders"
	orders = make([]models.Order, 0)

	rows, err := s.db.Query(
		ctx,
		"SELECT number, status, accrual, uploaded_at FROM orders WHERE user_login = $1 ORDER by uploaded_at",
		userLogin,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
	}

	defer rows.Close()
	for rows.Next() {
		var order = models.Order{}
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		order.UploadedAtFormated = order.UploadedAt.Format(time.RFC3339)
		orders = append(orders, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return orders, nil
}

func (s *Storage) SaveOrder(
	ctx context.Context,
	number string,
	userLogin string,
	status models.OrderStatus,
) (err error) {
	const op = "storage.SaveOrder"
	_, err = s.db.Exec(ctx, "INSERT INTO orders(number, user_login, status) VALUES($1, $2, $3)",
		number, userLogin, status)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrAlreadyExists
			}
			return fmt.Errorf("%s: %w", op, err)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func (s *Storage) RunMigrations(dsn string) error {
	const op = "storage.runMigrations"

	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		d,
		dsn,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}
