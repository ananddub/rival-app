package handler

import (
	"context"
	"testing"

	paymentpb "rival/gen/proto/proto/api"
)

func TestHistoryMaintenance_Complete(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-history-complete@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

	// Step 1: Purchase coins
	purchaseReq := &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(user.ID),
		Amount:        100.0,
		PaymentMethod: "stripe",
	}
	purchaseResp, err := h.InitiateCoinPurchase(ctx, purchaseReq)
	if err != nil {
		t.Fatalf("Failed to initiate purchase: %v", err)
	}
	t.Logf("✓ Purchase initiated: PaymentID=%s", purchaseResp.PaymentId)

	// Step 2: Verify payment
	_, err = h.VerifyPayment(ctx, &paymentpb.VerifyPaymentRequest{
		PaymentId: purchaseResp.PaymentId,
	})
	if err != nil {
		t.Fatalf("Failed to verify payment: %v", err)
	}
	t.Logf("✓ Payment verified")

	// Step 3: Check payment history
	historyResp, err := h.GetPaymentHistory(ctx, &paymentpb.GetPaymentHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Failed to get payment history: %v", err)
	}

	if len(historyResp.Purchases) == 0 {
		t.Fatal("❌ Payment history is empty - history NOT maintained!")
	}

	t.Logf("✓ Payment history maintained: %d purchases found", len(historyResp.Purchases))
	for i, purchase := range historyResp.Purchases {
		t.Logf("  Purchase %d: Amount=%.2f, Coins=%.2f, Status=%s, Method=%s",
			i+1, purchase.Amount, purchase.CoinsReceived, purchase.Status, purchase.PaymentMethod)
	}

	// Step 4: Create a merchant and make payment
	_, repo2, merchantUser := NewUser(ctx, "test-merchant-history@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchantUser.Email, repo2, t)
		repo2.DleteUser(ctx, merchantUser.ID)
	}()

	merchant := CreateMerchantRecord(ctx, merchantUser, repo2, t)

	paymentReq := &paymentpb.PayToMerchantRequest{
		UserId:     int64(user.ID),
		MerchantId: int64(merchant.ID),
		Amount:     10.0,
		OrderId:    "test-order-history",
	}
	paymentResp, err := h.PayToMerchant(ctx, paymentReq)
	if err != nil {
		t.Fatalf("Failed to pay merchant: %v", err)
	}
	t.Logf("✓ Payment to merchant successful: TransactionID=%s", paymentResp.TransactionId)

	// Step 5: Check transaction history
	txHistoryResp, err := h.GetTransactionHistory(ctx, &paymentpb.GetTransactionHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Failed to get transaction history: %v", err)
	}

	if len(txHistoryResp.Transactions) == 0 {
		t.Fatal("❌ Transaction history is empty - history NOT maintained!")
	}

	t.Logf("✓ Transaction history maintained: %d transactions found", len(txHistoryResp.Transactions))
	for i, tx := range txHistoryResp.Transactions {
		t.Logf("  Transaction %d: ID=%d, Amount=%.2f, Status=%s, MerchantID=%d",
			i+1, tx.Id, tx.CoinsSpent, tx.Status, tx.MerchantId)
	}

	// Step 6: Transfer to another user
	_, repo3, receiver := NewUser(ctx, "test-receiver-history@example.com", t)
	defer repo3.DleteUser(ctx, receiver.ID)

	transferReq := &paymentpb.TransferToUserRequest{
		FromUserId:  int64(user.ID),
		ToUserId:    int64(receiver.ID),
		Amount:      5.0,
		Description: "Test transfer for history",
	}
	transferResp, err := h.TransferToUser(ctx, transferReq)
	if err != nil {
		t.Fatalf("Failed to transfer: %v", err)
	}
	t.Logf("✓ Transfer successful: TransactionID=%s", transferResp.TransactionId)

	// Step 7: Check updated transaction history
	txHistoryResp2, err := h.GetTransactionHistory(ctx, &paymentpb.GetTransactionHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Failed to get updated transaction history: %v", err)
	}

	if len(txHistoryResp2.Transactions) < 2 {
		t.Fatalf("❌ Transaction history not updated properly. Expected at least 2, got %d", len(txHistoryResp2.Transactions))
	}

	t.Logf("✓ Transaction history updated: %d total transactions", len(txHistoryResp2.Transactions))

	// Final Summary
	t.Log("\n=== HISTORY MAINTENANCE SUMMARY ===")
	t.Logf("✓ Payment History: %d purchases recorded", len(historyResp.Purchases))
	t.Logf("✓ Transaction History: %d transactions recorded", len(txHistoryResp2.Transactions))
	t.Log("✓ ALL HISTORY IS PROPERLY MAINTAINED!")
}
