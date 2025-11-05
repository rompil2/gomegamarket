package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithUserID(t *testing.T) {
	t.Run("add userID to context", func(t *testing.T) {
		ctx := context.Background()
		userID := "user-123"

		ctxWithUser := WithUserID(ctx, userID)
		require.NotNil(t, ctxWithUser)

		retrieved, ok := GetUserIDFromContext(ctxWithUser)
		assert.True(t, ok)
		assert.Equal(t, userID, retrieved)
	})

	t.Run("empty userID returns original context", func(t *testing.T) {
		ctx := context.Background()
		originalAddr := &ctx

		ctxWithEmpty := WithUserID(ctx, "")
		require.NotNil(t, ctxWithEmpty)

		// Контекст должен остаться тем же объектом
		assert.Equal(t, originalAddr, &ctxWithEmpty)

		_, ok := GetUserIDFromContext(ctxWithEmpty)
		assert.False(t, ok)
	})

	t.Run("nil context becomes background", func(t *testing.T) {
		ctx := WithUserID(context.TODO(), "user-123")
		require.NotNil(t, ctx)

		userID, ok := GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "user-123", userID)
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	t.Run("existing userID", func(t *testing.T) {
		ctx := WithUserID(context.Background(), "user-456")

		userID, ok := GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "user-456", userID)
	})

	t.Run("no userID in context", func(t *testing.T) {
		ctx := context.Background()

		userID, ok := GetUserIDFromContext(ctx)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})

	t.Run("nil context", func(t *testing.T) {
		userID, ok := GetUserIDFromContext(context.TODO())
		assert.False(t, ok)
		assert.Empty(t, userID)
	})

	t.Run("wrong type in context", func(t *testing.T) {
		// Создаем контекст с неправильным типом значения
		wrongKey := &contextKey{name: "userID"}                       // Другой экземпляр ключа
		ctx := context.WithValue(context.Background(), wrongKey, 123) // int вместо string

		userID, ok := GetUserIDFromContext(ctx)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})
}

func TestMustGetUserIDFromContext(t *testing.T) {
	t.Run("existing userID", func(t *testing.T) {
		ctx := WithUserID(context.Background(), "user-789")

		userID := MustGetUserIDFromContext(ctx)
		assert.Equal(t, "user-789", userID)
	})

	t.Run("panics when no userID", func(t *testing.T) {
		ctx := context.Background()

		assert.Panics(t, func() {
			MustGetUserIDFromContext(ctx)
		})
	})
}

func TestSafeGetUserIDFromContext(t *testing.T) {
	t.Run("existing userID", func(t *testing.T) {
		ctx := WithUserID(context.Background(), "user-999")

		userID, err := SafeGetUserIDFromContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "user-999", userID)
	})

	t.Run("no userID in context", func(t *testing.T) {
		ctx := context.Background()

		userID, err := SafeGetUserIDFromContext(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Empty(t, userID)
	})

	t.Run("nil context", func(t *testing.T) {
		userID, err := SafeGetUserIDFromContext(context.TODO())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
		assert.Empty(t, userID)
	})

	t.Run("wrong value type", func(t *testing.T) {
		// Создаем контекст с неправильным типом используя тот же ключ
		type otherKey struct{ name string }
		wrongKey := &otherKey{name: "userID"}
		ctx := context.WithValue(context.Background(), wrongKey, 456)

		userID, err := SafeGetUserIDFromContext(ctx)
		assert.NoError(t, err) // Разные ключи - не конфликтуют
		assert.Empty(t, userID)
	})
}

func TestContextWithOptionalUserID(t *testing.T) {
	t.Run("with userID", func(t *testing.T) {
		ctx := ContextWithOptionalUserID(context.Background(), "user-111")

		userID, ok := GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "user-111", userID)
	})

	t.Run("empty userID returns original", func(t *testing.T) {
		original := context.Background()
		ctx := ContextWithOptionalUserID(original, "")

		assert.Equal(t, &original, &ctx) // Тот же объект контекста
	})
}

func TestCloneContextWithUserID(t *testing.T) {
	t.Run("clone with userID", func(t *testing.T) {
		parent := context.Background()
		ctx := CloneContextWithUserID(parent, "user-222")

		userID, ok := GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "user-222", userID)
	})

	t.Run("nil parent", func(t *testing.T) {
		ctx := CloneContextWithUserID(context.TODO(), "user-333")

		userID, ok := GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "user-333", userID)
	})
}

func TestContextIsolation(t *testing.T) {
	t.Run("context values are isolated", func(t *testing.T) {
		parent := context.Background()

		child1 := WithUserID(parent, "user-1")
		child2 := WithUserID(parent, "user-2")

		// Каждый child имеет свое значение
		user1, ok1 := GetUserIDFromContext(child1)
		user2, ok2 := GetUserIDFromContext(child2)

		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.Equal(t, "user-1", user1)
		assert.Equal(t, "user-2", user2)

		// Parent остается неизменным
		_, ok := GetUserIDFromContext(parent)
		assert.False(t, ok)
	})
}

// Benchmark тесты для измерения производительности
func BenchmarkWithUserID(b *testing.B) {
	ctx := context.Background()
	userID := "benchmark-user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithUserID(ctx, userID)
	}
}

func BenchmarkGetUserIDFromContext(b *testing.B) {
	ctx := WithUserID(context.Background(), "benchmark-user")

	for b.Loop() {
		GetUserIDFromContext(ctx)
	}
}

// Пример использования в документации
func ExampleWithUserID() {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user-123")

	userID, ok := GetUserIDFromContext(ctx)
	if ok {
		// Использовать userID
		_ = userID
	}
}
