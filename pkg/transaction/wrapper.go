package transaction

import (
	"context"
	"fmt"

	schema "rival/gen/sql"
	"rival/pkg/tb"
	"rival/pkg/utils"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionWrapper struct {
	db      *pgxpool.Pool
	queries *schema.Queries
	tb      tb.Service
}

func NewTransactionWrapper(db *pgxpool.Pool, tb tb.Service) *TransactionWrapper {
	return &TransactionWrapper{
		db:      db,
		queries: schema.New(db),
		tb:      tb,
	}
}

func (tw *TransactionWrapper) ProcessPaymentWithRecord(ctx context.Context, userID, merchantID int, amount float64, paymentType string) error {
	// Start PostgreSQL transaction
	tx, err := tw.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// Check user balance first
	balance, err := tw.tb.GetBalance(userID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %v", err)
	}
	if balance < amount {
		return fmt.Errorf("insufficient balance: have %.2f, need %.2f", balance, amount)
	}

	// Process TigerBeetle payment
	err = tw.tb.ProcessPayment(userID, merchantID, amount)
	if err != nil {
		return fmt.Errorf("payment failed: %v", err)
	}

	// Record transaction in PostgreSQL using sqlc
	qtx := tw.queries.WithTx(tx)

	_, err = qtx.CreateTransaction(ctx, schema.CreateTransactionParams{
		UserID:          pgtype.Int8{Int64: int64(userID), Valid: true},
		MerchantID:      pgtype.Int8{Int64: int64(merchantID), Valid: true},
		CoinsSpent:      utils.Float64ToNumeric(amount),
		OriginalAmount:  utils.Float64ToNumeric(amount),
		TransactionType: pgtype.Text{String: paymentType, Valid: true},
		Status:          pgtype.Text{String: "completed", Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to record transaction: %v", err)
	}

	// Commit PostgreSQL transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (tw *TransactionWrapper) AddCoinsWithRecord(ctx context.Context, userID int, amount float64, source string) error {
	tx, err := tw.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// Add coins in TigerBeetle
	err = tw.tb.AddCoins(userID, amount)
	if err != nil {
		return fmt.Errorf("failed to add coins: %v", err)
	}

	// Record coin purchase using sqlc
	qtx := tw.queries.WithTx(tx)

	_, err = qtx.CreateCoinPurchase(ctx, schema.CreateCoinPurchaseParams{
		UserID:        pgtype.Int8{Int64: int64(userID), Valid: true},
		Amount:        utils.Float64ToNumeric(amount),
		CoinsReceived: utils.Float64ToNumeric(amount),
		PaymentMethod: pgtype.Text{String: source, Valid: true},
		Status:        pgtype.Text{String: "completed", Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to record coin purchase: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
