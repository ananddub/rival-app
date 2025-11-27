package service

import (
	"context"
	"fmt"
	"time"

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
	GetFinancialHistory(ctx context.Context, req *paymentpb.GetFinancialHistoryRequest) (*paymentpb.GetFinancialHistoryResponse, error)

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
	coinsToReceive := req.Amount // 1:1 ratio

	createParams := schema.CreateCoinPurchaseParams{
		UserID:        pgtype.Int8{Int64: req.UserId, Valid: true},
		Amount:        utils.Float64ToNumeric(req.Amount),
		CoinsReceived: utils.Float64ToNumeric(coinsToReceive),
		PaymentMethod: pgtype.Text{String: req.PaymentMethod, Valid: true},
		Status:        pgtype.Text{String: "completed", Valid: true},
	}

	purchase, err := s.repo.CreateCoinPurchase(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create coin purchase: %w", err)
	}

	// Add coins to TigerBeetle
	err = s.repo.AddCoins(ctx, int(req.UserId), coinsToReceive)
	if err != nil {
		return nil, fmt.Errorf("failed to add coins: %w", err)
	}

	// Get new balance
	newBalance, err := s.repo.GetBalance(ctx, int(req.UserId))
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &paymentpb.InitiateCoinPurchaseResponse{
		PaymentId:      fmt.Sprintf("%d", purchase.ID),
		PaymentUrl:     "",
		CoinsToReceive: coinsToReceive,
		Status:         "completed",
		NewBalance:     newBalance,
	}, nil
}

func (s *paymentService) VerifyPayment(ctx context.Context, req *paymentpb.VerifyPaymentRequest) (*paymentpb.VerifyPaymentResponse, error) {
	// Convert paymentID string to int
	var paymentID int
	fmt.Sscanf(req.PaymentId, "%d", &paymentID)

	purchase, err := s.repo.GetCoinPurchaseByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase: %w", err)
	}

	var userID int
	if purchase.UserID.Valid {
		userID = int(purchase.UserID.Int64)
	}

	coinsToAdd := utils.NumericToFloat64(purchase.CoinsReceived)

	// Add coins to TigerBeetle
	err = s.repo.AddCoins(ctx, userID, coinsToAdd)
	if err != nil {
		return nil, fmt.Errorf("failed to add coins: %w", err)
	}

	// Update purchase status
	err = s.repo.UpdateCoinPurchaseStatus(ctx, schema.UpdateCoinPurchaseStatusParams{
		ID:     purchase.ID,
		Status: pgtype.Text{String: "completed", Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update purchase status: %w", err)
	}

	// Get new balance
	newBalance, err := s.repo.GetBalance(ctx, userID)
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
	userID := int(req.UserId)

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
	// Convert paymentID string to int
	var paymentID int
	fmt.Sscanf(req.PaymentId, "%d", &paymentID)

	// Get the original purchase
	purchase, err := s.repo.GetCoinPurchaseByID(ctx, paymentID)
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

	var userID int
	if purchase.UserID.Valid {
		userID = int(purchase.UserID.Int64)
	}

	refundAmount := utils.NumericToFloat64(purchase.CoinsReceived)

	// Remove coins from user account (reverse the add operation)
	err = s.repo.ProcessRefund(ctx, userID, 0, refundAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// Update purchase status to refunded
	err = s.repo.UpdateCoinPurchaseStatus(ctx, schema.UpdateCoinPurchaseStatusParams{
		ID:     purchase.ID,
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
	userID := int(req.UserId)
	merchantID := int(req.MerchantId)

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
	createParams := schema.CreateTransactionParams{
		UserID:          pgtype.Int8{Int64: req.UserId, Valid: true},
		MerchantID:      pgtype.Int8{Int64: req.MerchantId, Valid: true},
		CoinsSpent:      utils.Float64ToNumeric(finalAmount),
		OriginalAmount:  utils.Float64ToNumeric(req.Amount),
		DiscountAmount:  utils.Float64ToNumeric(discountAmount),
		FinalAmount:     utils.Float64ToNumeric(finalAmount),
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

	return &paymentpb.PayToMerchantResponse{
		Success:          true,
		TransactionId:    fmt.Sprintf("%d", transaction.ID),
		DiscountAmount:   discountAmount,
		FinalAmount:      finalAmount,
		RemainingBalance: remainingBalance,
		Transaction:      convertToProtoTransaction(transaction),
	}, nil
}

func (s *paymentService) TransferToUser(ctx context.Context, req *paymentpb.TransferToUserRequest) (*paymentpb.TransferToUserResponse, error) {
	fromUserID := int(req.FromUserId)
	toUserID := int(req.ToUserId)

	// Process transfer in TigerBeetle
	err := s.repo.ProcessRefund(ctx, fromUserID, toUserID, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to process transfer: %w", err)
	}

	// Create transaction record for sender (debit)
	senderTx, err := s.repo.CreateTransaction(ctx, schema.CreateTransactionParams{
		UserID:          pgtype.Int8{Int64: req.FromUserId, Valid: true},
		CoinsSpent:      utils.Float64ToNumeric(req.Amount),
		OriginalAmount:  utils.Float64ToNumeric(req.Amount),
		DiscountAmount:  utils.Float64ToNumeric(0),
		FinalAmount:     utils.Float64ToNumeric(req.Amount),
		TransactionType: pgtype.Text{String: "transfer_out", Valid: true},
		Status:          pgtype.Text{String: "completed", Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create sender transaction: %w", err)
	}

	// Create transaction record for receiver (credit)
	_, err = s.repo.CreateTransaction(ctx, schema.CreateTransactionParams{
		UserID:          pgtype.Int8{Int64: req.ToUserId, Valid: true},
		CoinsSpent:      utils.Float64ToNumeric(-req.Amount), // Negative for credit
		OriginalAmount:  utils.Float64ToNumeric(req.Amount),
		DiscountAmount:  utils.Float64ToNumeric(0),
		FinalAmount:     utils.Float64ToNumeric(req.Amount),
		TransactionType: pgtype.Text{String: "transfer_in", Valid: true},
		Status:          pgtype.Text{String: "completed", Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create receiver transaction: %w", err)
	}

	// Get remaining balance
	remainingBalance, err := s.repo.GetBalance(ctx, fromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &paymentpb.TransferToUserResponse{
		Success:          true,
		TransactionId:    fmt.Sprintf("%d", senderTx.ID),
		RemainingBalance: remainingBalance,
		Transaction:      convertToProtoTransaction(senderTx),
	}, nil
}

func (s *paymentService) GetBalance(ctx context.Context, req *paymentpb.GetBalanceRequest) (*paymentpb.GetBalanceResponse, error) {
	userID := int(req.UserId)

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
	userID := int(req.UserId)

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
	// Convert transactionID string to int
	var transactionID int
	fmt.Sscanf(req.TransactionId, "%d", &transactionID)

	// Get original transaction
	transaction, err := s.repo.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	var userID int
	if transaction.UserID.Valid {
		userID = int(transaction.UserID.Int64)
	}

	// Process refund in TigerBeetle (add coins back to user)
	err = s.repo.AddCoins(ctx, userID, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// Create refund transaction
	createParams := schema.CreateTransactionParams{
		UserID:          pgtype.Int8{Int64: int64(userID), Valid: true},
		CoinsSpent:      utils.Float64ToNumeric(-req.Amount), // Negative for refund
		OriginalAmount:  utils.Float64ToNumeric(req.Amount),
		DiscountAmount:  utils.Float64ToNumeric(0),
		FinalAmount:     utils.Float64ToNumeric(req.Amount),
		TransactionType: pgtype.Text{String: "refund", Valid: true},
		Status:          pgtype.Text{String: "completed", Valid: true},
	}

	refundTransaction, err := s.repo.CreateTransaction(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund transaction: %w", err)
	}

	return &paymentpb.ProcessRefundResponse{
		Success:             true,
		RefundTransactionId: fmt.Sprintf("%d", refundTransaction.ID),
		RefundedAmount:      req.Amount,
		RefundTransaction:   convertToProtoTransaction(refundTransaction),
	}, nil
}

// Merchant Settlements
func (s *paymentService) InitiateSettlement(ctx context.Context, req *paymentpb.InitiateSettlementRequest) (*paymentpb.InitiateSettlementResponse, error) {
	now := time.Now()
	periodStart := now.AddDate(0, 0, -30) // Last 30 days

	createParams := schema.CreateSettlementParams{
		MerchantID:          pgtype.Int8{Int64: req.MerchantId, Valid: true},
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

	return &paymentpb.InitiateSettlementResponse{
		Success:          true,
		SettlementId:     fmt.Sprintf("%d", settlement.ID),
		SettlementAmount: req.Amount,
		Settlement:       convertToProtoSettlement(settlement),
	}, nil
}

func (s *paymentService) GetSettlements(ctx context.Context, req *paymentpb.GetSettlementsRequest) (*paymentpb.GetSettlementsResponse, error) {
	merchantID := int(req.MerchantId)

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

func (s *paymentService) GetFinancialHistory(ctx context.Context, req *paymentpb.GetFinancialHistoryRequest) (*paymentpb.GetFinancialHistoryResponse, error) {
	userID := int(req.UserId)
	var items []*paymentpb.FinancialHistoryItem

	// Get all transfers from TigerBeetle
	tbTransfers, err := s.repo.GetAccountTransfers(ctx, userID)
	if err != nil {
		fmt.Printf("Error getting TB transfers: %v\n", err)
	} else {
		fmt.Printf("Got %d TB transfers\n", len(tbTransfers))
		for _, t := range tbTransfers {
			txType := t["type"].(string)
			if req.Type != "" && req.Type != "all" && txType != req.Type {
				continue
			}

			amount := t["amount"].(float64)
			desc := t["description"].(string)
			
			items = append(items, &paymentpb.FinancialHistoryItem{
				Id:          t["id"].(string),
				Type:        txType,
				Amount:      amount,
				Coins:       amount,
				Description: desc,
				Status:      "completed",
				CreatedAt:   int64(t["timestamp"].(uint64)),
			})
		}
	}

	balance, _ := s.repo.GetBalance(ctx, userID)

	return &paymentpb.GetFinancialHistoryResponse{
		Items:          items,
		TotalCount:     int32(len(items)),
		CurrentBalance: balance,
	}, nil
}

// Conversion functions
func convertToProtoCoinPurchase(purchase schema.CoinPurchase) *schemapb.CoinPurchase {
	var userID int64
	if purchase.UserID.Valid {
		userID = purchase.UserID.Int64
	}

	return &schemapb.CoinPurchase{
		Id:            purchase.ID,
		UserId:        userID,
		Amount:        utils.NumericToFloat64(purchase.Amount),
		CoinsReceived: utils.NumericToFloat64(purchase.CoinsReceived),
		PaymentMethod: purchase.PaymentMethod.String,
		Status:        purchase.Status.String,
		CreatedAt:     purchase.CreatedAt.Time.Unix(),
	}
}

func convertToProtoTransaction(tx schema.Transaction) *schemapb.Transaction {
	var userID int64
	if tx.UserID.Valid {
		userID = tx.UserID.Int64
	}

	var merchantID int64
	if tx.MerchantID.Valid {
		merchantID = tx.MerchantID.Int64
	}

	return &schemapb.Transaction{
		Id:             tx.ID,
		UserId:         userID,
		MerchantId:     merchantID,
		CoinsSpent:     utils.NumericToFloat64(tx.CoinsSpent),
		OriginalAmount: utils.NumericToFloat64(tx.OriginalAmount),
		DiscountAmount: utils.NumericToFloat64(tx.DiscountAmount),
		Status:         tx.Status.String,
		CreatedAt:      tx.CreatedAt.Time.Unix(),
	}
}

func convertToProtoSettlement(settlement schema.Settlement) *schemapb.Settlement {
	var merchantID int64
	if settlement.MerchantID.Valid {
		merchantID = settlement.MerchantID.Int64
	}

	return &schemapb.Settlement{
		Id:                  settlement.ID,
		MerchantId:          merchantID,
		PeriodStart:         settlement.PeriodStart.Time.Format("2006-01-02"),
		PeriodEnd:           settlement.PeriodEnd.Time.Format("2006-01-02"),
		TotalTransactions:   settlement.TotalTransactions.Int32,
		TotalDiscountAmount: utils.NumericToFloat64(settlement.TotalDiscountAmount),
		SettlementAmount:    utils.NumericToFloat64(settlement.SettlementAmount),
		Status:              settlement.Status.String,
		CreatedAt:           settlement.CreatedAt.Time.Unix(),
	}
}
