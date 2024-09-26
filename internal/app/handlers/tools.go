package handlers

import (
	"strconv"
)

func LuhnCheck(number string) (bool, error) {
	sum := 0
	isSecondDigit := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false, err
		}

		if isSecondDigit {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecondDigit = !isSecondDigit
	}

	return sum%10 == 0, nil
}
