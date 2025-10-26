package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rompil2/gomegamarket/internal/config"
	"github.com/rompil2/gomegamarket/internal/handlers"
	"github.com/rompil2/gomegamarket/internal/repository/postgres"
	"github.com/rompil2/gomegamarket/internal/service"
	"github.com/rompil2/gomegamarket/internal/service/accrual"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load(os.Args[1:])

	// Настраиваем логгер
	setupLogger()

	slog.Info("Starting Gophermart loyalty system",
		slog.String("address", cfg.RunAddress),
		slog.String("database", maskDBPassword(cfg.DatabaseURI)),
		slog.String("accrual", cfg.AccrualAddress),
	)

	// Инициализируем репозиторий
	repo, err := postgres.New(cfg.DatabaseURI)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()
	slog.Info("Database connection established")

	// Инициализируем сервис начислений
	var accrualService service.AccrualService
	if cfg.AccrualAddress != "" {
		accrualService = accrual.NewClient(cfg.AccrualAddress)
		slog.Info("Accrual service configured", "address", cfg.AccrualAddress)
	} else {
		slog.Warn("Accrual service address not provided, order processing will be disabled")
	}

	// Инициализируем сервисный слой
	svc := service.NewService(repo, accrualService, cfg.JWTSecret)

	// Инициализируем Gin router
	router := gin.New()

	// Middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(handlers.GinLogger())
	router.Use(handlers.ErrorHandlingMiddleware())

	// Настраиваем обработчики
	handler := handlers.NewHandler(svc)
	handler.SetupRoutes(router)

	// Создаем HTTP сервер с Gin
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	// Канал для graceful shutdown
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем горутину для обработки заказов
	if accrualService != nil {
		go startOrderProcessing(context.Background(), svc)
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		slog.Info("Starting HTTP server", "address", cfg.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			quit <- syscall.SIGTERM
		}
	}()

	// Ожидаем сигнал завершения
	<-quit
	slog.Info("Shutdown signal received")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Could not gracefully shutdown the server", "error", err)
	}

	slog.Info("Server stopped")
	done <- true
}

// setupLogger настраивает глобальный логгер
func setupLogger() {
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		logLevel = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logLevel == slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// startOrderProcessing запускает периодическую обработку заказов
func startOrderProcessing(ctx context.Context, svc service.Service) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	slog.Info("Starting order processing worker")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Order processing worker stopped")
			return
		case <-ticker.C:
			if err := svc.ProcessOrders(ctx); err != nil {
				slog.Error("Failed to process orders", "error", err)
			}
		}
	}
}

// maskDBPassword маскирует пароль в строке подключения к БД для логов
func maskDBPassword(dbURI string) string {
	return dbURI
}

// init выполняется при инициализации пакета
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
