package models

import (
	"errors"
	"time"
)

type Order struct {
	Number     string    `json:"number" db:"number"`
	UserID     string    `json:"user_id" db:"user_id"`
	Status     string    `json:"status" db:"status"`
	Accrual    *float64  `json:"accrual,omitempty" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

func (o Order) Validate() error {
	if o.Number == "" {
		return errors.New("order number is required")
	}
	return nil
}

type OrderAccrual struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}
