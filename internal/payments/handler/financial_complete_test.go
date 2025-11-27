package handler

import (
	"context"
	"math/big"
	"testing"

	paymentpb "rival/gen/proto/proto/api"
	schema "rival/gen/sql"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestGetFinancialHistoryComplete(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "complete-test@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	_, _, user2 := NewUser(ctx, "user2@example.com", t)
	defer repo.DleteUser(ctx, user2.ID)

	h, _ := NewPaymentHandler()

	// 1. Purchase (credit)
	_, err := h.InitiateCoinPurchase(ctx, &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(user.ID),
		Amount:        1000,
		PaymentMethod: "stripe",
	})
	if err != nil {
		t.Fatalf("Purchase failed: %v", err)
	}

	// 2. Create merchant and make payment (debit)
	merchant, err := repo.CreateMerchant(ctx, schema.CreateMerchantParams{
		Name:               "Test Merchant",
		Email:              "merchant-complete@test.com",
		Phone:              pgtype.Text{String: "1234567890", Valid: true},
		Category:           pgtype.Text{String: "restaurant", Valid: true},
		DiscountPercentage: pgtype.Numeric{Int: big.NewInt(10), Exp: 0, Valid: true},
		IsActive:           pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("Merchant creation failed: %v", err)
	}
	defer repo.DeleteMerchant(ctx, merchant.ID)

	_, err = h.PayToMerchant(ctx, &paymentpb.PayToMerchantRequest{
		UserId:     int64(user.ID),
		MerchantId: int64(merchant.ID),
		Amount:     200,
	})
	if err != nil {
		t.Fatalf("Payment failed: %v", err)
	}

	// 3. Transfer to another user (debit for sender, credit for receiver)
	_, err = h.TransferToUser(ctx, &paymentpb.TransferToUserRequest{
		FromUserId: int64(user.ID),
		ToUserId:   int64(user2.ID),
		Amount:     100,
	})
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	// Get financial history
	resp, err := h.GetFinancialHistory(ctx, &paymentpb.GetFinancialHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  50,
		Type:   "all",
	})
	if err != nil {
		t.Fatalf("GetFinancialHistory failed: %v", err)
	}

	t.Logf("\n=== User Financial History ===")
	t.Logf("Balance: %.2f coins", resp.CurrentBalance)
	t.Logf("Total Items: %d\n", len(resp.Items))

	creditCount := 0
	debitCount := 0
	totalCredit := 0.0
	totalDebit := 0.0

	for i, item := range resp.Items {
		t.Logf("%d. Type=%s, Amount=%.2f, Desc=%s", i+1, item.Type, item.Amount, item.Description)
		
		if item.Type == "credit" {
			creditCount++
			totalCredit += item.Amount
		} else if item.Type == "debit" {
			debitCount++
			totalDebit += item.Amount
		}
	}

	t.Logf("\n=== Summary ===")
	t.Logf("Credits: %d transactions, Total: %.2f", creditCount, totalCredit)
	t.Logf("Debits: %d transactions, Total: %.2f", debitCount, totalDebit)
	t.Logf("Net: %.2f (should match balance)", totalCredit-totalDebit)

	// Verify we have both credit and debit
	if creditCount == 0 {
		t.Error("❌ No credit transactions found!")
	}
	if debitCount == 0 {
		t.Error("❌ No debit transactions found!")
	}

	// Expected: signup(10) + purchase(1000) - payment(180 after 10% discount) - transfer(100) = 730
	expectedBalance := 10.0 + 1000.0 - 180.0 - 100.0
	if resp.CurrentBalance != expectedBalance {
		t.Errorf("❌ Balance mismatch: expected %.2f, got %.2f", expectedBalance, resp.CurrentBalance)
	}

	// Test type filtering
	t.Logf("\n=== Testing Type Filters ===")
	
	// Filter credit only
	creditResp, _ := h.GetFinancialHistory(ctx, &paymentpb.GetFinancialHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  50,
		Type:   "credit",
	})
	t.Logf("Credit filter: %d items", len(creditResp.Items))
	
	// Filter debit only
	debitResp, _ := h.GetFinancialHistory(ctx, &paymentpb.GetFinancialHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  50,
		Type:   "debit",
	})
	t.Logf("Debit filter: %d items", len(debitResp.Items))

	if len(creditResp.Items) != creditCount {
		t.Errorf("❌ Credit filter mismatch: expected %d, got %d", creditCount, len(creditResp.Items))
	}
	if len(debitResp.Items) != debitCount {
		t.Errorf("❌ Debit filter mismatch: expected %d, got %d", debitCount, len(debitResp.Items))
	}

	t.Logf("\n✓ GetFinancialHistory working correctly!")
}
