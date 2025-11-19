package service

import (
	"context"
	"fmt"
	"time"

	"rival/config"
	paymentpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	"rival/internal/payments/repo"
	"rival/pkg/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type PaymentService interface {
	// Coin Purchase
	InitiateCoinPurchase(ctx context.Context, req *paymentpb.InitiateCoinPurchaseRequest) (*paymentpb.InitiateCoinPurchaseResponse, error)
	VerifyPayment(ctx context.Context, req *paymentpb.VerifyPaymentRequest) (*paymentpb.VerifyPaymentResponse, error)
	GetPaymentHistory(ctx context.Context, req *paymentpb.GetPaymentHistoryRequest) (*paymentpb.GetPaymentHistoryResponse, error)
	RefundPayment(ctx context.Context, req *paymentpb.RefundPaymentRequest) (*paymentpb.RefundPaymentResponse, error)

	// Payment Transfers
	PayToMerchant(ctx context.Context, req *paymentpb.PayToMerchantRequest) (*paymentpb.PayToMerchantResponse, error)
	TransferToUser(ctx context.Context, req *paymentpb.TransferToUserRequest) (*paymentpb.TransferToUserResponse, error)
	GetBalance(ctx context.Context, req *paymentpb.GetBalanceRequest) (*paymentpb.GetBalanceResponse, error)
	GetTransactionHistory(ctx context.Context, req *paymentpb.GetTransactionHistoryRequest) (*paymentpb.GetTransactionHistoryResponse, error)
	ProcessRefund(ctx context.Context, req *paymentpb.ProcessRefundRequest) (*paymentpb.ProcessRefundResponse, error)

	// Merchant Settlements
	InitiateSettlement(ctx context.Context, req *paymentpb.InitiateSettlementRequest) (*paymentpb.InitiateSettlementResponse, error)
	GetSettlements(ctx context.Context, req *paymentpb.GetSettlementsRequest) (*paymentpb.GetSettlementsResponse, error)
}

type paymentService struct {
	repo repo.PaymentRepository
}

func NewPaymentService(repo repo.PaymentRepository) PaymentService {
	return &paymentService{
		repo: repo,
	}
}

// Coin Purchase
func (s *paymentService) InitiateCoinPurchase(ctx context.Context, req *paymentpb.InitiateCoinPurchaseRequest) (*paymentpb.InitiateCoinPurchaseResponse, error) {
	userID := req.UserId

	paymentID := uuid.New().String()
	coinsToReceive := req.Amount // 1:1 ratio

	pgUserID := pgtype.UUID{}
	pgUserID.Scan(userID)

	createParams := schema.CreateCoinPurchaseParams{
		UserID:          pgtype.Int8{Int64: int64(userID), Valid: true},
		Amount:          utils.Float64ToNumeric(req.Amount),
		CoinsReceived:   utils.Float64ToNumeric(coinsToReceive),
		PaymentMethod:   pgtype.Text{String: req.PaymentMethod, Valid: true},
		Status:          pgtype.Text{String: "pending", Valid: true},
	}

	purchase, err := s.repo.CreateCoinPurchase(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create coin purchase: %w", err)
	}

	cfg := config.GetConfig()
	paymentURL := fmt.Sprintf("%s/pay/%s", cfg.PaymentGateway.BaseURL, paymentID)

	return &paymentpb.InitiateCoinPurchaseResponse{
		PaymentId:      fmt.Sprintf("%d", purchase.ID),
		PaymentUrl:     paymentURL,
		CoinsToReceive: coinsToReceive,
		Status:         "pending",
	}, nil
}

func (s *paymentService) VerifyPayment(ctx context.Context, req *paymentpb.VerifyPaymentRequest) (*paymentpb.VerifyPaymentResponse, error) {
	

	purchase, err := s.repo.GetCoinPurchaseByID(ctx, req.PaymentId)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase: %w", err)
	}

	userID, _ := purchase.UserID.Value()
	userUUID, _ := uuid.Parse(userID.(string))
	coinsToAdd := utils.NumericToFloat64(purchase.CoinsReceived)

	// Add coins to TigerBeetle
	err = s.repo.AddCoins(ctx, userUUID, coinsToAdd)
	if err != nil {
		return nil, fmt.Errorf("failed to add coins: %w", err)
	}

	// Update purchase status
	pgPurchaseID := pgtype.UUID{}
	pgPurchaseID.Scan(purchaseID)

	err = s.repo.UpdateCoinPurchaseStatus(ctx, schema.UpdateCoinPurchaseStatusParams{
		ID:     pgPurchaseID,
		Status: pgtype.Text{String: "completed", Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update purchase status: %w", err)
	}

	// Get new balance
	newBalance, err := s.repo.GetBalance(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &paymentpb.VerifyPaymentResponse{
		Success:    true,
		CoinsAdded: coinsToAdd,
		NewBalance: newBalance,
		Purchase:   convertToProtoCoinPurchase(purchase),
	}, nil
}

func (s *paymentService) GetPaymentHistory(ctx context.Context, req *paymentpb.GetPaymentHistoryRequest) (*paymentpb.GetPaymentHistoryResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	purchases, err := s.repo.GetUserCoinPurchases(ctx, userID, req.Limit, (req.Page-1)*req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchases: %w", err)
	}

	var protoPurchases []*schemapb.CoinPurchase
	for _, purchase := range purchases {
		protoPurchases = append(protoPurchases, convertToProtoCoinPurchase(purchase))
	}

	return &paymentpb.GetPaymentHistoryResponse{
		Purchases:  protoPurchases,
		TotalCount: int32(len(protoPurchases)),
	}, nil
}

func (s *paymentService) RefundPayment(ctx context.Context, req *paymentpb.RefundPaymentRequest) (*paymentpb.RefundPaymentResponse, error) {
	purchaseID, err := uuid.Parse(req.PaymentId)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}

	// Get the original purchase
	purchase, err := s.repo.GetCoinPurchaseByID(ctx, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase: %w", err)
	}

	// Check if already refunded
	if purchase.Status.String == "refunded" {
		return &paymentpb.RefundPaymentResponse{
			Success:        false,
			RefundId:       "",
			RefundedAmount: 0,
		}, nil
	}

	userID, _ := purchase.UserID.Value()
	userUUID, _ := uuid.Parse(userID.(string))
	refundAmount := utils.NumericToFloat64(purchase.CoinsReceived)

	// Remove coins from user account (reverse the add operation)
	err = s.repo.ProcessRefund(ctx, userUUID, uuid.Nil, refundAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// Update purchase status to refunded
	pgPurchaseID := pgtype.UUID{}
	pgPurchaseID.Scan(purchaseID)

	err = s.repo.UpdateCoinPurchaseStatus(ctx, schema.UpdateCoinPurchaseStatusParams{
		ID:     pgPurchaseID,
		Status: pgtype.Text{String: "refunded", Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update purchase status: %w", err)
	}

	refundID := uuid.New().String()

	return &paymentpb.RefundPaymentResponse{
		Success:        true,
		RefundId:       refundID,
		RefundedAmount: refundAmount,
	}, nil
}

// Payment Transfers
func (s *paymentService) PayToMerchant(ctx context.Context, req *paymentpb.PayToMerchantRequest) (*paymentpb.PayToMerchantResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant ID: %w", err)
	}

	// Get merchant's discount percentage from database
	merchant, err := s.repo.GetMerchantByID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	discountPercentage := utils.NumericToFloat64(merchant.DiscountPercentage)
	discountAmount := req.Amount * (discountPercentage / 100)
	finalAmount := req.Amount - discountAmount

	// Process payment in TigerBeetle
	err = s.repo.ProcessPayment(ctx, userID, merchantID, finalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to process payment: %w", err)
	}

	// Create transaction record
	pgUserID := pgtype.UUID{}
	pgUserID.Scan(userID)
	pgMerchantID := pgtype.UUID{}
	pgMerchantID.Scan(merchantID)

	createParams := schema.CreateTransactionParams{
		UserID:          pgUserID,
		MerchantID:      pgMerchantID,
		CoinsSpent:      utils.Float64ToNumeric(finalAmount),
		OriginalAmount:  utils.Float64ToNumeric(req.Amount),
		TransactionType: pgtype.Text{String: "payment", Valid: true},
		Status:          pgtype.Text{String: "completed", Valid: true},
	}

	transaction, err := s.repo.CreateTransaction(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Get remaining balance
	remainingBalance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	transactionID, _ := transaction.ID.Value()

	return &paymentpb.PayToMerchantResponse{
		Success:          true,
		TransactionId:    transactionID.(string),
		DiscountAmount:   discountAmount,
		FinalAmount:      finalAmount,
		RemainingBalance: remainingBalance,
		Transaction:      convertToProtoTransaction(transaction),
	}, nil
}

func (s *paymentService) TransferToUser(ctx context.Context, req *paymentpb.TransferToUserRequest) (*paymentpb.TransferToUserResponse, error) {
	fromUserID, err := uuid.Parse(req.FromUserId)
	if err != nil {
		return nil, fmt.Errorf("invalid from user ID: %w", err)
	}

	toUserID, err := uuid.Parse(req.ToUserId)
	if err != nil {
		return nil, fmt.Errorf("invalid to user ID: %w", err)
	}

	// Process transfer in TigerBeetle
	err = s.repo.ProcessRefund(ctx, fromUserID, toUserID, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to process transfer: %w", err)
	}

	// Create transaction record
	pgFromUserID := pgtype.UUID{}
	pgFromUserID.Scan(fromUserID)

	createParams := schema.CreateTransactionParams{
		UserID:          pgFromUserID,
		CoinsSpent:      utils.Float64ToNumeric(req.Amount),
		OriginalAmount:  utils.Float64ToNumeric(req.Amount),
		TransactionType: pgtype.Text{String: "transfer", Valid: true},
		Status:          pgtype.Text{String: "completed", Valid: true},
	}

	transaction, err := s.repo.CreateTransaction(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Get remaining balance
	remainingBalance, err := s.repo.GetBalance(ctx, fromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	transactionID, _ := transaction.ID.Value()

	return &paymentpb.TransferToUserResponse{
		Success:          true,
		TransactionId:    transactionID.(string),
		RemainingBalance: remainingBalance,
		Transaction:      convertToProtoTransaction(transaction),
	}, nil
}

func (s *paymentService) GetBalance(ctx context.Context, req *paymentpb.GetBalanceRequest) (*paymentpb.GetBalanceResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &paymentpb.GetBalanceResponse{
		Balance: balance,
		UserId:  req.UserId,
	}, nil
}

func (s *paymentService) GetTransactionHistory(ctx context.Context, req *paymentpb.GetTransactionHistoryRequest) (*paymentpb.GetTransactionHistoryResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	transactions, err := s.repo.GetUserTransactions(ctx, userID, req.Limit, (req.Page-1)*req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	var protoTransactions []*schemapb.Transaction
	for _, tx := range transactions {
		protoTransactions = append(protoTransactions, convertToProtoTransaction(tx))
	}

	return &paymentpb.GetTransactionHistoryResponse{
		Transactions: protoTransactions,
		TotalCount:   int32(len(protoTransactions)),
	}, nil
}

func (s *paymentService) ProcessRefund(ctx context.Context, req *paymentpb.ProcessRefundRequest) (*paymentpb.ProcessRefundResponse, error) {
	transactionID, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID: %w", err)
	}

	// Get original transaction
	transaction, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	userID, _ := transaction.UserID.Value()
	userUUID, _ := uuid.Parse(userID.(string))

	// Process refund in TigerBeetle (add coins back to user)
	err = s.repo.AddCoins(ctx, userUUID, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// Create refund transaction
	pgUserID := pgtype.UUID{}
	pgUserID.Scan(userUUID)

	createParams := schema.CreateTransactionParams{
		UserID:          pgUserID,
		CoinsSpent:      utils.Float64ToNumeric(-req.Amount), // Negative for refund
		OriginalAmount:  utils.Float64ToNumeric(req.Amount),
		TransactionType: pgtype.Text{String: "refund", Valid: true},
		Status:          pgtype.Text{String: "completed", Valid: true},
	}

	refundTransaction, err := s.repo.CreateTransaction(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund transaction: %w", err)
	}

	refundTransactionID, _ := refundTransaction.ID.Value()

	return &paymentpb.ProcessRefundResponse{
		Success:             true,
		RefundTransactionId: refundTransactionID.(string),
		RefundedAmount:      req.Amount,
		RefundTransaction:   convertToProtoTransaction(refundTransaction),
	}, nil
}

// Merchant Settlements
func (s *paymentService) InitiateSettlement(ctx context.Context, req *paymentpb.InitiateSettlementRequest) (*paymentpb.InitiateSettlementResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant ID: %w", err)
	}

	pgMerchantID := pgtype.UUID{}
	pgMerchantID.Scan(merchantID)

	now := time.Now()
	periodStart := now.AddDate(0, 0, -30) // Last 30 days

	createParams := schema.CreateSettlementParams{
		MerchantID:          pgMerchantID,
		PeriodStart:         pgtype.Date{Time: periodStart, Valid: true},
		PeriodEnd:           pgtype.Date{Time: now, Valid: true},
		TotalTransactions:   pgtype.Int4{Int32: 0, Valid: true},
		TotalDiscountAmount: utils.Float64ToNumeric(0),
		SettlementAmount:    utils.Float64ToNumeric(req.Amount),
		Status:              pgtype.Text{String: "pending", Valid: true},
	}

	settlement, err := s.repo.CreateSettlement(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create settlement: %w", err)
	}

	settlementID, _ := settlement.ID.Value()

	return &paymentpb.InitiateSettlementResponse{
		Success:          true,
		SettlementId:     settlementID.(string),
		SettlementAmount: req.Amount,
		Settlement:       convertToProtoSettlement(settlement),
	}, nil
}

func (s *paymentService) GetSettlements(ctx context.Context, req *paymentpb.GetSettlementsRequest) (*paymentpb.GetSettlementsResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant ID: %w", err)
	}

	settlements, err := s.repo.GetMerchantSettlements(ctx, merchantID, req.Limit, (req.Page-1)*req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get settlements: %w", err)
	}

	var protoSettlements []*schemapb.Settlement
	for _, settlement := range settlements {
		protoSettlements = append(protoSettlements, convertToProtoSettlement(settlement))
	}

	return &paymentpb.GetSettlementsResponse{
		Settlements: protoSettlements,
		TotalCount:  int32(len(protoSettlements)),
	}, nil
}

// Conversion functions
func convertToProtoCoinPurchase(purchase schema.CoinPurchase) *schemapb.CoinPurchase {
	purchaseID, _ := purchase.ID.Value()
	userID, _ := purchase.UserID.Value()

	return &schemapb.CoinPurchase{
		Id:            purchaseID.(string),
		UserId:        userID.(string),
		Amount:        utils.NumericToFloat64(purchase.Amount),
		CoinsReceived: utils.NumericToFloat64(purchase.CoinsReceived),
		PaymentMethod: purchase.PaymentMethod.String,
		Status:        purchase.Status.String,
		CreatedAt:     purchase.CreatedAt.Time.Unix(),
	}
}

func convertToProtoTransaction(tx schema.Transaction) *schemapb.Transaction {
	txID, _ := tx.ID.Value()
	userID, _ := tx.UserID.Value()

	var merchantID string
	if tx.MerchantID.Valid {
		merchantIDVal, _ := tx.MerchantID.Value()
		merchantID = merchantIDVal.(string)
	}

	return &schemapb.Transaction{
		Id:             txID.(string),
		UserId:         userID.(string),
		MerchantId:     merchantID,
		CoinsSpent:     utils.NumericToFloat64(tx.CoinsSpent),
		OriginalAmount: utils.NumericToFloat64(tx.OriginalAmount),
		DiscountAmount: utils.NumericToFloat64(tx.DiscountAmount),
		Status:         tx.Status.String,
		CreatedAt:      tx.CreatedAt.Time.Unix(),
	}
}

func convertToProtoSettlement(settlement schema.Settlement) *schemapb.Settlement {
	settlementID, _ := settlement.ID.Value()
	merchantID, _ := settlement.MerchantID.Value()

	return &schemapb.Settlement{
		Id:                  settlementID.(string),
		MerchantId:          merchantID.(string),
		PeriodStart:         settlement.PeriodStart.Time.Format("2006-01-02"),
		PeriodEnd:           settlement.PeriodEnd.Time.Format("2006-01-02"),
		TotalTransactions:   settlement.TotalTransactions.Int32,
		TotalDiscountAmount: utils.NumericToFloat64(settlement.TotalDiscountAmount),
		SettlementAmount:    utils.NumericToFloat64(settlement.SettlementAmount),
		Status:              settlement.Status.String,
		CreatedAt:           settlement.CreatedAt.Time.Unix(),
	}
}
