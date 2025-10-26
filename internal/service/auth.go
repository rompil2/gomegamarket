package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rompil2/gomegamarket/internal/models"
	"github.com/rompil2/gomegamarket/internal/repository"
	"github.com/rompil2/gomegamarket/internal/utils"
)

type AuthServiceImpl struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewAuthService(repo repository.UserRepository, jwtSecret string) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, login, password string) (*models.User, string, error) {
	// Validate input
	if login == "" || password == "" {
		return nil, "", fmt.Errorf("login and password are required")
	}

	if len(password) < 6 {
		return nil, "", fmt.Errorf("password must be at least 6 characters")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Login:    login,
		Password: hashedPassword,
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return nil, "", ErrUserExists
		}
		return nil, "", fmt.Errorf("create user: %w", err)
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, login, password string) (*models.User, string, error) {
	// Get user by login
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("get user: %w", err)
	}

	// Check password
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	return user, token, nil
}

func (s *AuthServiceImpl) ValidateToken(token string) (string, error) {
	claims, err := utils.ValidateJWT(token, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	return claims.UserID, nil
}
