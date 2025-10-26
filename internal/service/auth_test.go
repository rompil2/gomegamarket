package service

import (
	"context"
	"testing"

	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/repository"
)

type mockUserRepo struct {
	users map[string]*models.User
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	if _, exists := m.users[user.Login]; exists {
		return repository.ErrUserExists
	}
	m.users[user.Login] = user
	user.ID = "user-" + user.Login
	return nil
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	user, exists := m.users[login]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := &mockUserRepo{users: make(map[string]*models.User)}
	authService := NewAuthService(mockRepo, "test-secret")

	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		user, token, err := authService.Register(ctx, "newuser", "password123")
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		if user.Login != "newuser" {
			t.Errorf("Register() got user.Login = %v, want %v", user.Login, "newuser")
		}

		if token == "" {
			t.Error("Register() returned empty token")
		}

		if user.ID == "" {
			t.Error("Register() should set user ID")
		}
	})

	t.Run("duplicate registration", func(t *testing.T) {
		_, _, err := authService.Register(ctx, "newuser", "password123")
		if err != ErrUserExists {
			t.Errorf("Register() error = %v, want %v", err, ErrUserExists)
		}
	})

	t.Run("empty login", func(t *testing.T) {
		_, _, err := authService.Register(ctx, "", "password123")
		if err == nil {
			t.Error("Register() should return error for empty login")
		}
	})

	t.Run("short password", func(t *testing.T) {
		_, _, err := authService.Register(ctx, "user2", "123")
		if err == nil {
			t.Error("Register() should return error for short password")
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := &mockUserRepo{users: make(map[string]*models.User)}
	authService := NewAuthService(mockRepo, "test-secret")

	ctx := context.Background()

	// First register a user
	_, _, err := authService.Register(ctx, "testuser", "password123")
	if err != nil {
		t.Fatalf("Setup: Register() error = %v", err)
	}

	t.Run("successful login", func(t *testing.T) {
		user, token, err := authService.Login(ctx, "testuser", "password123")
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}

		if user.Login != "testuser" {
			t.Errorf("Login() got user.Login = %v, want %v", user.Login, "testuser")
		}

		if token == "" {
			t.Error("Login() returned empty token")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, _, err := authService.Login(ctx, "testuser", "wrongpassword")
		if err != ErrInvalidCredentials {
			t.Errorf("Login() error = %v, want %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, _, err := authService.Login(ctx, "nonexistent", "password")
		if err != ErrInvalidCredentials {
			t.Errorf("Login() error = %v, want %v", err, ErrInvalidCredentials)
		}
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := &mockUserRepo{users: make(map[string]*models.User)}
	authService := NewAuthService(mockRepo, "test-secret")

	ctx := context.Background()

	// Register and get token
	user, token, err := authService.Register(ctx, "tokenuser", "password123")
	if err != nil {
		t.Fatalf("Setup: Register() error = %v", err)
	}

	t.Run("valid token", func(t *testing.T) {
		userID, err := authService.ValidateToken(token)
		if err != nil {
			t.Fatalf("ValidateToken() error = %v", err)
		}

		if userID != user.ID {
			t.Errorf("ValidateToken() got userID = %v, want %v", userID, user.ID)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := authService.ValidateToken("invalid-token")
		if err == nil {
			t.Error("ValidateToken() should return error for invalid token")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := authService.ValidateToken("")
		if err == nil {
			t.Error("ValidateToken() should return error for empty token")
		}
	})
}
