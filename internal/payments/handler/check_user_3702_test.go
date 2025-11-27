package handler

import (
	"context"
	"testing"

	paymentpb "rival/gen/proto/proto/api"
)

func TestUser3702Financial(t *testing.T) {
	ctx := context.Background()
	h, _ := NewPaymentHandler()

	resp, err := h.GetFinancialHistory(ctx, &paymentpb.GetFinancialHistoryRequest{
		UserId: 3702,
		Page:   1,
		Limit:  100,
		Type:   "all",
	})

	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	t.Logf("User 3702 - Balance: %.2f, Items: %d", resp.CurrentBalance, len(resp.Items))
	for i, item := range resp.Items {
		t.Logf("%d. ID=%s, Type=%s, Amount=%.2f, Coins=%.2f, Desc=%s",
			i+1, item.Id, item.Type, item.Amount, item.Coins, item.Description)
	}
}
