package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rompil2/gomegamarket/internal/service"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Public routes
	public := router.Group("/api/user")
	{
		public.POST("/register", h.Register)
		public.POST("/login", h.Login)
	}

	// Protected routes
	protected := router.Group("/api/user")
	protected.Use(h.AuthMiddleware())
	{
		protected.POST("/orders", h.UploadOrder)
		protected.GET("/orders", h.GetOrders)
		protected.GET("/balance", h.GetBalance)
		protected.POST("/balance/withdraw", h.Withdraw)
		protected.GET("/withdrawals", h.GetWithdrawals)
	}
}
