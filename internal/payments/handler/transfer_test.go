package handler

import (
	"context"
	"testing"

	paymentpb "rival/gen/proto/proto/api"
)

func TestTransferToUser(t *testing.T) {
	ctx := context.Background()
	
	// Create sender
	_, repo, sender := NewUser(ctx, "sender@test.com", t)
	defer repo.DleteUser(ctx, sender.ID)
	
	// Create receiver
	_, _, receiver := NewUser(ctx, "receiver@test.com", t)
	defer repo.DleteUser(ctx, receiver.ID)

	h, _ := NewPaymentHandler()

	// Add coins to sender
	_, err := h.InitiateCoinPurchase(ctx, &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(sender.ID),
		Amount:        500,
		PaymentMethod: "stripe",
	})
	if err != nil {
		t.Fatalf("Purchase failed: %v", err)
	}

	// Transfer from sender to receiver
	transferResp, err := h.TransferToUser(ctx, &paymentpb.TransferToUserRequest{
		FromUserId: int64(sender.ID),
		ToUserId:   int64(receiver.ID),
		Amount:     100,
	})
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	t.Logf("Transfer successful: %v, Sender balance: %.2f", transferResp.Success, transferResp.RemainingBalance)

	// Check sender's financial history
	senderHistory, err := h.GetFinancialHistory(ctx, &paymentpb.GetFinancialHistoryRequest{
		UserId: int64(sender.ID),
		Page:   1,
		Limit:  20,
		Type:   "all",
	})
	if err != nil {
		t.Fatalf("Failed to get sender history: %v", err)
	}

	t.Logf("\nSender (ID=%d) - Balance: %.2f, Items: %d", sender.ID, senderHistory.CurrentBalance, len(senderHistory.Items))
	for i, item := range senderHistory.Items {
		t.Logf("  %d. Type=%s, Amount=%.2f", i+1, item.Type, item.Amount)
	}

	// Check receiver's financial history
	receiverHistory, err := h.GetFinancialHistory(ctx, &paymentpb.GetFinancialHistoryRequest{
		UserId: int64(receiver.ID),
		Page:   1,
		Limit:  20,
		Type:   "all",
	})
	if err != nil {
		t.Fatalf("Failed to get receiver history: %v", err)
	}

	t.Logf("\nReceiver (ID=%d) - Balance: %.2f, Items: %d", receiver.ID, receiverHistory.CurrentBalance, len(receiverHistory.Items))
	for i, item := range receiverHistory.Items {
		t.Logf("  %d. Type=%s, Amount=%.2f", i+1, item.Type, item.Amount)
	}

	// Verify balances
	expectedSenderBalance := 10.0 + 500.0 - 100.0 // signup + purchase - transfer
	expectedReceiverBalance := 10.0 + 100.0       // signup + transfer

	if senderHistory.CurrentBalance != expectedSenderBalance {
		t.Errorf("Sender balance mismatch: expected %.2f, got %.2f", expectedSenderBalance, senderHistory.CurrentBalance)
	}

	if receiverHistory.CurrentBalance != expectedReceiverBalance {
		t.Errorf("Receiver balance mismatch: expected %.2f, got %.2f", expectedReceiverBalance, receiverHistory.CurrentBalance)
	}

	t.Logf("✓ Transfer logic working correctly!")
}
