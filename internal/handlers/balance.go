package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/service"
)

func (h *Handler) GetBalance(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	balance, err := h.service.GetBalance(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, balance)
}

func (h *Handler) Withdraw(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Order == "" || req.Sum <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order number and positive sum are required"})
		return
	}

	fmt.Printf("Withdraw request: userID=%v, order=%s, sum=%f\n", userID, req.Order, req.Sum)
	err := h.service.Withdraw(c.Request.Context(), userID.(string), &req)
	if err != nil {
		fmt.Printf("Withdraw error: %v\n", err)
		switch err {
		case service.ErrInvalidOrderNumber:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order number"})
		case service.ErrInsufficientBalance:
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient balance"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) GetWithdrawals(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	withdrawals, err := h.service.GetWithdrawals(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if len(withdrawals) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}
