package service

import (
	"github.com/rompil2/gomegamarket/internal/repository"
)

type ServiceImpl struct {
	*AuthServiceImpl
	*OrderServiceImpl
	*BalanceServiceImpl
}

func NewService(repo repository.Repository, accrualService AccrualService, jwtSecret string) *ServiceImpl {
	authService := NewAuthService(repo, jwtSecret)
	orderService := NewOrderService(repo, accrualService)
	balanceService := NewBalanceService(repo)

	return &ServiceImpl{
		AuthServiceImpl:    authService,
		OrderServiceImpl:   orderService,
		BalanceServiceImpl: balanceService,
	}
}
