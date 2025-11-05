package service

import (
	"context"
	"errors"

	"github.com/rompil2/gomegamarket/internal/models"
)

var (
	ErrUserExists                = errors.New("user already exists")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrOrderExists               = errors.New("order already exists")
	ErrOrderExistsForUser        = errors.New("order already exists for this user")
	ErrOrderExistsForOtherUser   = errors.New("order exists for other user")
	ErrInvalidOrderNumber        = errors.New("invalid order number")
	ErrInsufficientBalance       = errors.New("insufficient balance")
	ErrOrderNotFound             = errors.New("order not found")
	ErrRateLimitExceeded         = errors.New("accrual service rate limit exceeded")
	ErrAccrualServiceInternal    = errors.New("accrual service internal error")
	ErrAccrualServiceUnavailable = errors.New("accrual service unavailable")
)

type AuthService interface {
	Register(ctx context.Context, login, password string) (*models.User, string, error)
	Login(ctx context.Context, login, password string) (*models.User, string, error)
	ValidateToken(token string) (string, error)
}

type OrderService interface {
	UploadOrder(ctx context.Context, userID, orderNumber string) (int, error)
	GetUserOrders(ctx context.Context, userID string) ([]*models.Order, error)
	ProcessOrders(ctx context.Context) error
}

type BalanceService interface {
	GetBalance(ctx context.Context, userID string) (*models.Balance, error)
	Withdraw(ctx context.Context, userID string, req *models.WithdrawRequest) error
	GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error)
}

type Service interface {
	AuthService
	OrderService
	BalanceService
}

// AccrualService defines the contract for interacting with the external accrual system.
type AccrualService interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (*models.OrderAccrual, error)
}
