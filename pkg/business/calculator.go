package business

import (
	"fmt"
	"time"
)

type DiscountCalculator struct {
	RestaurantDiscount float64 // 15%
	GroceryDiscount    float64 // 2%
	MinBalance         float64 // Minimum balance required
	DailyLimit         float64 // Daily spending limit
	MonthlyLimit       float64 // Monthly spending limit
}

func NewDiscountCalculator() *DiscountCalculator {
	return &DiscountCalculator{
		RestaurantDiscount: 0.15, // 15%
		GroceryDiscount:    0.02, // 2%
		MinBalance:         1.0,  // $1 minimum
		DailyLimit:         100.0, // $100 daily
		MonthlyLimit:       1000.0, // $1000 monthly
	}
}

type PaymentCalculation struct {
	OriginalAmount    float64
	DiscountAmount    float64
	CoinsRequired     float64
	DiscountPercent   float64
	Valid             bool
	ErrorMessage      string
}

func (dc *DiscountCalculator) CalculatePayment(originalAmount float64, merchantType string, userBalance float64) PaymentCalculation {
	var discountPercent float64
	
	// Determine discount based on merchant type
	switch merchantType {
	case "restaurant":
		discountPercent = dc.RestaurantDiscount
	case "grocery":
		discountPercent = dc.GroceryDiscount
	default:
		return PaymentCalculation{
			Valid:        false,
			ErrorMessage: "invalid merchant type",
		}
	}
	
	// Calculate discount amount
	discountAmount := originalAmount * discountPercent
	coinsRequired := discountAmount
	
	// Validate minimum amount
	if originalAmount < 1.0 {
		return PaymentCalculation{
			Valid:        false,
			ErrorMessage: "minimum purchase amount is $1.00",
		}
	}
	
	// Validate user balance
	if userBalance < coinsRequired {
		return PaymentCalculation{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("insufficient balance: need %.2f coins, have %.2f", coinsRequired, userBalance),
		}
	}
	
	// Check minimum balance after payment
	if userBalance-coinsRequired < dc.MinBalance {
		return PaymentCalculation{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("payment would leave balance below minimum of $%.2f", dc.MinBalance),
		}
	}
	
	return PaymentCalculation{
		OriginalAmount:  originalAmount,
		DiscountAmount:  discountAmount,
		CoinsRequired:   coinsRequired,
		DiscountPercent: discountPercent * 100, // Convert to percentage
		Valid:           true,
	}
}

func (dc *DiscountCalculator) ValidateSpendingLimits(userID string, amount float64, dailySpent, monthlySpent float64) error {
	// Check daily limit
	if dailySpent+amount > dc.DailyLimit {
		return fmt.Errorf("daily spending limit exceeded: limit $%.2f, would spend $%.2f", dc.DailyLimit, dailySpent+amount)
	}
	
	// Check monthly limit
	if monthlySpent+amount > dc.MonthlyLimit {
		return fmt.Errorf("monthly spending limit exceeded: limit $%.2f, would spend $%.2f", dc.MonthlyLimit, monthlySpent+amount)
	}
	
	return nil
}

type ReferralBonus struct {
	ReferrerBonus float64 // Bonus for referrer
	RefereeBonus  float64 // Bonus for new user
}

func (dc *DiscountCalculator) CalculateReferralBonus() ReferralBonus {
	return ReferralBonus{
		ReferrerBonus: 5.0,  // $5 for referrer
		RefereeBonus:  10.0, // $10 for new user
	}
}

func (dc *DiscountCalculator) ValidateBusinessHours(merchantType string) error {
	now := time.Now()
	hour := now.Hour()
	
	switch merchantType {
	case "restaurant":
		// Restaurant hours: 6 AM to 11 PM
		if hour < 6 || hour > 23 {
			return fmt.Errorf("restaurant payments only allowed between 6 AM and 11 PM")
		}
	case "grocery":
		// Grocery available 24/7
		return nil
	}
	
	return nil
}
