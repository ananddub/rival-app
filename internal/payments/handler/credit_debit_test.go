package handler

import (
	"context"
	"math/big"
	"testing"

	paymentpb "rival/gen/proto/proto/api"
	schema "rival/gen/sql"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreditDebitHistory(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-credit-debit@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

	// Create purchase (credit)
	_, err := h.InitiateCoinPurchase(ctx, &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(user.ID),
		Amount:        500,
		PaymentMethod: "stripe",
	})
	if err != nil {
		t.Fatalf("Purchase failed: %v", err)
	}

	// Create merchant
	merchant, err := repo.CreateMerchant(ctx, schema.CreateMerchantParams{
		Name:               "Test Merchant",
		Email:              "merchant@test.com",
		Phone:              pgtype.Text{String: "1234567890", Valid: true},
		Category:           pgtype.Text{String: "restaurant", Valid: true},
		DiscountPercentage: pgtype.Numeric{Int: big.NewInt(10), Exp: 0, Valid: true},
		IsActive:           pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("Merchant creation failed: %v", err)
	}
	defer repo.DeleteMerchant(ctx, merchant.ID)

	// Make payment (debit)
	_, err = h.PayToMerchant(ctx, &paymentpb.PayToMerchantRequest{
		UserId:     int64(user.ID),
		MerchantId: int64(merchant.ID),
		Amount:     100,
	})
	if err != nil {
		t.Fatalf("Payment failed: %v", err)
	}

	// Get financial history
	resp, err := h.GetFinancialHistory(ctx, &paymentpb.GetFinancialHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  20,
		Type:   "all",
	})
	if err != nil {
		t.Fatalf("GetFinancialHistory failed: %v", err)
	}

	t.Logf("Balance: %.2f, Items: %d", resp.CurrentBalance, len(resp.Items))
	
	hasCredit := false
	hasDebit := false
	
	for i, item := range resp.Items {
		t.Logf("%d. Type=%s, Amount=%.2f, Desc=%s", i+1, item.Type, item.Amount, item.Description)
		if item.Type == "credit" {
			hasCredit = true
		}
		if item.Type == "debit" {
			hasDebit = true
		}
	}

	if !hasCredit {
		t.Error("No credit transactions found!")
	}
	if !hasDebit {
		t.Error("No debit transactions found!")
	}

	t.Logf("✓ Has credit: %v, Has debit: %v", hasCredit, hasDebit)
}
