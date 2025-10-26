package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/service"
	"github.com/stretchr/testify/assert"
)

// MockService представляет мок сервиса для тестирования
type MockService struct {
	RegisterFunc       func(ctx context.Context, login, password string) (*models.User, string, error)
	LoginFunc          func(ctx context.Context, login, password string) (*models.User, string, error)
	ValidateTokenFunc  func(token string) (string, error)
	UploadOrderFunc    func(ctx context.Context, userID, orderNumber string) (int, error)
	GetUserOrdersFunc  func(ctx context.Context, userID string) ([]*models.Order, error)
	GetBalanceFunc     func(ctx context.Context, userID string) (*models.Balance, error)
	WithdrawFunc       func(ctx context.Context, userID string, req *models.WithdrawRequest) error
	GetWithdrawalsFunc func(ctx context.Context, userID string) ([]*models.Withdrawal, error)
	ProcessOrdersFunc  func(ctx context.Context) error
}

func (m *MockService) Register(ctx context.Context, login, password string) (*models.User, string, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, login, password)
	}
	return &models.User{ID: "123", Login: login}, "test-token", nil
}

func (m *MockService) Login(ctx context.Context, login, password string) (*models.User, string, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, login, password)
	}
	return &models.User{ID: "123", Login: login}, "test-token", nil
}

func (m *MockService) ValidateToken(token string) (string, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(token)
	}
	if token == "valid-token" {
		return "user123", nil
	}
	return "", service.ErrInvalidCredentials
}

func (m *MockService) UploadOrder(ctx context.Context, userID, orderNumber string) (int, error) {
	if m.UploadOrderFunc != nil {
		return m.UploadOrderFunc(ctx, userID, orderNumber)
	}
	return 202, nil
}

func (m *MockService) GetUserOrders(ctx context.Context, userID string) ([]*models.Order, error) {
	if m.GetUserOrdersFunc != nil {
		return m.GetUserOrdersFunc(ctx, userID)
	}
	return []*models.Order{
		{Number: "12345678903", Status: "PROCESSED", Accrual: floatPtr(500.0)},
	}, nil
}

func (m *MockService) GetBalance(ctx context.Context, userID string) (*models.Balance, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, userID)
	}
	return &models.Balance{Current: 500.5, Withdrawn: 42}, nil
}

func (m *MockService) Withdraw(ctx context.Context, userID string, req *models.WithdrawRequest) error {
	if m.WithdrawFunc != nil {
		return m.WithdrawFunc(ctx, userID, req)
	}
	return nil
}

func (m *MockService) GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
	if m.GetWithdrawalsFunc != nil {
		return m.GetWithdrawalsFunc(ctx, userID)
	}
	return []*models.Withdrawal{
		{Order: "456", Sum: 100, ProcessedAt: time.Now()},
	}, nil
}

func (m *MockService) ProcessOrders(ctx context.Context) error {
	if m.ProcessOrdersFunc != nil {
		return m.ProcessOrdersFunc(ctx)
	}
	return nil
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestNewHandler(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
}

func TestHandler_SetupRoutes(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService)

	// Создаем Gin router в тестовом режиме
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Настраиваем маршруты
	handler.SetupRoutes(router)

	// Проверяем что маршруты зарегистрированы
	routes := router.Routes()
	assert.NotEmpty(t, routes)

	// Проверяем наличие основных маршрутов
	expectedRoutes := []string{
		"POST /api/user/register",
		"POST /api/user/login",
		"POST /api/user/orders",
		"GET /api/user/orders",
		"GET /api/user/balance",
		"POST /api/user/balance/withdraw",
		"GET /api/user/withdrawals",
	}

	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route.Method+" "+route.Path] = true
	}

	for _, expectedRoute := range expectedRoutes {
		assert.True(t, routeMap[expectedRoute], "Route %s should be registered", expectedRoute)
	}
}

func TestHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful registration",
			requestBody: models.LoginRequest{
				Login:    "newuser",
				Password: "password123",
			},
			mockSetup: func(ms *MockService) {
				ms.RegisterFunc = func(ctx context.Context, login, password string) (*models.User, string, error) {
					return &models.User{ID: "user-123", Login: login}, "auth-token", nil
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Header().Get("Set-Cookie"), "token=auth-token")

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "user-123", response["id"])
				assert.Equal(t, "newuser", response["login"])
			},
		},
		{
			name: "registration with existing user",
			requestBody: models.LoginRequest{
				Login:    "existinguser",
				Password: "password123",
			},
			mockSetup: func(ms *MockService) {
				ms.RegisterFunc = func(ctx context.Context, login, password string) (*models.User, string, error) {
					return nil, "", service.ErrUserExists
				}
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "User already exists", response["error"])
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    "{invalid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid request format")
			},
		},
		{
			name: "empty login",
			requestBody: models.LoginRequest{
				Login:    "",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty password",
			requestBody: models.LoginRequest{
				Login:    "user",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			var bodyBytes []byte
			switch body := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(body)
			default:
				bodyBytes, _ = json.Marshal(body)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.LoginRequest
		mockSetup      func(*MockService)
		expectedStatus int
	}{
		{
			name: "successful login",
			requestBody: models.LoginRequest{
				Login:    "testuser",
				Password: "password123",
			},
			mockSetup: func(ms *MockService) {
				ms.LoginFunc = func(ctx context.Context, login, password string) (*models.User, string, error) {
					return &models.User{ID: "user-123", Login: login}, "login-token", nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			requestBody: models.LoginRequest{
				Login:    "wronguser",
				Password: "wrongpass",
			},
			mockSetup: func(ms *MockService) {
				ms.LoginFunc = func(ctx context.Context, login, password string) (*models.User, string, error) {
					return nil, "", service.ErrInvalidCredentials
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			bodyBytes, _ := json.Marshal(tt.requestBody)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_UploadOrder(t *testing.T) {
	tests := []struct {
		name           string
		orderNumber    string
		token          string
		mockSetup      func(*MockService)
		expectedStatus int
	}{
		{
			name:        "successful upload with valid token",
			orderNumber: "79927398713",
			token:       "valid-token",
			mockSetup: func(ms *MockService) {
				ms.UploadOrderFunc = func(ctx context.Context, userID, orderNumber string) (int, error) {
					assert.Equal(t, "user123", userID)
					return 202, nil
				}
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "upload without token",
			orderNumber:    "79927398713",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "upload with invalid token",
			orderNumber: "79927398713",
			token:       "invalid-token",
			mockSetup: func(ms *MockService) {
				ms.ValidateTokenFunc = func(token string) (string, error) {
					return "", service.ErrInvalidCredentials
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "invalid order number",
			orderNumber: "invalid",
			token:       "valid-token",
			mockSetup: func(ms *MockService) {
				ms.UploadOrderFunc = func(ctx context.Context, userID, orderNumber string) (int, error) {
					return 0, service.ErrInvalidOrderNumber
				}
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/user/orders", bytes.NewReader([]byte(tt.orderNumber)))
			req.Header.Set("Content-Type", "text/plain")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_GetOrders(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		mockSetup      func(*MockService)
		expectedStatus int
	}{
		{
			name:  "get orders with valid token",
			token: "valid-token",
			mockSetup: func(ms *MockService) {
				ms.GetUserOrdersFunc = func(ctx context.Context, userID string) ([]*models.Order, error) {
					return []*models.Order{
						{Number: "123", Status: "PROCESSED"},
						{Number: "456", Status: "PROCESSING"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get orders without token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "no orders found",
			token: "valid-token",
			mockSetup: func(ms *MockService) {
				ms.GetUserOrdersFunc = func(ctx context.Context, userID string) ([]*models.Order, error) {
					return []*models.Order{}, nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/user/orders", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_GetBalance(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		mockSetup      func(*MockService)
		expectedStatus int
	}{
		{
			name:  "get balance with valid token",
			token: "valid-token",
			mockSetup: func(ms *MockService) {
				ms.GetBalanceFunc = func(ctx context.Context, userID string) (*models.Balance, error) {
					return &models.Balance{Current: 1000.5, Withdrawn: 200}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get balance without token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/user/balance", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var balance models.Balance
				err := json.Unmarshal(w.Body.Bytes(), &balance)
				assert.NoError(t, err)
				assert.Equal(t, 1000.5, balance.Current)
				assert.Equal(t, 200.0, balance.Withdrawn)
			}
		})
	}
}

func TestHandler_Withdraw(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		requestBody    interface{}
		mockSetup      func(*MockService)
		expectedStatus int
	}{
		{
			name:  "successful withdraw",
			token: "valid-token",
			requestBody: models.WithdrawRequest{
				Order: "79927398713",
				Sum:   100,
			},
			mockSetup: func(ms *MockService) {
				ms.WithdrawFunc = func(ctx context.Context, userID string, req *models.WithdrawRequest) error {
					assert.Equal(t, "user123", userID)
					assert.Equal(t, "79927398713", req.Order)
					assert.Equal(t, 100.0, req.Sum)
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "withdraw without token",
			token:          "",
			requestBody:    models.WithdrawRequest{Order: "123", Sum: 100},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "insufficient balance",
			token: "valid-token",
			requestBody: models.WithdrawRequest{
				Order: "79927398713",
				Sum:   5000,
			},
			mockSetup: func(ms *MockService) {
				ms.WithdrawFunc = func(ctx context.Context, userID string, req *models.WithdrawRequest) error {
					return service.ErrInsufficientBalance
				}
			},
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:           "invalid JSON",
			token:          "valid-token",
			requestBody:    "{invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			var bodyBytes []byte
			switch body := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(body)
			default:
				bodyBytes, _ = json.Marshal(body)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_GetWithdrawals(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		mockSetup      func(*MockService)
		expectedStatus int
	}{
		{
			name:  "get withdrawals with valid token",
			token: "valid-token",
			mockSetup: func(ms *MockService) {
				ms.GetWithdrawalsFunc = func(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
					return []*models.Withdrawal{
						{Order: "123", Sum: 50, ProcessedAt: time.Now()},
						{Order: "456", Sum: 100, ProcessedAt: time.Now()},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get withdrawals without token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "no withdrawals found",
			token: "valid-token",
			mockSetup: func(ms *MockService) {
				ms.GetWithdrawalsFunc = func(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
					return []*models.Withdrawal{}, nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := NewHandler(mockService)
			router := setupTestRouter(handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/user/withdrawals", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	mockService := &MockService{}
	handler := NewHandler(mockService)

	router := gin.New()
	protected := router.Group("/api/protected")
	protected.Use(handler.AuthMiddleware())
	protected.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if exists {
			c.JSON(http.StatusOK, gin.H{"userID": userID})
		} else {
			c.JSON(http.StatusOK, gin.H{"userID": "none"})
		}
	})

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "valid token",
			token:          "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "no token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/protected/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "user123", response["userID"])
			}
		})
	}
}

// Benchmark тесты
func BenchmarkHandler_Register(b *testing.B) {
	mockService := &MockService{}
	handler := NewHandler(mockService)
	router := setupTestRouter(handler)

	body, _ := json.Marshal(models.LoginRequest{
		Login:    "benchuser",
		Password: "benchpass",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_GetBalance(b *testing.B) {
	mockService := &MockService{}
	handler := NewHandler(mockService)
	router := setupTestRouter(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/user/balance", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		router.ServeHTTP(w, req)
	}
}

// Вспомогательная функция для настройки тестового router
func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.SetupRoutes(router)
	return router
}

// TestMain для настройки тестового окружения
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}
