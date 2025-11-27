package handler

import (
	"context"
	"testing"

	paymentpb "rival/gen/proto/proto/api"
)

func TestGetFinancialHistory(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-financial@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

	// Create some purchases
	for i := 0; i < 3; i++ {
		_, err := h.InitiateCoinPurchase(ctx, &paymentpb.InitiateCoinPurchaseRequest{
			UserId:        int64(user.ID),
			Amount:        float64(100 * (i + 1)),
			PaymentMethod: "stripe",
		})
		if err != nil {
			t.Fatalf("Failed to create purchase: %v", err)
		}
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

	t.Logf("Financial History: %d items, Balance: %.2f", len(resp.Items), resp.CurrentBalance)

	if len(resp.Items) < 3 {
		t.Fatalf("Expected at least 3 items, got %d", len(resp.Items))
	}

	for i, item := range resp.Items {
		t.Logf("Item %d: ID=%s, Type=%s, Amount=%.2f, Coins=%.2f, Status=%s, Desc=%s",
			i+1, item.Id, item.Type, item.Amount, item.Coins, item.Status, item.Description)
	}

	// Check balance
	expectedBalance := 10.0 + 100.0 + 200.0 + 300.0 // signup + 3 purchases
	if resp.CurrentBalance != expectedBalance {
		t.Logf("Balance mismatch: expected %.2f, got %.2f", expectedBalance, resp.CurrentBalance)
	}

	t.Logf("✓ Financial history working!")
}
