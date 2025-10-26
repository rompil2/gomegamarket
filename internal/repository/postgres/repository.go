package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rompil2/gomegamarket/internal/repository"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func New(databaseURI string) (*PostgresRepository, error) {
	config, err := pgxpool.ParseConfig(databaseURI)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() error {
	if r.db != nil {
		r.db.Close()
	}
	return nil
}

func (r *PostgresRepository) BeginTx(ctx context.Context) (repository.Transaction, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	return &PostgresTransaction{tx: tx}, nil
}

// Helper function to check if error is a unique violation
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
