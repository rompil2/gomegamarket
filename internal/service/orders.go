package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/repository"
	"github.com/rompil2/gomegamarket/internal/utils"
)

type OrderServiceImpl struct {
	repo           repository.Repository
	accrualService AccrualService
}

func NewOrderService(repo repository.Repository, accrualService AccrualService) *OrderServiceImpl {
	return &OrderServiceImpl{
		repo:           repo,
		accrualService: accrualService,
	}
}

func (s *OrderServiceImpl) UploadOrder(ctx context.Context, userID, orderNumber string) (int, error) {
	// Validate order number using Luhn algorithm
	if !utils.ValidLuhn(orderNumber) {
		return 0, ErrInvalidOrderNumber
	}

	// Check if order already exists
	existingOrder, err := s.repo.GetOrderByNumber(ctx, orderNumber)
	if err != nil && !errors.Is(err, repository.ErrOrderNotFound) {
		return 0, fmt.Errorf("get order: %w", err)
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			return 200, ErrOrderExistsForUser
		}
		return 0, ErrOrderExistsForOtherUser
	}

	// Create new order
	order := &models.Order{
		Number: orderNumber,
		UserID: userID,
		Status: "NEW",
	}

	err = s.repo.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, repository.ErrOrderExists) {
			return 0, ErrOrderExistsForOtherUser
		}
		return 0, fmt.Errorf("create order: %w", err)
	}

	return 202, nil
}

func (s *OrderServiceImpl) GetUserOrders(ctx context.Context, userID string) ([]*models.Order, error) {
	orders, err := s.repo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user orders: %w", err)
	}

	return orders, nil
}

func (s *OrderServiceImpl) ProcessOrders(ctx context.Context) error {
	if s.accrualService == nil {
		return fmt.Errorf("accrual service not configured")
	}

	// Get orders for processing
	orders, err := s.repo.GetOrdersForProcessing(ctx, 10)
	if err != nil {
		return fmt.Errorf("get orders for processing: %w", err)
	}

	for _, order := range orders {
		err := s.processOrder(ctx, order)
		if err != nil {
			// Log error but continue with other orders
			fmt.Printf("Error processing order %s: %v\n", order.Number, err)
			continue
		}
	}

	return nil
}

func (s *OrderServiceImpl) processOrder(ctx context.Context, order *models.Order) error {
	// Get order status from accrual service
	accrualInfo, err := s.accrualService.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		// If service is unavailable, keep order in current status
		return fmt.Errorf("get order accrual: %w", err)
	}

	if accrualInfo == nil {
		// Order not found in accrual system, mark as INVALID
		order.Status = "INVALID"
		return s.repo.UpdateOrder(ctx, order)
	}

	// Update order status
	order.Status = mapAccrualStatus(accrualInfo.Status)
	if accrualInfo.Accrual != nil {
		order.Accrual = accrualInfo.Accrual
	}

	err = s.repo.UpdateOrder(ctx, order)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}

	// If order is processed and has accrual, update user balance
	if order.Status == "PROCESSED" && order.Accrual != nil && *order.Accrual > 0 {
		currentBalance, err := s.repo.GetBalance(ctx, order.UserID)
		if err != nil {
			return fmt.Errorf("get balance: %w", err)
		}

		newCurrent := currentBalance.Current + *order.Accrual
		err = s.repo.UpdateBalance(ctx, order.UserID, newCurrent, currentBalance.Withdrawn)
		if err != nil {
			return fmt.Errorf("update balance: %w", err)
		}
	}

	return nil
}

func mapAccrualStatus(accrualStatus string) string {
	switch accrualStatus {
	case "REGISTERED":
		return "PROCESSING"
	case "PROCESSING":
		return "PROCESSING"
	case "INVALID":
		return "INVALID"
	case "PROCESSED":
		return "PROCESSED"
	default:
		return "PROCESSING"
	}
}
