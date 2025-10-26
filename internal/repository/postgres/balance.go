package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/rompil2/gomegamarket/internal/models"

	"github.com/rompil2/gomegamarket/internal/repository"
)

func (r *PostgresRepository) GetBalance(ctx context.Context, userID string) (*models.Balance, error) {
	query := `
		SELECT current_balance, withdrawn_balance
		FROM balances
		WHERE user_id = $1
	`

	var balance models.Balance
	err := r.db.QueryRow(ctx, query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return zero balance if user doesn't have a balance record
			return &models.Balance{Current: 0, Withdrawn: 0}, nil
		}
		return nil, fmt.Errorf("get balance: %w", err)
	}

	return &balance, nil
}

func (r *PostgresRepository) CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check current balance
	var currentBalance float64
	err = tx.QueryRow(ctx, "SELECT current_balance FROM balances WHERE user_id = $1 FOR UPDATE", withdrawal.UserID).Scan(&currentBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrInsufficientBalance
		}
		return fmt.Errorf("get balance for update: %w", err)
	}

	if currentBalance < withdrawal.Sum {
		return repository.ErrInsufficientBalance
	}

	// Insert withdrawal
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
		RETURNING id, processed_at
	`

	err = tx.QueryRow(ctx, query, withdrawal.UserID, withdrawal.Order, withdrawal.Sum).Scan(&withdrawal.UserID, &withdrawal.ProcessedAt)
	if err != nil {
		return fmt.Errorf("create withdrawal: %w", err)
	}

	// Update balance
	updateQuery := `
		INSERT INTO balances (user_id, current_balance, withdrawn_balance)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			current_balance = balances.current_balance - EXCLUDED.current_balance,
			withdrawn_balance = balances.withdrawn_balance + EXCLUDED.withdrawn_balance,
			updated_at = NOW()
	`

	_, err = tx.Exec(ctx, updateQuery, withdrawal.UserID, withdrawal.Sum, withdrawal.Sum)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
	query := `
		SELECT order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
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

func (r *PostgresRepository) UpdateBalance(ctx context.Context, userID string, current, withdrawn float64) error {
	query := `
		INSERT INTO balances (user_id, current_balance, withdrawn_balance)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			current_balance = EXCLUDED.current_balance,
			withdrawn_balance = EXCLUDED.withdrawn_balance,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, userID, current, withdrawn)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	return nil
}
