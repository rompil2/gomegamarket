package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/repository"
)

type PostgresTransaction struct {
	tx pgx.Tx
}

func (t *PostgresTransaction) Commit() error {
	return t.tx.Commit(context.Background())
}

func (t *PostgresTransaction) Rollback() error {
	return t.tx.Rollback(context.Background())
}

func (t *PostgresTransaction) BeginTx(ctx context.Context) (repository.Transaction, error) {
	return t, nil
}

func (t *PostgresTransaction) Close() error {
	return nil
}

// UserRepository methods
func (t *PostgresTransaction) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := t.tx.QueryRow(ctx, query, user.Login, user.Password).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return repository.ErrUserExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (t *PostgresTransaction) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := t.tx.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (t *PostgresTransaction) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`

	var user models.User
	err := t.tx.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by login: %w", err)
	}

	return &user, nil
}

// OrderRepository methods
func (t *PostgresTransaction) CreateOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (number, user_id, status)
		VALUES ($1, $2, $3)
		RETURNING uploaded_at
	`

	err := t.tx.QueryRow(ctx, query, order.Number, order.UserID, order.Status).Scan(&order.UploadedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return repository.ErrOrderExists
		}
		return fmt.Errorf("create order: %w", err)
	}

	return nil
}

func (t *PostgresTransaction) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	query := `
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE number = $1
	`

	var order models.Order
	var accrual *float64

	err := t.tx.QueryRow(ctx, query, number).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrOrderNotFound
		}
		return nil, fmt.Errorf("get order by number: %w", err)
	}

	order.Accrual = accrual
	return &order, nil
}

func (t *PostgresTransaction) GetUserOrders(ctx context.Context, userID string) ([]*models.Order, error) {
	query := `
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := t.tx.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var accrual *float64

		err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}

		order.Accrual = accrual
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

func (t *PostgresTransaction) GetOrdersForProcessing(ctx context.Context, limit int) ([]*models.Order, error) {
	query := `
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
		ORDER BY uploaded_at ASC
		LIMIT $1
	`

	rows, err := t.tx.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get orders for processing: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		var accrual *float64

		err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}

		order.Accrual = accrual
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

func (t *PostgresTransaction) UpdateOrder(ctx context.Context, order *models.Order) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE number = $3
	`

	_, err := t.tx.Exec(ctx, query, order.Status, order.Accrual, order.Number)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}

	return nil
}

// BalanceRepository methods
func (t *PostgresTransaction) GetBalance(ctx context.Context, userID string) (*models.Balance, error) {
	query := `
		SELECT current_balance, withdrawn_balance
		FROM balances
		WHERE user_id = $1
	`

	var balance models.Balance
	err := t.tx.QueryRow(ctx, query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &models.Balance{Current: 0, Withdrawn: 0}, nil
		}
		return nil, fmt.Errorf("get balance: %w", err)
	}

	return &balance, nil
}

func (t *PostgresTransaction) CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
		RETURNING processed_at
	`

	err := t.tx.QueryRow(ctx, query, withdrawal.UserID, withdrawal.Order, withdrawal.Sum).Scan(
		&withdrawal.ProcessedAt,
	)
	if err != nil {
		return fmt.Errorf("create withdrawal: %w", err)
	}

	return nil
}

func (t *PostgresTransaction) GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
	query := `
		SELECT order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := t.tx.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []*models.Withdrawal
	for rows.Next() {
		var withdrawal models.Withdrawal

		err := rows.Scan(
			&withdrawal.Order,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}

		withdrawals = append(withdrawals, &withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return withdrawals, nil
}

func (t *PostgresTransaction) UpdateBalance(ctx context.Context, userID string, current, withdrawn float64) error {
	query := `
		INSERT INTO balances (user_id, current_balance, withdrawn_balance)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			current_balance = EXCLUDED.current_balance,
			withdrawn_balance = EXCLUDED.withdrawn_balance,
			updated_at = NOW()
	`

	_, err := t.tx.Exec(ctx, query, userID, current, withdrawn)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	return nil
}
