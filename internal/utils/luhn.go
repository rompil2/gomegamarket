package utils

// ValidLuhn проверяет номер заказа с помощью алгоритма Луна
func ValidLuhn(number string) bool {
	if number == "" {
		return false
	}

	var sum int
	alternate := false

	// Проходим по цифрам справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		// Проверяем, что символ является цифрой
		if digit < 0 || digit > 9 {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
