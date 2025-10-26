package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccrualClient_GetOrderAccrual(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/orders/valid-order":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"order":"valid-order","status":"PROCESSED","accrual":500}`))
		case "/api/orders/processing-order":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"order":"processing-order","status":"PROCESSING"}`))
		case "/api/orders/not-found":
			w.WriteHeader(http.StatusNoContent)
		case "/api/orders/rate-limit":
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)

	tests := []struct {
		name        string
		orderNumber string
		wantStatus  string
		wantAccrual float64
		wantErr     bool
	}{
		{
			name:        "processed order",
			orderNumber: "valid-order",
			wantStatus:  "PROCESSED",
			wantAccrual: 500,
			wantErr:     false,
		},
		{
			name:        "processing order",
			orderNumber: "processing-order",
			wantStatus:  "PROCESSING",
			wantAccrual: 0,
			wantErr:     false,
		},
		{
			name:        "not found",
			orderNumber: "not-found",
			wantErr:     false, // Not an error, just no data
		},
		{
			name:        "rate limit",
			orderNumber: "rate-limit",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accrual, err := client.GetOrderAccrual(context.Background(), tt.orderNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrderAccrual() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && accrual != nil {
				if accrual.Status != tt.wantStatus {
					t.Errorf("GetOrderAccrual() status = %v, want %v", accrual.Status, tt.wantStatus)
				}

				if tt.wantAccrual > 0 && (accrual.Accrual == nil || *accrual.Accrual != tt.wantAccrual) {
					t.Errorf("GetOrderAccrual() accrual = %v, want %v", accrual.Accrual, tt.wantAccrual)
				}
			}
		})
	}
}
