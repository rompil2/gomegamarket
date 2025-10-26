package utils

import "testing"

func TestValidLuhn(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		// Valid numbers
		{"valid visa", "4111111111111111", true},
		{"valid mastercard", "5555555555554444", true},
		{"valid amex", "378282246310005", true},
		{"valid number 1", "79927398713", true},
		{"valid number 2", "49927398716", true},
		{"valid number 3", "1234567812345670", true},
		{"single zero", "0", true},

		// Invalid numbers
		{"invalid number 1", "79927398710", false},
		{"invalid number 2", "49927398717", false},
		{"invalid number 3", "1234567812345678", false},
		{"empty string", "", false},
		{"non-digit chars", "7992a398713", false},
		{"special chars", "7992-3987-13", false},
		{"spaces", "7992 7398 713", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidLuhn(tt.number)
			if result != tt.expected {
				t.Errorf("ValidLuhn(%q) = %v, want %v", tt.number, result, tt.expected)
			}
		})
	}
}

func BenchmarkValidLuhn(b *testing.B) {
	numbers := []string{
		"4111111111111111",
		"79927398713",
		"1234567812345670",
		"49927398716",
	}

	for i := 0; i < b.N; i++ {
		for _, number := range numbers {
			ValidLuhn(number)
		}
	}
}
