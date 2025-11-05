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
	// 1. Валидация номера заказа
	if !utils.ValidLuhn(req.Order) {
		return ErrInvalidOrderNumber
	}

	// 2. Проверка существования заказа (для списания)
	existingOrder, err := s.repo.GetOrderByNumber(ctx, req.Order)
	if err != nil && !errors.Is(err, repository.ErrOrderNotFound) {
		return fmt.Errorf("check order: %w", err)
	}
	if existingOrder != nil {
		return fmt.Errorf("order %s already exists", req.Order)
	}

	// 3. Начать транзакцию
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() // Всегда откатываем при ошибке

	// 4. Получить баланс ВНУТРИ транзакции (с блокировкой для обновления)
	balance, err := tx.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("get balance: %w", err)
	}

	// 5. Проверить достаточность средств
	if balance.Current < req.Sum {
		return ErrInsufficientBalance
	}

	// 6. Создать запись о списании ВНУТРИ транзакции
	withdrawal := &models.Withdrawal{
		UserID: userID,
		Order:  req.Order,
		Sum:    req.Sum,
	}

	err = tx.CreateWithdrawal(ctx, withdrawal)
	if err != nil {
		return fmt.Errorf("create withdrawal: %w", err)
	}

	// 7. Обновить баланс ВНУТРИ транзакции
	newCurrent := balance.Current - req.Sum
	newWithdrawn := balance.Withdrawn + req.Sum

	err = tx.UpdateBalance(ctx, userID, newCurrent, newWithdrawn)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	// 8. Зафиксировать транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
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
