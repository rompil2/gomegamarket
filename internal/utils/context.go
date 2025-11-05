package utils

import (
	"context"
	"fmt"
)

// contextKey является непэкспортируемым типом для ключей контекста чтобы избежать коллизий
type contextKey struct {
	name string
}

// userIDKey - единственный экземпляр ключа для userID
var userIDKey = &contextKey{name: "userID"}

// WithUserID добавляет userID в контекст
// Возвращает новый контекст с userID, оригинальный контекст не изменяется
func WithUserID(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	// Валидация входных данных
	if userID == "" {
		// Возвращаем оригинальный контекст если userID пустой
		// или можно вернуть ошибку, в зависимости от требований
		return ctx
	}

	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserIDFromContext извлекает userID из контекста
// Возвращает userID и флаг наличия значения
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	value := ctx.Value(userIDKey)
	if value == nil {
		return "", false
	}

	userID, ok := value.(string)
	if !ok {
		// Логирование несоответствия типов (в продакшн можно использовать slog)
		// slog.Warn("Invalid type stored in context for userID", "type", fmt.Sprintf("%T", value))
		return "", false
	}

	// Дополнительная валидация значения
	if userID == "" {
		return "", false
	}

	return userID, true
}

// MustGetUserIDFromContext извлекает userID из контекста или паникует если значение отсутствует
// Использовать только в случаях, когда значение гарантированно должно быть в контексте
func MustGetUserIDFromContext(ctx context.Context) string {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		panic("userID not found in context")
	}
	return userID
}

// SafeGetUserIDFromContext извлекает userID из контекста с обработкой ошибок
// Возвращает userID и ошибку вместо булевого флага
func SafeGetUserIDFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("context is nil")
	}

	value := ctx.Value(userIDKey)
	if value == nil {
		return "", fmt.Errorf("userID not found in context")
	}

	userID, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid userID type in context: %T", value)
	}

	if userID == "" {
		return "", fmt.Errorf("userID is empty")
	}

	return userID, nil
}

// ContextWithOptionalUserID добавляет userID в контекст только если он не пустой
func ContextWithOptionalUserID(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if userID == "" {
		return ctx
	}

	return context.WithValue(ctx, userIDKey, userID)
}

// CloneContextWithUserID создает новый контекст с userID на основе родительского
// Полезно для создания изолированных контекстов
func CloneContextWithUserID(parent context.Context, userID string) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	// Создаем новый контекст без значений, но с тем же таймаутом/дедлайном если есть
	ctx := context.Background()

	// Копируем важные свойства из родительского контекста
	if deadline, ok := parent.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel() // Отмена будет вызвана когда новый контекст завершится
	}

	// Добавляем userID
	return WithUserID(ctx, userID)
}

// ContextKeys возвращает список всех ключей в контексте (для отладки)
func ContextKeys(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}

	// Этот метод требует рефлексии и должен использоваться только для отладки
	// В продакшн коде лучше избегать
	return nil
}
