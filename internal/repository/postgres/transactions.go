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

func (t *PostgresTransaction) BeginTx(ctx context.Context) (repository.Transaction, error) {
	return t, nil // Nested transactions not supported in this simple implementation
}

func (t *PostgresTransaction) Commit() error {
	return t.tx.Commit(context.Background())
}

func (t *PostgresTransaction) Rollback() error {
	return t.tx.Rollback(context.Background())
}

func (t *PostgresTransaction) Close() error {
	return nil // Transaction handles its own lifecycle
}

// Implement all repository methods for transaction...
// Each method should use t.tx instead of r.db

func (t *PostgresTransaction) CreateUser(ctx context.Context, user *models.User) error {
	// Same implementation as PostgresRepository but using t.tx
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

// CreateOrder implements repository.Transaction.
func (t *PostgresTransaction) CreateOrder(ctx context.Context, order *models.Order) error {
	panic("unimplemented")
}

// CreateWithdrawal implements repository.Transaction.
func (t *PostgresTransaction) CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error {
	panic("unimplemented")
}

// GetBalance implements repository.Transaction.
func (t *PostgresTransaction) GetBalance(ctx context.Context, userID string) (*models.Balance, error) {
	panic("unimplemented")
}

// GetOrderByNumber implements repository.Transaction.
func (t *PostgresTransaction) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	panic("unimplemented")
}

// GetOrdersForProcessing implements repository.Transaction.
func (t *PostgresTransaction) GetOrdersForProcessing(ctx context.Context, limit int) ([]*models.Order, error) {
	panic("unimplemented")
}

// GetUserByID implements repository.Transaction.
func (t *PostgresTransaction) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	panic("unimplemented")
}

// GetUserByLogin implements repository.Transaction.
func (t *PostgresTransaction) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	panic("unimplemented")
}

// GetUserOrders implements repository.Transaction.
func (t *PostgresTransaction) GetUserOrders(ctx context.Context, userID string) ([]*models.Order, error) {
	panic("unimplemented")
}

// GetWithdrawals implements repository.Transaction.
func (t *PostgresTransaction) GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
	panic("unimplemented")
}

// UpdateBalance implements repository.Transaction.
func (t *PostgresTransaction) UpdateBalance(ctx context.Context, userID string, current float64, withdrawn float64) error {
	panic("unimplemented")
}

// UpdateOrder implements repository.Transaction.
func (t *PostgresTransaction) UpdateOrder(ctx context.Context, order *models.Order) error {
	panic("unimplemented")
}
