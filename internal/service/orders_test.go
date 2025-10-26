package service

import (
	"context"
	"testing"

	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/repository"
)

type mockRepo struct {
	repository.Repository
	orders      map[string]*models.Order
	users       map[string]*models.User
	balances    map[string]*models.Balance
	withdrawals []*models.Withdrawal
}

func (m *mockRepo) CreateOrder(ctx context.Context, order *models.Order) error {
	if _, exists := m.orders[order.Number]; exists {
		return repository.ErrOrderExists
	}
	m.orders[order.Number] = order
	return nil
}

func (m *mockRepo) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	order, exists := m.orders[number]
	if !exists {
		return nil, repository.ErrOrderNotFound
	}
	return order, nil
}

func (m *mockRepo) GetUserOrders(ctx context.Context, userID string) ([]*models.Order, error) {
	var userOrders []*models.Order
	for _, order := range m.orders {
		if order.UserID == userID {
			userOrders = append(userOrders, order)
		}
	}
	return userOrders, nil
}

func (m *mockRepo) GetOrdersForProcessing(ctx context.Context, limit int) ([]*models.Order, error) {
	var processingOrders []*models.Order
	for _, order := range m.orders {
		if order.Status == "NEW" || order.Status == "PROCESSING" {
			processingOrders = append(processingOrders, order)
			if len(processingOrders) >= limit {
				break
			}
		}
	}
	return processingOrders, nil
}

func (m *mockRepo) UpdateOrder(ctx context.Context, order *models.Order) error {
	if _, exists := m.orders[order.Number]; !exists {
		return repository.ErrOrderNotFound
	}
	m.orders[order.Number] = order
	return nil
}

func (m *mockRepo) GetBalance(ctx context.Context, userID string) (*models.Balance, error) {
	balance, exists := m.balances[userID]
	if !exists {
		return &models.Balance{Current: 0, Withdrawn: 0}, nil
	}
	return balance, nil
}

func (m *mockRepo) UpdateBalance(ctx context.Context, userID string, current, withdrawn float64) error {
	m.balances[userID] = &models.Balance{
		Current:   current,
		Withdrawn: withdrawn,
	}
	return nil
}

// Implement other required repository methods with empty implementations
func (m *mockRepo) CreateUser(ctx context.Context, user *models.User) error          { return nil }
func (m *mockRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) { return nil, nil }
func (m *mockRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return nil, nil
}
func (m *mockRepo) CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error {
	return nil
}
func (m *mockRepo) GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
	return nil, nil
}
func (m *mockRepo) BeginTx(ctx context.Context) (repository.Transaction, error) { return nil, nil }
func (m *mockRepo) Close() error                                                { return nil }

type mockAccrualService struct {
	orders map[string]*models.OrderAccrual
}

func (m *mockAccrualService) GetOrderAccrual(ctx context.Context, orderNumber string) (*models.OrderAccrual, error) {
	return m.orders[orderNumber], nil
}

func TestOrderService_UploadOrder(t *testing.T) {
	mockRepo := &mockRepo{
		orders:   make(map[string]*models.Order),
		balances: make(map[string]*models.Balance),
	}
	orderService := NewOrderService(mockRepo, nil)

	ctx := context.Background()
	userID := "user123"

	t.Run("valid order", func(t *testing.T) {
		status, err := orderService.UploadOrder(ctx, userID, "79927398713")
		if err != nil {
			t.Fatalf("UploadOrder() error = %v", err)
		}

		if status != 202 {
			t.Errorf("UploadOrder() status = %v, want %v", status, 202)
		}

		// Verify order was created
		order, err := mockRepo.GetOrderByNumber(ctx, "79927398713")
		if err != nil {
			t.Fatalf("Order was not created: %v", err)
		}

		if order.UserID != userID {
			t.Errorf("Order UserID = %v, want %v", order.UserID, userID)
		}

		if order.Status != "NEW" {
			t.Errorf("Order Status = %v, want %v", order.Status, "NEW")
		}
	})

	t.Run("duplicate order for same user", func(t *testing.T) {
		status, err := orderService.UploadOrder(ctx, userID, "79927398713")
		if err != ErrOrderExistsForUser {
			t.Errorf("UploadOrder() error = %v, want %v", err, ErrOrderExistsForUser)
		}

		if status != 200 {
			t.Errorf("UploadOrder() status = %v, want %v", status, 200)
		}
	})

	t.Run("duplicate order for different user", func(t *testing.T) {
		// Create order for first user
		orderService.UploadOrder(ctx, "user1", "1234567812345670")

		// Try to create same order for different user
		_, err := orderService.UploadOrder(ctx, "user2", "1234567812345670")
		if err != ErrOrderExistsForOtherUser {
			t.Errorf("UploadOrder() error = %v, want %v", err, ErrOrderExistsForOtherUser)
		}
	})

	t.Run("invalid order number", func(t *testing.T) {
		_, err := orderService.UploadOrder(ctx, userID, "123")
		if err != ErrInvalidOrderNumber {
			t.Errorf("UploadOrder() error = %v, want %v", err, ErrInvalidOrderNumber)
		}
	})

	t.Run("empty order number", func(t *testing.T) {
		_, err := orderService.UploadOrder(ctx, userID, "")
		if err != ErrInvalidOrderNumber {
			t.Errorf("UploadOrder() error = %v, want %v", err, ErrInvalidOrderNumber)
		}
	})
}

func TestOrderService_GetUserOrders(t *testing.T) {
	mockRepo := &mockRepo{
		orders: make(map[string]*models.Order),
	}
	orderService := NewOrderService(mockRepo, nil)

	ctx := context.Background()
	userID := "user123"

	// Create test orders
	mockRepo.orders["order1"] = &models.Order{Number: "order1", UserID: userID, Status: "PROCESSED"}
	mockRepo.orders["order2"] = &models.Order{Number: "order2", UserID: userID, Status: "PROCESSING"}
	mockRepo.orders["order3"] = &models.Order{Number: "order3", UserID: "other-user", Status: "PROCESSED"}

	t.Run("get user orders", func(t *testing.T) {
		orders, err := orderService.GetUserOrders(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserOrders() error = %v", err)
		}

		if len(orders) != 2 {
			t.Errorf("GetUserOrders() got %v orders, want %v", len(orders), 2)
		}

		// Verify only user's orders are returned
		for _, order := range orders {
			if order.UserID != userID {
				t.Errorf("GetUserOrders() returned order with wrong UserID: %v", order.UserID)
			}
		}
	})

	t.Run("get orders for user with no orders", func(t *testing.T) {
		orders, err := orderService.GetUserOrders(ctx, "no-orders-user")
		if err != nil {
			t.Fatalf("GetUserOrders() error = %v", err)
		}

		if len(orders) != 0 {
			t.Errorf("GetUserOrders() got %v orders, want %v", len(orders), 0)
		}
	})
}

func TestOrderService_ProcessOrders(t *testing.T) {
	mockRepo := &mockRepo{
		orders:   make(map[string]*models.Order),
		balances: make(map[string]*models.Balance),
	}

	mockAccrual := &mockAccrualService{
		orders: map[string]*models.OrderAccrual{
			"processed_order": {
				Order:   "processed_order",
				Status:  "PROCESSED",
				Accrual: floatPtr(500.0),
			},
			"invalid_order": {
				Order:  "invalid_order",
				Status: "INVALID",
			},
			"processing_order": {
				Order:  "processing_order",
				Status: "PROCESSING",
			},
		},
	}

	orderService := NewOrderService(mockRepo, mockAccrual)

	ctx := context.Background()
	userID := "user123"

	// Create test orders
	mockRepo.orders["processed_order"] = &models.Order{
		Number: "processed_order",
		UserID: userID,
		Status: "NEW",
	}
	mockRepo.orders["invalid_order"] = &models.Order{
		Number: "invalid_order",
		UserID: userID,
		Status: "NEW",
	}
	mockRepo.orders["processing_order"] = &models.Order{
		Number: "processing_order",
		UserID: userID,
		Status: "NEW",
	}

	// Set initial balance
	mockRepo.balances[userID] = &models.Balance{Current: 100, Withdrawn: 0}

	err := orderService.ProcessOrders(ctx)
	if err != nil {
		t.Fatalf("ProcessOrders() error = %v", err)
	}

	// Check order statuses were updated
	processedOrder, _ := mockRepo.GetOrderByNumber(ctx, "processed_order")
	if processedOrder.Status != "PROCESSED" {
		t.Errorf("Processed order status = %v, want %v", processedOrder.Status, "PROCESSED")
	}

	invalidOrder, _ := mockRepo.GetOrderByNumber(ctx, "invalid_order")
	if invalidOrder.Status != "INVALID" {
		t.Errorf("Invalid order status = %v, want %v", invalidOrder.Status, "INVALID")
	}

	processingOrder, _ := mockRepo.GetOrderByNumber(ctx, "processing_order")
	if processingOrder.Status != "PROCESSING" {
		t.Errorf("Processing order status = %v, want %v", processingOrder.Status, "PROCESSING")
	}

	// Check balance was updated for processed order
	balance, _ := mockRepo.GetBalance(ctx, userID)
	if balance.Current != 600.0 {
		t.Errorf("Balance after accrual = %v, want %v", balance.Current, 600.0)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
