package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rompil2/gomegamarket/internal/service"
)

func (h *Handler) UploadOrder(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Read order number from request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	orderNumber := string(body)
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order number is required"})
		return
	}

	// Upload order
	status, err := h.service.UploadOrder(c.Request.Context(), userID.(string), orderNumber)
	if err != nil {
		switch err {
		case service.ErrInvalidOrderNumber:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order number"})
		case service.ErrOrderExistsForOtherUser:
			c.JSON(http.StatusConflict, gin.H{"error": "Order already exists for another user"})
		case service.ErrOrderExistsForUser:
			c.Status(http.StatusOK)
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.Status(status)
}

func (h *Handler) GetOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orders, err := h.service.GetUserOrders(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, orders)
}
