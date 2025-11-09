package utils

import (
	"errors"
	"regexp"
)

func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if amount > 1000000 {
		return errors.New("amount too large")
	}
	return nil
}

func ValidateUserID(userID int64) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	return nil
}

func ValidateTransactionTitle(title string) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}
	if len(title) > 100 {
		return errors.New("title too long")
	}
	
	// Check for valid characters
	validTitle := regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,!?]+$`)
	if !validTitle.MatchString(title) {
		return errors.New("title contains invalid characters")
	}
	
	return nil
}

func ValidateDescription(description string) error {
	if len(description) > 500 {
		return errors.New("description too long")
	}
	return nil
}

func ValidatePage(page int) error {
	if page < 1 {
		return errors.New("page must be positive")
	}
	if page > 1000 {
		return errors.New("page number too large")
	}
	return nil
}

func FormatAmount(amount float64) float64 {
	// Round to 2 decimal places
	return float64(int(amount*100)) / 100
}

func ValidateCurrency(currency string) error {
	validCurrencies := map[string]bool{
		"INR": true,
		"USD": true,
		"EUR": true,
	}
	
	if !validCurrencies[currency] {
		return errors.New("unsupported currency")
	}
	
	return nil
}
