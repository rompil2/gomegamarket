package utils

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() returned empty hash")
	}

	if hash == password {
		t.Error("HashPassword() returned plain text password")
	}

	// Проверяем формат хеша
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Error("Hash should start with $argon2id$")
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("Hash should have 6 parts, got %d", len(parts))
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testPassword123"
	wrongPassword := "wrongPassword456"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Setup: HashPassword() error = %v", err)
	}

	t.Run("correct password", func(t *testing.T) {
		if !CheckPasswordHash(password, hash) {
			t.Error("CheckPasswordHash() should return true for correct password")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		if CheckPasswordHash(wrongPassword, hash) {
			t.Error("CheckPasswordHash() should return false for wrong password")
		}
	})

	t.Run("empty password", func(t *testing.T) {
		if CheckPasswordHash("", hash) {
			t.Error("CheckPasswordHash() should return false for empty password")
		}
	})

	t.Run("invalid hash format", func(t *testing.T) {
		if CheckPasswordHash(password, "invalid_hash") {
			t.Error("CheckPasswordHash() should return false for invalid hash format")
		}
	})

	t.Run("different algorithm", func(t *testing.T) {
		// Поддельный хеш с другим алгоритмом
		fakeHash := "$bcrypt$v=1$m=65536,t=2,p=1$c2FsdA$hash"
		if CheckPasswordHash(password, fakeHash) {
			t.Error("CheckPasswordHash() should return false for different algorithm")
		}
	})
}

func TestPasswordSecurity(t *testing.T) {
	// Тест на то, что одинаковые пароли дают разные хеши
	password := "samePassword"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword() error = %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword() error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("Same password should produce different hashes due to different salts")
	}

	// Оба хеша должны проверяться успешно
	if !CheckPasswordHash(password, hash1) {
		t.Error("First hash should validate correctly")
	}

	if !CheckPasswordHash(password, hash2) {
		t.Error("Second hash should validate correctly")
	}
}

func TestHashPassword_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"empty password", "", false},
		{"very long password", strings.Repeat("a", 10000), false},
		{"special chars", "p@$$w0rd!№%", false},
		{"unicode", "пароль123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash for valid password")
			}

			// Проверяем, что пароль можно верифицировать
			if !tt.wantErr && !CheckPasswordHash(tt.password, hash) {
				t.Error("CheckPasswordHash() failed for the same password")
			}
		})
	}
}
