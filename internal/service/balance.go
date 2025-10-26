package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/repository"
	"github.com/rompil2/gomegamarket/internal/utils"
)

type BalanceServiceImpl struct {
	repo repository.Repository
}

func NewBalanceService(repo repository.Repository) *BalanceServiceImpl {
	return &BalanceServiceImpl{
		repo: repo,
	}
}

func (s *BalanceServiceImpl) GetBalance(ctx context.Context, userID string) (*models.Balance, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	return balance, nil
}

func (s *BalanceServiceImpl) Withdraw(ctx context.Context, userID string, req *models.WithdrawRequest) error {
	// Validate order number
	if !utils.ValidLuhn(req.Order) {
		return ErrInvalidOrderNumber
	}

	// Check if order already exists for withdrawal
	existingOrder, err := s.repo.GetOrderByNumber(ctx, req.Order)
	if err != nil && !errors.Is(err, repository.ErrOrderNotFound) {
		return fmt.Errorf("check order: %w", err)
	}

	if existingOrder != nil {
		return fmt.Errorf("order %s already exists", req.Order)
	}

	// Check sufficient balance
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	if balance.Current < req.Sum {
		return ErrInsufficientBalance
	}

	// Create withdrawal
	withdrawal := &models.Withdrawal{
		UserID: userID,
		Order:  req.Order,
		Sum:    req.Sum,
	}

	err = s.repo.CreateWithdrawal(ctx, withdrawal)
	if err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return ErrInsufficientBalance
		}
		return fmt.Errorf("create withdrawal: %w", err)
	}

	return nil
}

func (s *BalanceServiceImpl) GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
	withdrawals, err := s.repo.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals: %w", err)
	}

	return withdrawals, nil
}
