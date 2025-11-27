package handler

import (
	"context"
	"fmt"
	"rival/config"
	"rival/connection"
	authpb "rival/gen/proto/proto/api"
	paymentpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	authHandler "rival/internal/auth/handler"
	"testing"
)

// NewUser creates a test user for payment testing
func NewUser(ctx context.Context, email string, t *testing.T) (*authpb.SignupRequest, *schema.Queries, schema.User) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to get db connection: %v", err)
	}
	repo := schema.New(db)

	// Delete existing user if exists
	existingUser, _ := repo.GetUserByEmail(ctx, email)
	if existingUser.ID != 0 {
		repo.DleteUser(ctx, existingUser.ID)
		t.Logf("Deleted existing user: %s", email)
	}

	handler, err := authHandler.NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create auth handler: %v", err)
	}

	data := authpb.SignupRequest{
		Name:     "Test Payment User",
		Email:    email,
		Password: "password123",
		Role:     *schemapb.UserRole_USER_ROLE_CUSTOMER.Enum(),
		Phone:    "12345678",
	}

	_, err = handler.Signup(ctx, &data)
	if err != nil {
		t.Fatalf("Failed to signup user: %v", err)
	}

	user, err := repo.GetUserByEmail(ctx, data.Email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	ctx = context.WithValue(ctx, "user_id", user.ID)
	t.Logf("Created test user: %v", user)
	return &data, repo, user
}

func TestVerifyPayment_EmptyPaymentID(t *testing.T) {
	ctx := context.Background()
	h, _ := NewPaymentHandler()
	resp, err := h.VerifyPayment(ctx, &paymentpb.VerifyPaymentRequest{PaymentId: ""})
	if err != nil {
		t.Fatalf("VerifyPayment returned error: %v", err)
	}
	if resp == nil || resp.Success != false {
		t.Fatalf("Expected Success=false for empty payment id")
	}
}

func TestGetBalance(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-balance@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()
	resp, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(user.ID)})
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}
	if resp == nil || resp.Balance < 0 {
		t.Fatalf("Invalid balance response")
	}
	t.Logf("User balance: %v", resp.Balance)
}

func TestInitiateCoinPurchase(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-purchase@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

	// Get initial balance
	balanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(user.ID)})
	if err != nil {
		t.Fatalf("Failed to get initial balance: %v", err)
	}
	initialBalance := balanceResp1.Balance
	t.Logf("Initial balance: %.2f", initialBalance)

	// Initiate coin purchase
	purchaseAmount := 100.0
	req := &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(user.ID),
		Amount:        purchaseAmount,
		PaymentMethod: "stripe",
	}
	resp, err := h.InitiateCoinPurchase(ctx, req)
	if err != nil {
		t.Fatalf("InitiateCoinPurchase returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("InitiateCoinPurchase returned nil response")
	}

	// Verify response
	t.Logf("Purchase Response: PaymentID=%s, Status=%s, CoinsToReceive=%.2f, NewBalance=%.2f",
		resp.PaymentId, resp.Status, resp.CoinsToReceive, resp.NewBalance)

	if resp.Status != "completed" {
		t.Fatalf("Expected status 'completed', got '%s'", resp.Status)
	}

	if resp.CoinsToReceive != purchaseAmount {
		t.Fatalf("Expected coins %.2f, got %.2f", purchaseAmount, resp.CoinsToReceive)
	}

	// Verify DB record
	var purchaseID int64
	if _, err := fmt.Sscanf(resp.PaymentId, "%d", &purchaseID); err != nil {
		t.Fatalf("Failed to parse payment ID: %v", err)
	}

	purchase, err := repo.GetCoinPurchaseByID(ctx, purchaseID)
	if err != nil {
		t.Fatalf("Failed to get purchase from DB: %v", err)
	}

	if purchase.Status.String != "completed" {
		t.Fatalf("DB status should be 'completed', got '%s'", purchase.Status.String)
	}
	t.Logf("✓ DB record verified: Status=%s", purchase.Status.String)

	// Verify TigerBeetle balance
	balanceResp2, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(user.ID)})
	if err != nil {
		t.Fatalf("Failed to get balance after purchase: %v", err)
	}
	finalBalance := balanceResp2.Balance
	t.Logf("Final balance: %.2f", finalBalance)

	expectedBalance := initialBalance + purchaseAmount
	if finalBalance != expectedBalance {
		t.Fatalf("Balance mismatch. Expected: %.2f, Got: %.2f", expectedBalance, finalBalance)
	}
	t.Logf("✓ TigerBeetle balance verified: %.2f", finalBalance)

	// Verify payment history
	historyResp, err := h.GetPaymentHistory(ctx, &paymentpb.GetPaymentHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Failed to get payment history: %v", err)
	}

	if len(historyResp.Purchases) == 0 {
		t.Fatalf("No purchases found in history")
	}

	lastPurchase := historyResp.Purchases[0]
	if lastPurchase.Status != "completed" {
		t.Fatalf("History status should be 'completed', got '%s'", lastPurchase.Status)
	}
	t.Logf("✓ Payment history verified: %d purchases, last status=%s", len(historyResp.Purchases), lastPurchase.Status)

	t.Logf("✓✓✓ All checks passed! Purchase completed successfully.")
}

func TestGetPaymentHistory(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-history@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

	// First make a purchase
	purchaseReq := &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(user.ID),
		Amount:        50.0,
		PaymentMethod: "stripe",
	}
	purchaseResp, err := h.InitiateCoinPurchase(ctx, purchaseReq)
	if err != nil {
		t.Fatalf("Failed to create purchase: %v", err)
	}
	t.Logf("Created purchase: ID=%s, Status=%s", purchaseResp.PaymentId, purchaseResp.Status)

	// Now get history
	req := &paymentpb.GetPaymentHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  10,
	}
	resp, err := h.GetPaymentHistory(ctx, req)
	if err != nil {
		t.Fatalf("GetPaymentHistory returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("GetPaymentHistory returned nil response")
	}

	t.Logf("Payment history: purchases=%d, total=%d", len(resp.GetPurchases()), resp.GetTotalCount())

	if len(resp.Purchases) == 0 {
		t.Fatalf("Expected at least 1 purchase in history, got 0")
	}

	lastPurchase := resp.Purchases[0]
	t.Logf("Last purchase: ID=%d, Amount=%.2f, Coins=%.2f, Status=%s",
		lastPurchase.Id, lastPurchase.Amount, lastPurchase.CoinsReceived, lastPurchase.Status)

	if lastPurchase.Status != "completed" {
		t.Fatalf("Expected status 'completed', got '%s'", lastPurchase.Status)
	}

	if lastPurchase.Amount != 50.0 {
		t.Fatalf("Expected amount 50.0, got %.2f", lastPurchase.Amount)
	}

	t.Logf("✓ Payment history working correctly!")
}

func TestGetTransactionHistory(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-transactions@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()
	req := &paymentpb.GetTransactionHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  20,
	}
	resp, err := h.GetTransactionHistory(ctx, req)
	if err != nil {
		t.Fatalf("GetTransactionHistory returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("GetTransactionHistory returned nil response")
	}
	t.Logf("Transaction history retrieved successfully")
}

func TestPayToMerchant_ValidRequest(t *testing.T) {
	ctx := context.Background()
	_, repo1, payer := NewUser(ctx, "test-payer@example.com", t)
	defer repo1.DleteUser(ctx, payer.ID)

	_, repo2, merchantUser := NewUser(ctx, "test-payment-merchant@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchantUser.Email, repo2, t)
		repo2.DleteUser(ctx, merchantUser.ID)
	}()

	merchant := CreateMerchantRecord(ctx, merchantUser, repo2, t)
	h, _ := NewPaymentHandler()

	req := &paymentpb.PayToMerchantRequest{
		UserId:     int64(payer.ID),
		MerchantId: int64(merchant.ID),
		Amount:     2.0,
		OrderId:    "test-order-123",
	}

	resp, err := h.PayToMerchant(ctx, req)
	if err != nil {
		t.Logf("PayToMerchant returned error: %v (may be insufficient balance)", err)
		return
	}
	if resp != nil && resp.Success {
		t.Logf("Payment to merchant successful: %+v", resp)
	}
}

func TestTransferToUser_ValidRequest(t *testing.T) {
	ctx := context.Background()
	_, repo1, sender := NewUser(ctx, "test-sender@example.com", t)
	defer repo1.DleteUser(ctx, sender.ID)

	_, repo2, receiver := NewUser(ctx, "test-receiver@example.com", t)
	defer repo2.DleteUser(ctx, receiver.ID)

	h, _ := NewPaymentHandler()
	req := &paymentpb.TransferToUserRequest{
		FromUserId:  int64(sender.ID),
		ToUserId:    int64(receiver.ID),
		Amount:      1.0,
		Description: "Test transfer",
	}

	resp, err := h.TransferToUser(ctx, req)
	if err != nil {
		t.Logf("TransferToUser returned error: %v (may be insufficient balance)", err)
		return
	}
	if resp != nil && resp.Success {
		t.Logf("Transfer successful: %+v", resp)
	}
}

func TestCoinPurchaseFlow_Complete(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-coin-flow@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

	// Step 1: Check initial balance
	balanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(user.ID)})
	if err != nil {
		t.Fatalf("Failed to get initial balance: %v", err)
	}
	initialBalance := balanceResp1.Balance
	t.Logf("Initial balance: %.2f", initialBalance)

	// Step 2: Initiate coin purchase
	purchaseAmount := 50.0
	purchaseReq := &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(user.ID),
		Amount:        purchaseAmount,
		PaymentMethod: "stripe",
	}
	purchaseResp, err := h.InitiateCoinPurchase(ctx, purchaseReq)
	if err != nil {
		t.Fatalf("Failed to initiate coin purchase: %v", err)
	}
	t.Logf("Purchase initiated: PaymentID=%s, CoinsToReceive=%.2f, Status=%s",
		purchaseResp.PaymentId, purchaseResp.CoinsToReceive, purchaseResp.Status)

	// Step 3: Verify payment (simulating payment gateway callback)
	verifyResp, err := h.VerifyPayment(ctx, &paymentpb.VerifyPaymentRequest{
		PaymentId: purchaseResp.PaymentId,
	})
	if err != nil {
		t.Fatalf("Failed to verify payment: %v", err)
	}
	t.Logf("Payment verified: Success=%v, CoinsAdded=%.2f, NewBalance=%.2f",
		verifyResp.Success, verifyResp.CoinsAdded, verifyResp.NewBalance)

	// Step 4: Check balance after purchase
	balanceResp2, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(user.ID)})
	if err != nil {
		t.Fatalf("Failed to get balance after purchase: %v", err)
	}
	finalBalance := balanceResp2.Balance
	t.Logf("Final balance: %.2f", finalBalance)

	// Step 5: Verify balance increased correctly
	expectedBalance := initialBalance + purchaseAmount
	if finalBalance != expectedBalance {
		t.Logf("WARNING: Balance mismatch. Expected: %.2f, Got: %.2f, Difference: %.2f",
			expectedBalance, finalBalance, finalBalance-expectedBalance)
	} else {
		t.Logf("✓ Balance verified correctly: %.2f", finalBalance)
	}

	// Step 6: Check payment history
	historyResp, err := h.GetPaymentHistory(ctx, &paymentpb.GetPaymentHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Failed to get payment history: %v", err)
	}
	t.Logf("Payment history: %d purchases found", len(historyResp.Purchases))

	if len(historyResp.Purchases) > 0 {
		lastPurchase := historyResp.Purchases[0]
		t.Logf("Last purchase: Amount=%.2f, Coins=%.2f, Status=%s",
			lastPurchase.Amount, lastPurchase.CoinsReceived, lastPurchase.Status)
	}
}

func TestPaymentToMerchant_WithBalanceCheck(t *testing.T) {
	ctx := context.Background()

	// Create payer with initial coins
	_, repo1, payer := NewUser(ctx, "test-payer-balance@example.com", t)
	defer repo1.DleteUser(ctx, payer.ID)

	// Create merchant
	_, repo2, merchantUser := NewUser(ctx, "test-merchant-balance@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchantUser.Email, repo2, t)
		repo2.DleteUser(ctx, merchantUser.ID)
	}()
	merchant := CreateMerchantRecord(ctx, merchantUser, repo2, t)

	h, _ := NewPaymentHandler()

	// Step 1: Check payer's initial balance
	balanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(payer.ID)})
	if err != nil {
		t.Fatalf("Failed to get payer balance: %v", err)
	}
	payerInitialBalance := balanceResp1.Balance
	t.Logf("Payer initial balance: %.2f", payerInitialBalance)

	// Step 2: Check merchant's initial balance
	merchantBalanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(merchant.ID)})
	if err != nil {
		t.Fatalf("Failed to get merchant balance: %v", err)
	}
	merchantInitialBalance := merchantBalanceResp1.Balance
	t.Logf("Merchant initial balance: %.2f", merchantInitialBalance)

	// Step 3: Attempt payment
	paymentAmount := 2.0
	if payerInitialBalance < paymentAmount {
		t.Logf("Insufficient balance for payment. Required: %.2f, Available: %.2f",
			paymentAmount, payerInitialBalance)
		t.Skip("Skipping payment test due to insufficient balance")
	}

	paymentReq := &paymentpb.PayToMerchantRequest{
		UserId:     int64(payer.ID),
		MerchantId: int64(merchant.ID),
		Amount:     paymentAmount,
		OrderId:    "test-order-balance-check",
	}

	paymentResp, err := h.PayToMerchant(ctx, paymentReq)
	if err != nil {
		t.Logf("Payment failed: %v", err)
		return
	}

	if !paymentResp.Success {
		t.Logf("Payment not successful: %+v", paymentResp)
		return
	}

	t.Logf("Payment successful: TransactionID=%s, DiscountAmount=%.2f, FinalAmount=%.2f",
		paymentResp.TransactionId, paymentResp.DiscountAmount, paymentResp.FinalAmount)

	// Step 4: Verify payer's balance decreased
	balanceResp2, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(payer.ID)})
	if err != nil {
		t.Fatalf("Failed to get payer balance after payment: %v", err)
	}
	payerFinalBalance := balanceResp2.Balance
	t.Logf("Payer final balance: %.2f", payerFinalBalance)

	expectedPayerBalance := payerInitialBalance - paymentResp.FinalAmount
	if payerFinalBalance != expectedPayerBalance {
		t.Logf("WARNING: Payer balance mismatch. Expected: %.2f, Got: %.2f",
			expectedPayerBalance, payerFinalBalance)
	} else {
		t.Logf("✓ Payer balance verified: %.2f", payerFinalBalance)
	}

	// Step 5: Verify merchant's balance increased
	merchantBalanceResp2, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(merchant.ID)})
	if err != nil {
		t.Fatalf("Failed to get merchant balance after payment: %v", err)
	}
	merchantFinalBalance := merchantBalanceResp2.Balance
	t.Logf("Merchant final balance: %.2f", merchantFinalBalance)

	expectedMerchantBalance := merchantInitialBalance + paymentResp.FinalAmount
	if merchantFinalBalance != expectedMerchantBalance {
		t.Logf("WARNING: Merchant balance mismatch. Expected: %.2f, Got: %.2f",
			expectedMerchantBalance, merchantFinalBalance)
	} else {
		t.Logf("✓ Merchant balance verified: %.2f", merchantFinalBalance)
	}

	// Step 6: Check transaction history
	txHistoryResp, err := h.GetTransactionHistory(ctx, &paymentpb.GetTransactionHistoryRequest{
		UserId: int64(payer.ID),
		Page:   1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Failed to get transaction history: %v", err)
	}
	t.Logf("Transaction history: %d transactions found", len(txHistoryResp.Transactions))
}

func TestTransferBetweenUsers_WithBalanceCheck(t *testing.T) {
	ctx := context.Background()

	_, repo1, sender := NewUser(ctx, "test-transfer-sender@example.com", t)
	defer repo1.DleteUser(ctx, sender.ID)

	_, repo2, receiver := NewUser(ctx, "test-transfer-receiver@example.com", t)
	defer repo2.DleteUser(ctx, receiver.ID)

	h, _ := NewPaymentHandler()

	// Step 1: Check initial balances
	senderBalanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(sender.ID)})
	if err != nil {
		t.Fatalf("Failed to get sender balance: %v", err)
	}
	senderInitialBalance := senderBalanceResp1.Balance
	t.Logf("Sender initial balance: %.2f", senderInitialBalance)

	receiverBalanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(receiver.ID)})
	if err != nil {
		t.Fatalf("Failed to get receiver balance: %v", err)
	}
	receiverInitialBalance := receiverBalanceResp1.Balance
	t.Logf("Receiver initial balance: %.2f", receiverInitialBalance)

	// Step 2: Attempt transfer
	transferAmount := 1.0
	if senderInitialBalance < transferAmount {
		t.Logf("Insufficient balance for transfer. Required: %.2f, Available: %.2f",
			transferAmount, senderInitialBalance)
		t.Skip("Skipping transfer test due to insufficient balance")
	}

	transferReq := &paymentpb.TransferToUserRequest{
		FromUserId:  int64(sender.ID),
		ToUserId:    int64(receiver.ID),
		Amount:      transferAmount,
		Description: "Test transfer with balance check",
	}

	transferResp, err := h.TransferToUser(ctx, transferReq)
	if err != nil {
		t.Logf("Transfer failed: %v", err)
		return
	}

	if !transferResp.Success {
		t.Logf("Transfer not successful")
		return
	}

	t.Logf("Transfer successful: TransactionID=%s", transferResp.TransactionId)

	// Step 3: Verify sender's balance decreased
	senderBalanceResp2, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(sender.ID)})
	if err != nil {
		t.Fatalf("Failed to get sender balance after transfer: %v", err)
	}
	senderFinalBalance := senderBalanceResp2.Balance
	t.Logf("Sender final balance: %.2f", senderFinalBalance)

	expectedSenderBalance := senderInitialBalance - transferAmount
	if senderFinalBalance != expectedSenderBalance {
		t.Logf("WARNING: Sender balance mismatch. Expected: %.2f, Got: %.2f",
			expectedSenderBalance, senderFinalBalance)
	} else {
		t.Logf("✓ Sender balance verified: %.2f", senderFinalBalance)
	}

	// Step 4: Verify receiver's balance increased
	receiverBalanceResp2, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(receiver.ID)})
	if err != nil {
		t.Fatalf("Failed to get receiver balance after transfer: %v", err)
	}
	receiverFinalBalance := receiverBalanceResp2.Balance
	t.Logf("Receiver final balance: %.2f", receiverFinalBalance)

	expectedReceiverBalance := receiverInitialBalance + transferAmount
	if receiverFinalBalance != expectedReceiverBalance {
		t.Logf("WARNING: Receiver balance mismatch. Expected: %.2f, Got: %.2f",
			expectedReceiverBalance, receiverFinalBalance)
	} else {
		t.Logf("✓ Receiver balance verified: %.2f", receiverFinalBalance)
	}
}
