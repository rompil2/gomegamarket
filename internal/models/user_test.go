package models

import (
	"testing"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				Login:    "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "empty login",
			user: User{
				Login:    "",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			user: User{
				Login:    "testuser",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "short password",
			user: User{
				Login:    "testuser",
				Password: "123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
