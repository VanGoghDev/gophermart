package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/lib/logger/sl"
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

var (
	failedToRollbackLogMsg = "failed to rollback %w"
)

type Storage struct {
	log *slog.Logger
	db  *pgxpool.Pool
}

func New(ctx context.Context, slg *slog.Logger, storagePath string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to init pool connection: %w", err)
	}
	return &Storage{
		log: slg,
		db:  pool,
	}, nil
}

func (s *Storage) RegisterUser(ctx context.Context, login string, password string) (lgn string, err error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
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
		return "", fmt.Errorf("failed to insert users: %w", err)
	}

	return login, nil
}

func (s *Storage) GetUser(ctx context.Context, login string) (user models.User, err error) {
	row := s.db.QueryRow(ctx, "SELECT login, pass_hash, balance FROM users WHERE login = $1", login)
	err = row.Scan(&user.Login, &user.PassHash, &user.Balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("failed to select users: %w", err)
	}
	return user, nil
}

func (s *Storage) GetOrder(ctx context.Context, number string) (order models.Order, err error) {
	row := s.db.QueryRow(ctx, "SELECT number, status, accrual, uploaded_at FROM orders WHERE number = $1", number)
	err = row.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, ErrNotFound
		}
		return models.Order{}, fmt.Errorf("failed to select order: %w", err)
	}

	return order, nil
}

func (s *Storage) GetOrders(ctx context.Context, userLogin string) (orders []models.Order, err error) {
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
			return nil, fmt.Errorf("failed to select orders: %w", err)
		}
		order.UploadedAtFormated = order.UploadedAt.Format(time.RFC3339)
		orders = append(orders, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate through rows: %w", err)
	}

	return orders, nil
}

func (s *Storage) SaveOrder(
	ctx context.Context,
	number string,
	userLogin string,
	status models.OrderStatus,
) (err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				s.log.ErrorContext(ctx, failedToRollbackLogMsg, sl.Err(err))
			}
		}
	}()

	var ordrNum, usrLogin string
	err = tx.QueryRow(ctx, "SELECT number, user_login FROM orders WHERE number = $1", number).Scan(&ordrNum, &usrLogin)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to select orders: %w", err)
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
		return fmt.Errorf("failed to prepare statement saveOrder: %w", err)
	}

	_, err = tx.Exec(ctx, "saveOrder", number, userLogin, status)
	if err != nil {
		return fmt.Errorf("failed to execute saveOrder: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) GetBalance(ctx context.Context, userLogin string) (balance models.Balance, err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Balance{}, fmt.Errorf("failed to init transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				s.log.ErrorContext(ctx, failedToRollbackLogMsg, sl.Err(err))
			}
		}
	}()

	var blnc float64
	err = tx.QueryRow(ctx, "SELECT balance FROM users WHERE login = $1", userLogin).
		Scan(&blnc)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Balance{}, ErrNotFound
		}
		return models.Balance{}, fmt.Errorf("failed to select balance: %w", err)
	}

	var withdraw float64
	err = tx.QueryRow(ctx,
		"SELECT COALESCE(SUM(withdrawal_sum), 0)  AS total FROM withdrawals where user_login = $1",
		userLogin,
	).
		Scan(&withdraw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Balance{}, ErrNotFound
		}
		return models.Balance{}, fmt.Errorf("failed to select withdrawals: %w", err)
	}

	return models.Balance{
		Current:   blnc,
		Withdrawn: withdraw,
	}, nil
}

func (s *Storage) GetWithdrawals(ctx context.Context, userLogin string) (withdrawals []models.Withdrawal, err error) {
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
		return nil, fmt.Errorf("failed to select order: %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		var w = models.Withdrawal{}
		err = rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}
		w.ProcessedAtFormated = w.ProcessedAt.Format(time.RFC3339)
		withdrawals = append(withdrawals, w)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate through rows: %w", err)
	}

	return withdrawals, nil
}

func (s *Storage) SaveWithdrawal(ctx context.Context, userLogin string, orderNum string, sum float64) (err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to init transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				s.log.ErrorContext(ctx, failedToRollbackLogMsg, sl.Err(err))
			}
		}
	}()

	var balance float64
	err = tx.QueryRow(ctx, "SELECT balance FROM users WHERE login=$1 FOR UPDATE", userLogin).Scan(&balance)
	if err != nil {
		return fmt.Errorf("failed to select balance: %w", err)
	}
	if balance < sum {
		return ErrNotEnoughFunds
	}

	_, err = tx.Prepare(ctx, "insrtWithdrawals", "INSERT INTO withdrawals(user_login, order_id, withdrawal_sum)"+
		"VALUES($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("failed to prepare insrtWithdrawals: %w", err)
	}

	_, err = tx.Prepare(ctx, "updBalance", "UPDATE users SET balance = balance - $1 WHERE login = $2")
	if err != nil {
		return fmt.Errorf("failed to prepare updBalance: %w", err)
	}

	_, err = tx.Exec(ctx, "insrtWithdrawals", userLogin, orderNum, sum)
	if err != nil {
		return fmt.Errorf("failed to execute insrtWithdrawals: %w", err)
	}

	_, err = tx.Exec(ctx, "updBalance", sum, userLogin)
	if err != nil {
		return fmt.Errorf("failed to execute updBalance: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) GetOrdersByStatus(
	ctx context.Context,
	statuses ...models.OrderStatus,
) (orders []models.Order, err error) {
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
			return nil, fmt.Errorf("failed to select order by status: %w", err)
		}
		orders = append(orders, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate throug rows: %w", err)
	}
	return orders, nil
}

func (s *Storage) UpdateStatusAndBalance(ctx context.Context, accrual models.Accrual) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to init transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				s.log.ErrorContext(ctx, failedToRollbackLogMsg, sl.Err(err))
			}
		}
	}()
	var userLogin string
	err = tx.QueryRow(ctx, "SELECT user_login FROM orders WHERE number = $1", accrual.OrderNum).Scan(&userLogin)
	if err != nil {
		return fmt.Errorf("failed to select user_login: %w", err)
	}

	_, err = tx.Prepare(ctx, "updtBalance", "UPDATE users SET balance = balance + $1 WHERE login = $2")
	if err != nil {
		return fmt.Errorf("failed to prepare updtBalance: %w", err)
	}

	_, err = tx.Prepare(ctx, "updStatus", "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3")
	if err != nil {
		return fmt.Errorf("failed to prepare updStatus: %w", err)
	}

	_, err = tx.Exec(ctx, "updtBalance", accrual.Accrual, userLogin)
	if err != nil {
		return fmt.Errorf("failed to execute updtBalance: %w", err)
	}

	_, err = tx.Exec(ctx, "updStatus", accrual.Status, accrual.Accrual, accrual.OrderNum)
	if err != nil {
		return fmt.Errorf("failed to execute updStatus: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
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
