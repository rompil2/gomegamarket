package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет JWT токен и добавляет userID в контекст
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из cookie или Authorization header
		var token string

		// Пробуем получить из cookie
		if cookie, err := c.Cookie("token"); err == nil {
			token = cookie
		}

		// Если нет в cookie, пробуем из Authorization header
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Валидируем токен
		userID, err := h.service.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Добавляем userID в контекст Gin
		c.Set("userID", userID)
		c.Next()
	}
}

// ErrorHandlingMiddleware централизованная обработка ошибок
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
	}
}

// GinLogger кастомный логгер для Gin
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		timestamp := time.Now()
		latency := timestamp.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		slog.Info("HTTP request",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.String("ip", clientIP),
			slog.String("latency", latency.String()),
			slog.String("error", errorMessage),
		)
	}
}
