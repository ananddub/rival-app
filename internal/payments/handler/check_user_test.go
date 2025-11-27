package handler

import (
	"context"
	"testing"

	"rival/config"
	"rival/connection"
	paymentpb "rival/gen/proto/proto/api"
	schema "rival/gen/sql"
	
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCheckUser3702History(t *testing.T) {
	ctx := context.Background()
	
	userID := int64(3702)
	
	// Check DB directly
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	queries := schema.New(db)
	
	// Check if user exists
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("User 3702 not found: %v", err)
	}
	t.Logf("User found: ID=%d, Email=%s", user.ID, user.Email)
	
	// Check coin purchases directly from DB
	purchases, err := queries.GetUserCoinPurchases(ctx, schema.GetUserCoinPurchasesParams{
		UserID: pgtype.Int8{Int64: user.ID, Valid: true},
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to get purchases from DB: %v", err)
	}
	t.Logf("DB Query Result: %d purchases found", len(purchases))
	
	for i, p := range purchases {
		t.Logf("Purchase %d: ID=%d, Amount=%v, Coins=%v, Method=%s, Status=%s, Created=%v",
			i+1, p.ID, p.Amount, p.CoinsReceived, p.PaymentMethod.String, p.Status.String, p.CreatedAt.Time)
	}
	
	// Now check via API
	h, err := NewPaymentHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	
	// Check balance
	balanceResp, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{
		UserId: userID,
	})
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}
	t.Logf("Balance: %.2f", balanceResp.Balance)
	
	// Check history via API
	historyResp, err := h.GetPaymentHistory(ctx, &paymentpb.GetPaymentHistoryRequest{
		UserId: userID,
		Page:   1,
		Limit:  100,
	})
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}
	
	t.Logf("API Result: %d purchases in history", len(historyResp.Purchases))
	
	for i, p := range historyResp.Purchases {
		t.Logf("History %d: ID=%d, Amount=%.2f, Coins=%.2f, Method=%s, Status=%s",
			i+1, p.Id, p.Amount, p.CoinsReceived, p.PaymentMethod, p.Status)
	}
	
	if len(purchases) != len(historyResp.Purchases) {
		t.Fatalf("Mismatch! DB has %d purchases but API returned %d", len(purchases), len(historyResp.Purchases))
	}
	
	t.Logf("✓ History working correctly!")
}
