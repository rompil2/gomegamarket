package repository

import (
	"context"
	"errors"

	"github.com/rompil2/gomegamarket/internal/models"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserExists          = errors.New("user already exists")
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderExists         = errors.New("order already exists")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidOrderNumber  = errors.New("invalid order number")
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID string) ([]*models.Order, error)
	GetOrdersForProcessing(ctx context.Context, limit int) ([]*models.Order, error)
	UpdateOrder(ctx context.Context, order *models.Order) error
}

type BalanceRepository interface {
	GetBalance(ctx context.Context, userID string) (*models.Balance, error)
	CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error)
	UpdateBalance(ctx context.Context, userID string, current, withdrawn float64) error
}

type Repository interface {
	UserRepository
	OrderRepository
	BalanceRepository
	BeginTx(ctx context.Context) (Transaction, error)
	Close() error
}

type Transaction interface {
	Repository
	Commit() error
	Rollback() error
}
