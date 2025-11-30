package repo

import (
	"context"

	"rival/config"
	"rival/connection"
	schema "rival/gen/sql"
	"rival/pkg/tb"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository interface {
	// Coin Purchases
	CreateCoinPurchase(ctx context.Context, params schema.CreateCoinPurchaseParams) (schema.CoinPurchase, error)
	GetCoinPurchaseByID(ctx context.Context, id int) (schema.CoinPurchase, error)
	UpdateCoinPurchaseStatus(ctx context.Context, params schema.UpdateCoinPurchaseStatusParams) error
	GetUserCoinPurchases(ctx context.Context, userID int, limit, offset int32) ([]schema.CoinPurchase, error)

	// Transactions
	CreateTransaction(ctx context.Context, params schema.CreateTransactionParams) (schema.Transaction, error)
	GetTransactionByID(ctx context.Context, id int) (schema.Transaction, error)
	UpdateTransactionStatus(ctx context.Context, params schema.UpdateTransactionStatusParams) error
	GetUserTransactions(ctx context.Context, userID int, limit, offset int32) ([]schema.Transaction, error)

	// Settlements
	CreateSettlement(ctx context.Context, params schema.CreateSettlementParams) (schema.Settlement, error)
	GetSettlementByID(ctx context.Context, id int) (schema.Settlement, error)
	UpdateSettlementStatus(ctx context.Context, params schema.UpdateSettlementStatusParams) error
	GetMerchantSettlements(ctx context.Context, merchantID int, limit, offset int32) ([]schema.Settlement, error)

	// Merchants
	GetMerchantByID(ctx context.Context, merchantID int) (schema.Merchant, error)
	GetUserByID(ctx context.Context, userID int64) (schema.User, error)

	// TigerBeetle Operations
	GetBalance(ctx context.Context, accountID int) (float64, error)
	AddCoins(ctx context.Context, userID int, amount float64) error
	ProcessPayment(ctx context.Context, userID, merchantID int, amount float64) error
	ProcessRefund(ctx context.Context, fromID, toID int, amount float64) error
	GetAccountTransfers(ctx context.Context, accountID int) ([]map[string]interface{}, error)
}

type paymentRepository struct {
	db      *pgxpool.Pool
	queries *schema.Queries
	tb      *tb.TbService
}

func NewPaymentRepository() (PaymentRepository, error) {
	cfg := config.GetConfig()

	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	tbService, err := tb.NewService()
	if err != nil {
		return nil, err
	}

	return &paymentRepository{
		db:      db,
		queries: schema.New(db),
		tb:      tbService,
	}, nil
}

func (r *paymentRepository) CreateCoinPurchase(ctx context.Context, params schema.CreateCoinPurchaseParams) (schema.CoinPurchase, error) {
	return r.queries.CreateCoinPurchase(ctx, params)
}

func (r *paymentRepository) GetCoinPurchaseByID(ctx context.Context, id int) (schema.CoinPurchase, error) {
	return r.queries.GetCoinPurchaseByID(ctx, int64(id))
}

func (r *paymentRepository) UpdateCoinPurchaseStatus(ctx context.Context, params schema.UpdateCoinPurchaseStatusParams) error {
	return r.queries.UpdateCoinPurchaseStatus(ctx, params)
}

func (r *paymentRepository) GetUserCoinPurchases(ctx context.Context, userID int, limit, offset int32) ([]schema.CoinPurchase, error) {

	return r.queries.GetUserCoinPurchases(ctx, schema.GetUserCoinPurchasesParams{
		UserID: pgtype.Int8{Int64: int64(userID), Valid: true},
		Limit:  limit,
		Offset: offset,
	})
}

func (r *paymentRepository) CreateTransaction(ctx context.Context, params schema.CreateTransactionParams) (schema.Transaction, error) {
	return r.queries.CreateTransaction(ctx, params)
}

func (r *paymentRepository) GetTransactionByID(ctx context.Context, id int) (schema.Transaction, error) {
	return r.queries.GetTransactionByID(ctx, int64(id))
}

func (r *paymentRepository) UpdateTransactionStatus(ctx context.Context, params schema.UpdateTransactionStatusParams) error {
	return r.queries.UpdateTransactionStatus(ctx, params)
}

func (r *paymentRepository) GetUserTransactions(ctx context.Context, userID int, limit, offset int32) ([]schema.Transaction, error) {

	return r.queries.GetUserTransactions(ctx, schema.GetUserTransactionsParams{
		UserID: pgtype.Int8{Int64: int64(userID), Valid: true},
		Limit:  limit,
		Offset: offset,
	})
}

func (r *paymentRepository) CreateSettlement(ctx context.Context, params schema.CreateSettlementParams) (schema.Settlement, error) {
	return r.queries.CreateSettlement(ctx, params)
}

func (r *paymentRepository) GetSettlementByID(ctx context.Context, id int) (schema.Settlement, error) {
	return r.queries.GetSettlementByID(ctx, int64(id))
}

func (r *paymentRepository) UpdateSettlementStatus(ctx context.Context, params schema.UpdateSettlementStatusParams) error {
	return r.queries.UpdateSettlementStatus(ctx, params)
}

func (r *paymentRepository) GetMerchantSettlements(ctx context.Context, merchantID int, limit, offset int32) ([]schema.Settlement, error) {

	return r.queries.GetMerchantSettlements(ctx, schema.GetMerchantSettlementsParams{
		MerchantID: pgtype.Int8{Int64: int64(merchantID), Valid: true},
		Limit:      limit,
		Offset:     offset,
	})
}

// Merchants
func (r *paymentRepository) GetMerchantByID(ctx context.Context, merchantID int) (schema.Merchant, error) {
	return r.queries.GetMerchantByID(ctx, int64(merchantID))
}

// TigerBeetle Operations
func (r *paymentRepository) GetBalance(ctx context.Context, accountID int) (float64, error) {
	return r.tb.GetBalance(accountID)
}

func (r *paymentRepository) AddCoins(ctx context.Context, userID int, amount float64) error {
	return r.tb.AddCoins(userID, amount)
}

func (r *paymentRepository) ProcessPayment(ctx context.Context, userID, merchantID int, amount float64) error {
	return r.tb.ProcessPayment(userID, merchantID, amount)
}

func (r *paymentRepository) ProcessRefund(ctx context.Context, fromID, toID int, amount float64) error {
	return r.tb.Transfer(fromID, toID, amount)
}

func (r *paymentRepository) GetAccountTransfers(ctx context.Context, accountID int) ([]map[string]interface{}, error) {
	transfers, err := r.tb.GetAccountTransfers(accountID)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, t := range transfers {
		creditID := t.CreditAccountID.BigInt()
		debitID := t.DebitAccountID.BigInt()
		amount := t.Amount.BigInt()

		isDebit := debitID.Uint64() == uint64(accountID)
		
		// Map code to transaction type
		var txType, desc string
		switch t.Code {
		case 1: // Coin purchase/add
			if isDebit {
				txType = "debit"
				desc = "Coin purchase fee"
			} else {
				txType = "credit"
				desc = "Coin purchase"
			}
		case 2: // Payment to merchant
			if isDebit {
				txType = "debit"
				desc = "Payment to merchant"
			} else {
				txType = "credit"
				desc = "Payment received"
			}
		case 3: // Transfer
			var otherUserID uint64
			if isDebit {
				txType = "debit"
				desc = "Transfer sent"
				otherUserID = creditID.Uint64() // Receiver
			} else {
				txType = "credit"
				desc = "Transfer received"
				otherUserID = debitID.Uint64() // Sender
			}
			
			result = append(result, map[string]interface{}{
				"id":            t.ID.String(),
				"type":          txType,
				"amount":        float64(amount.Uint64()) / 100,
				"credit_id":     creditID.Uint64(),
				"debit_id":      debitID.Uint64(),
				"code":          t.Code,
				"timestamp":     t.Timestamp,
				"description":   desc,
				"other_user_id": otherUserID,
			})
			continue
		default:
			if isDebit {
				txType = "debit"
				desc = "Debit transaction"
			} else {
				txType = "credit"
				desc = "Credit transaction"
			}
		}

		result = append(result, map[string]interface{}{
			"id":          t.ID.String(),
			"type":        txType,
			"amount":      float64(amount.Uint64()) / 100,
			"credit_id":   creditID.Uint64(),
			"debit_id":    debitID.Uint64(),
			"code":        t.Code,
			"timestamp":   t.Timestamp,
			"description": desc,
		})
	}
	return result, nil
}

func (r *paymentRepository) GetUserByID(ctx context.Context, userID int64) (schema.User, error) {
	return r.queries.GetUserByID(ctx, userID)
}
