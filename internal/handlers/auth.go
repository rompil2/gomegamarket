package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/service"
)

func (h *Handler) Register(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Login == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login and password are required"})
		return
	}

	user, token, err := h.service.Register(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Set JWT token as cookie
	c.SetCookie("token", token, 24*3600, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"login": user.Login,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Login == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login and password are required"})
		return
	}

	user, token, err := h.service.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login or password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Set JWT token as cookie
	c.SetCookie("token", token, 24*3600, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"login": user.Login,
	})
}
