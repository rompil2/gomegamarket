package models

import (
	"errors"
	"time"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Login     string    `json:"login" db:"login"`
	Password  string    `json:"password" db:"password_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (u User) Validate() error {
	if u.Login == "" {
		return errors.New("login is required")
	} else if u.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
