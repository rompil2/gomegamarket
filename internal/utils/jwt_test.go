package utils

import (
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {
	userID := "test-user-123"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, secret)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateJWT() returned empty token")
	}

	// Проверяем, что токен содержит три части (header.payload.signature)
	parts := splitToken(token)
	if len(parts) != 3 {
		t.Errorf("JWT should have 3 parts, got %d", len(parts))
	}
}

func TestValidateJWT(t *testing.T) {
	userID := "test-user-456"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, secret)
	if err != nil {
		t.Fatalf("Setup: GenerateJWT() error = %v", err)
	}

	t.Run("valid token", func(t *testing.T) {
		claims, err := ValidateJWT(token, secret)
		if err != nil {
			t.Fatalf("ValidateJWT() error = %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("ValidateJWT() got userID = %v, want %v", claims.UserID, userID)
		}

		if claims.Issuer != "gophermart" {
			t.Errorf("ValidateJWT() got issuer = %v, want %v", claims.Issuer, "gophermart")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		_, err := ValidateJWT(token, "wrong-secret")
		if err == nil {
			t.Error("ValidateJWT() should return error for wrong secret")
		}
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := ValidateJWT("invalid.token.format", secret)
		if err == nil {
			t.Error("ValidateJWT() should return error for invalid token format")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := ValidateJWT("", secret)
		if err != ErrInvalidToken {
			t.Errorf("ValidateJWT() error = %v, want %v", err, ErrInvalidToken)
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		_, err := ValidateJWT(token, "")
		if err != ErrInvalidToken {
			t.Errorf("ValidateJWT() error = %v, want %v", err, ErrInvalidToken)
		}
	})
}

func TestJWT_Expiration(t *testing.T) {
	userID := "test-user-789"
	secret := "test-secret-key"

	t.Run("expired token", func(t *testing.T) {
		// Создаем токен с истекшим сроком действия
		token, err := GenerateJWTWithExpiry(userID, secret, -time.Hour)
		if err != nil {
			t.Fatalf("GenerateJWTWithExpiry() error = %v", err)
		}

		_, err = ValidateJWT(token, secret)
		if err != ErrExpiredToken {
			t.Errorf("ValidateJWT() error = %v, want %v", err, ErrExpiredToken)
		}
	})

	t.Run("valid expiry", func(t *testing.T) {
		token, err := GenerateJWTWithExpiry(userID, secret, time.Hour)
		if err != nil {
			t.Fatalf("GenerateJWTWithExpiry() error = %v", err)
		}

		claims, err := ValidateJWT(token, secret)
		if err != nil {
			t.Fatalf("ValidateJWT() error = %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("ValidateJWT() got userID = %v, want %v", claims.UserID, userID)
		}
	})
}

func TestJWT_EdgeCases(t *testing.T) {
	secret := "test-secret"

	t.Run("empty userID", func(t *testing.T) {
		_, err := GenerateJWT("", secret)
		if err == nil {
			t.Error("GenerateJWT() should return error for empty userID")
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		_, err := GenerateJWT("user123", "")
		if err == nil {
			t.Error("GenerateJWT() should return error for empty secret")
		}
	})

	t.Run("very long userID", func(t *testing.T) {
		longUserID := string(make([]byte, 1000))
		token, err := GenerateJWT(longUserID, secret)
		if err != nil {
			t.Fatalf("GenerateJWT() error = %v", err)
		}

		claims, err := ValidateJWT(token, secret)
		if err != nil {
			t.Fatalf("ValidateJWT() error = %v", err)
		}

		if claims.UserID != longUserID {
			t.Error("ValidateJWT() should preserve long userID")
		}
	})
}

func TestGenerateJWTWithExpiry(t *testing.T) {
	userID := "test-user"
	secret := "test-secret"

	token, err := GenerateJWTWithExpiry(userID, secret, 2*time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWTWithExpiry() error = %v", err)
	}

	claims, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	// Проверяем, что токен действителен
	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}

	// Проверяем, что время истечения примерно через 2 часа
	expectedExpiry := time.Now().Add(2 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	diff := actualExpiry.Sub(expectedExpiry)

	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("Token expiry is too far from expected: %v", diff)
	}
}

// Вспомогательная функция для разделения JWT токена
func splitToken(token string) []string {
	// Простая реализация - в реальном коде используйте jwt.Parse
	// Это только для базовой проверки структуры
	var parts []string
	start := 0
	for i, char := range token {
		if char == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	if start < len(token) {
		parts = append(parts, token[start:])
	}
	return parts
}
