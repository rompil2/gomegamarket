package models

import (
	"testing"
)

func TestOrder_Validate(t *testing.T) {

	tests := []struct {
		name    string
		order   Order
		wantErr bool
	}{
		{
			name: "valid order",
			order: Order{
				Number: "79927398713",
				Status: "NEW",
			},
			wantErr: false,
		},
		{
			name: "invalid order number",
			order: Order{
				Number: "123",
				Status: "NEW",
			},
			wantErr: true,
		},
		{
			name: "empty status",
			order: Order{
				Number: "79927398713",
				Status: "",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			order: Order{
				Number: "79927398713",
				Status: "UNKNOWN",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Order.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
