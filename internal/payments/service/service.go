package service

import (
	"context"
	"fmt"

	"encore.app/config"
	"encore.app/connection"
	schema "encore.app/gen/sql"
	"encore.app/internal/auth/repo"
	merchantRepo "encore.app/internal/merchants/repo"
	"encore.app/pkg/business"
	"encore.app/pkg/tigerbeetle"
	"encore.app/pkg/transaction"
	"encore.app/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type PaymentParams struct {
	UserID         string
	MerchantID     string
	OriginalAmount float64
	MerchantType   string // "restaurant" or "grocery"
}

type PaymentResponse struct {
	Success         bool
	CoinsSpent      float64
	DiscountAmount  float64
	DiscountPercent float64
	NewBalance      float64
	Message         string
}

type PaymentService interface {
	ProcessPayment(ctx context.Context, params PaymentParams) (*PaymentResponse, error)
	CalculateDiscount(ctx context.Context, userID string, amount float64, merchantType string) (*business.PaymentCalculation, error)
}

type paymentService struct {
	authRepo     repo.AuthRepository
	merchantRepo merchantRepo.MerchantRepository
	tb           tigerbeetle.Service
	calculator   *business.DiscountCalculator
	txWrapper    *transaction.TransactionWrapper
}

func NewPaymentService() (PaymentService, error) {
	authRepo, err := repo.NewAuthRepository()
	if err != nil {
		return nil, err
	}

	merchantRepo, err := merchantRepo.NewMerchantRepository()
	if err != nil {
		return nil, err
	}

	tbService, err := tigerbeetle.NewService()
	if err != nil {
		return nil, err
	}

	calculator := business.NewDiscountCalculator()
	
	// Initialize transaction wrapper
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	
	txWrapper := transaction.NewTransactionWrapper(db, tbService)
	
	return &paymentService{
		authRepo:     authRepo,
		merchantRepo: merchantRepo,
		tb:           tbService,
		calculator:   calculator,
		txWrapper:    txWrapper,
	}, nil
}

func (s *paymentService) CalculateDiscount(ctx context.Context, userID string, amount float64, merchantType string) (*business.PaymentCalculation, error) {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	// Get user balance
	balance, err := s.tb.GetBalance(uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	// Calculate payment
	calculation := s.calculator.CalculatePayment(amount, merchantType, balance)
	
	return &calculation, nil
}

func (s *paymentService) ProcessPayment(ctx context.Context, params PaymentParams) (*PaymentResponse, error) {
	// Parse UUIDs
	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		return &PaymentResponse{
			Success: false,
			Message: "Invalid user ID",
		}, nil
	}

	merchantID, err := uuid.Parse(params.MerchantID)
	if err != nil {
		return &PaymentResponse{
			Success: false,
			Message: "Invalid merchant ID",
		}, nil
	}

	// Verify merchant exists and get type
	merchant, err := s.merchantRepo.GetMerchantByID(ctx, merchantID)
	if err != nil {
		return &PaymentResponse{
			Success: false,
			Message: "Merchant not found",
		}, nil
	}

	// Use merchant's category if not provided
	merchantType := params.MerchantType
	if merchantType == "" {
		merchantType = merchant.Category.String
	}

	// Validate business hours
	err = s.calculator.ValidateBusinessHours(merchantType)
	if err != nil {
		return &PaymentResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Get user balance
	userBalance, err := s.tb.GetBalance(userID)
	if err != nil {
		return &PaymentResponse{
			Success: false,
			Message: "Failed to get user balance",
		}, nil
	}

	// Calculate discount and validate
	calculation := s.calculator.CalculatePayment(params.OriginalAmount, merchantType, userBalance)
	if !calculation.Valid {
		return &PaymentResponse{
			Success: false,
			Message: calculation.ErrorMessage,
		}, nil
	}

	// Check daily/monthly limits
	dailySpent, err := s.getDailySpending(ctx, userID)
	if err != nil {
		dailySpent = 0
	}
	
	monthlySpent, err := s.getMonthlySpending(ctx, userID)
	if err != nil {
		monthlySpent = 0
	}
	
	err = s.calculator.ValidateSpendingLimits(params.UserID, calculation.CoinsRequired, dailySpent, monthlySpent)
	if err != nil {
		return &PaymentResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	
	// Process payment with transaction wrapper
	if s.txWrapper != nil {
		err = s.txWrapper.ProcessPaymentWithRecord(ctx, userID, merchantID, calculation.CoinsRequired, merchantType)
	} else {
		// Fallback to direct TigerBeetle payment
		err = s.tb.ProcessPayment(userID, merchantID, calculation.CoinsRequired)
	}
	
	if err != nil {
		return &PaymentResponse{
			Success: false,
			Message: fmt.Sprintf("Payment failed: %v", err),
		}, nil
	}

	// Get new balance
	newBalance, _ := s.tb.GetBalance(userID)

	return &PaymentResponse{
		Success:         true,
		CoinsSpent:      calculation.CoinsRequired,
		DiscountAmount:  calculation.DiscountAmount,
		DiscountPercent: calculation.DiscountPercent,
		NewBalance:      newBalance,
		Message:         fmt.Sprintf("Payment successful! You saved $%.2f (%.1f%% discount)", calculation.DiscountAmount, calculation.DiscountPercent),
	}, nil
}
func (s *paymentService) getDailySpending(ctx context.Context, userID uuid.UUID) (float64, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}
	
	queries := schema.New(db)
	pgUUID := pgtype.UUID{}
	pgUUID.Scan(userID)
	
	result, err := queries.GetUserDailySpending(ctx, pgUUID)
	if err != nil {
		return 0, err
	}
	
	return utils.NumericToFloat64(result.(pgtype.Numeric)), nil
}

func (s *paymentService) getMonthlySpending(ctx context.Context, userID uuid.UUID) (float64, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}
	
	queries := schema.New(db)
	pgUUID := pgtype.UUID{}
	pgUUID.Scan(userID)
	
	result, err := queries.GetUserMonthlySpending(ctx, pgUUID)
	if err != nil {
		return 0, err
	}
	
	return utils.NumericToFloat64(result.(pgtype.Numeric)), nil
}
