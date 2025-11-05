package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/repository"
)

func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query, user.Login, user.Password).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return repository.ErrUserExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
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

func (r *PostgresRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, login).Scan(
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
