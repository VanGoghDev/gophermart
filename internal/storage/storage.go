package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"log/slog"
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
	ErrNotFound       = errors.New("records not found")
	ErrAlreadyExists  = errors.New("record alreay exists")
	ErrNotEnoughFunds = errors.New("not enough funds")
	ErrConflict       = errors.New("conflict")
	ErrGoodConflict   = errors.New("positive conflict")
)

type Storage struct {
	log *slog.Logger
	db  *pgxpool.Pool
}

func New(ctx context.Context, log *slog.Logger, storagePath string) (*Storage, error) {
	const op = "storage.New"

	pool, err := pgxpool.New(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{
		log: log,
		db:  pool,
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

func (s *Storage) GetUser(ctx context.Context, login string) (user models.User, err error) {
	const op = "storage.GetUser"
	row := s.db.QueryRow(ctx, "SELECT login, pass_hash, balance FROM users WHERE login = $1", login)
	err = row.Scan(&user.Login, &user.PassHash, &user.Balance)
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

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				s.log.ErrorContext(ctx, "%s: %w", op, err)
			}
		}
	}()

	var ordrNum, usrLogin string
	err = tx.QueryRow(ctx, "SELECT number, user_login FROM orders WHERE number = $1", number).Scan(&ordrNum, &usrLogin)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	if ordrNum != "" {
		if userLogin == usrLogin {
			return ErrGoodConflict
		}
		return ErrConflict
	}

	_, err = tx.Prepare(ctx, "saveOrder", "INSERT INTO orders(number, user_login, status) VALUES($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, "saveOrder", number, userLogin, status)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetBalance(ctx context.Context, userLogin string) (balance models.Balance, err error) {
	const op = "storage.GetBalance"

	row := s.db.QueryRow(ctx, "SELECT u.balance, total FROM users u "+
		"INNER JOIN (SELECT user_login, SUM(withdrawal_sum) AS total FROM withdrawals GROUP BY user_login) "+
		"W On W.user_login = u.login WHERE u.login = $1;", userLogin)
	err = row.Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Balance{}, ErrNotFound
		}
		return models.Balance{}, fmt.Errorf("%s: %w", op, err)
	}

	return balance, nil
}

func (s *Storage) GetWithdrawals(ctx context.Context, userLogin string) (withdrawals []models.Withdrawal, err error) {
	const op = "storage.GetWithdrawals"
	withdrawals = make([]models.Withdrawal, 0)

	rows, err := s.db.Query(
		ctx,
		"SELECT order_id, withdrawal_sum, processed_at FROM withdrawals WHERE user_login = $1 ORDER by processed_at",
		userLogin,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()
	for rows.Next() {
		var w = models.Withdrawal{}
		err = rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		w.ProcessedAtFormated = w.ProcessedAt.Format(time.RFC3339)
		withdrawals = append(withdrawals, w)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return withdrawals, nil
}

func (s *Storage) SaveWithdrawal(ctx context.Context, userLogin string, orderNum string, sum int64) (err error) {
	const op = "storage.SaveWithdrawal"
	log.Print(op)
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				s.log.ErrorContext(ctx, "%s: %w", op, err)
			}
		}
	}()

	var ordrNum string
	err = tx.QueryRow(ctx, "SELECT number FROM orders WHERE number = $1", orderNum).
		Scan(&ordrNum)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	var balance int64
	err = tx.QueryRow(ctx, "SELECT balance FROM users WHERE login=$1 FOR UPDATE", userLogin).Scan(&balance)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if balance < sum {
		return ErrNotEnoughFunds
	}

	_, err = tx.Prepare(ctx, "insrtWithdrawals", "INSERT INTO withdrawals(user_login, order_id, withdrawal_sum)"+
		"VALUES($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Prepare(ctx, "updBalance", "UPDATE users SET balance = balance - $1 WHERE login = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, "insrtWithdrawals", userLogin, orderNum, sum)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, "updBalance", sum, userLogin)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetOrdersByStatus(
	ctx context.Context,
	statuses ...models.OrderStatus,
) (orders []models.Order, err error) {
	const op = "storage.GetOrdersByStatus"
	orders = make([]models.Order, 0)

	rows, err := s.db.Query(ctx, "SELECT number FROM orders WHERE status = ANY($1)", pq.Array(statuses))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
	}
	defer rows.Close()
	for rows.Next() {
		var order = models.Order{}
		err = rows.Scan(&order.Number)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		orders = append(orders, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return orders, nil
}

func (s *Storage) UpdateStatusAndBalance(ctx context.Context, accrual models.Accrual) error {
	const op = "storage.UpdateStatusAndBalance"

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				s.log.ErrorContext(ctx, "%s: %w", op, err)
			}
		}
	}()
	var userLogin string
	err = tx.QueryRow(ctx, "SELECT user_login FROM orders WHERE number = $1", accrual.OrderNum).Scan(&userLogin)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Prepare(ctx, "updtBalance", "UPDATE users SET balance = balance + $1 WHERE login = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Prepare(ctx, "updStatus", "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, "updtBalance", accrual.Accrual, userLogin)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, "updStatus", accrual.Status, accrual.Accrual, accrual.OrderNum)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
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
