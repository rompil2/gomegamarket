package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/repository"
)

func (r *PostgresRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (number, user_id, status)
		VALUES ($1, $2, $3)
		RETURNING uploaded_at
	`

	err := r.db.QueryRow(ctx, query, order.Number, order.UserID, order.Status).Scan(&order.UploadedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return repository.ErrOrderExists
		}
		return fmt.Errorf("create order: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	query := `
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE number = $1
	`

	var order models.Order
	var accrual *float64

	err := r.db.QueryRow(ctx, query, number).Scan(
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

func (r *PostgresRepository) GetUserOrders(ctx context.Context, userID string) ([]*models.Order, error) {
	query := `
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
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

func (r *PostgresRepository) GetOrdersForProcessing(ctx context.Context, limit int) ([]*models.Order, error) {
	query := `
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
		ORDER BY uploaded_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
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

func (r *PostgresRepository) UpdateOrder(ctx context.Context, order *models.Order) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE number = $3
	`

	_, err := r.db.Exec(ctx, query, order.Status, order.Accrual, order.Number)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}

	return nil
}
