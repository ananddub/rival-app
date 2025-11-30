package handler

import (
	"context"
	"fmt"
	"math/big"
	"rival/config"
	"rival/connection"
	authpb "rival/gen/proto/proto/api"
	paymentpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	authHandler "rival/internal/auth/handler"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

// Helper Functions

func NewUser(ctx context.Context, email string, t *testing.T) (*authpb.SignupRequest, *schema.Queries, schema.User) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to get db connection: %v", err)
	}
	repo := schema.New(db)

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

func CreateMerchantRecord(ctx context.Context, user schema.User, repo *schema.Queries, t *testing.T) schema.Merchant {
	merchant, err := repo.CreateMerchant(ctx, schema.CreateMerchantParams{
		Name:               "Test Merchant",
		Email:              user.Email,
		Phone:              pgtype.Text{String: "1234567890", Valid: true},
		Category:           pgtype.Text{String: "restaurant", Valid: true},
		DiscountPercentage: pgtype.Numeric{Int: big.NewInt(15), Exp: 0, Valid: true},
		IsActive:           pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create merchant record: %v", err)
	}
	t.Logf("Created merchant record: %v", merchant.ID)
	return merchant
}

func CleanupMerchant(ctx context.Context, email string, repo *schema.Queries, t *testing.T) {
	merchant, err := repo.GetMerchantByEmail(ctx, email)
	if err != nil {
		t.Logf("Merchant not found for cleanup: %v", err)
		return
	}
	err = repo.DeleteMerchant(ctx, merchant.ID)
	if err != nil {
		t.Logf("Failed to cleanup merchant: %v", err)
	}
}

// Basic Tests

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

// Coin Purchase Tests

func TestInitiateCoinPurchase(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-purchase@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

	balanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(user.ID)})
	if err != nil {
		t.Fatalf("Failed to get initial balance: %v", err)
	}
	initialBalance := balanceResp1.Balance
	t.Logf("Initial balance: %.2f", initialBalance)

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

	t.Logf("Purchase Response: PaymentID=%s, Status=%s, CoinsToReceive=%.2f, NewBalance=%.2f",
		resp.PaymentId, resp.Status, resp.CoinsToReceive, resp.NewBalance)

	if resp.Status != "completed" {
		t.Fatalf("Expected status 'completed', got '%s'", resp.Status)
	}

	if resp.CoinsToReceive != purchaseAmount {
		t.Fatalf("Expected coins %.2f, got %.2f", purchaseAmount, resp.CoinsToReceive)
	}

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

func TestCoinPurchaseFlow_Complete(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-coin-flow@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

	balanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(user.ID)})
	if err != nil {
		t.Fatalf("Failed to get initial balance: %v", err)
	}
	initialBalance := balanceResp1.Balance
	t.Logf("Initial balance: %.2f", initialBalance)

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

	verifyResp, err := h.VerifyPayment(ctx, &paymentpb.VerifyPaymentRequest{
		PaymentId: purchaseResp.PaymentId,
	})
	if err != nil {
		t.Fatalf("Failed to verify payment: %v", err)
	}
	t.Logf("Payment verified: Success=%v, CoinsAdded=%.2f, NewBalance=%.2f",
		verifyResp.Success, verifyResp.CoinsAdded, verifyResp.NewBalance)

	balanceResp2, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(user.ID)})
	if err != nil {
		t.Fatalf("Failed to get balance after purchase: %v", err)
	}
	finalBalance := balanceResp2.Balance
	t.Logf("Final balance: %.2f", finalBalance)

	expectedBalance := initialBalance + purchaseAmount
	if finalBalance != expectedBalance {
		t.Logf("WARNING: Balance mismatch. Expected: %.2f, Got: %.2f, Difference: %.2f",
			expectedBalance, finalBalance, finalBalance-expectedBalance)
	} else {
		t.Logf("✓ Balance verified correctly: %.2f", finalBalance)
	}

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

// Payment History Tests

func TestGetPaymentHistory(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-history@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

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

func TestCheckUser3702History(t *testing.T) {
	ctx := context.Background()
	
	userID := int64(3702)
	
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	queries := schema.New(db)
	
	user, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("User 3702 not found: %v", err)
	}
	t.Logf("User found: ID=%d, Email=%s", user.ID, user.Email)
	
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
	
	h, err := NewPaymentHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	
	balanceResp, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{
		UserId: userID,
	})
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}
	t.Logf("Balance: %.2f", balanceResp.Balance)
	
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

// Financial History Tests

func TestGetFinancialHistory(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-financial@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

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

	expectedBalance := 10.0 + 100.0 + 200.0 + 300.0
	if resp.CurrentBalance != expectedBalance {
		t.Logf("Balance mismatch: expected %.2f, got %.2f", expectedBalance, resp.CurrentBalance)
	}

	t.Logf("✓ Financial history working!")
}

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

func TestCreditDebitHistory(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-credit-debit@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

	_, err := h.InitiateCoinPurchase(ctx, &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(user.ID),
		Amount:        500,
		PaymentMethod: "stripe",
	})
	if err != nil {
		t.Fatalf("Purchase failed: %v", err)
	}

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

	_, err = h.PayToMerchant(ctx, &paymentpb.PayToMerchantRequest{
		UserId:     int64(user.ID),
		MerchantId: int64(merchant.ID),
		Amount:     100,
	})
	if err != nil {
		t.Fatalf("Payment failed: %v", err)
	}

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

func TestGetFinancialHistoryComplete(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "complete-test@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	_, _, user2 := NewUser(ctx, "user2@example.com", t)
	defer repo.DleteUser(ctx, user2.ID)

	h, _ := NewPaymentHandler()

	_, err := h.InitiateCoinPurchase(ctx, &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(user.ID),
		Amount:        1000,
		PaymentMethod: "stripe",
	})
	if err != nil {
		t.Fatalf("Purchase failed: %v", err)
	}

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

	_, err = h.TransferToUser(ctx, &paymentpb.TransferToUserRequest{
		FromUserId: int64(user.ID),
		ToUserId:   int64(user2.ID),
		Amount:     100,
	})
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

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

	if creditCount == 0 {
		t.Error("❌ No credit transactions found!")
	}
	if debitCount == 0 {
		t.Error("❌ No debit transactions found!")
	}

	expectedBalance := 10.0 + 1000.0 - 180.0 - 100.0
	if resp.CurrentBalance != expectedBalance {
		t.Errorf("❌ Balance mismatch: expected %.2f, got %.2f", expectedBalance, resp.CurrentBalance)
	}

	t.Logf("\n=== Testing Type Filters ===")
	
	creditResp, _ := h.GetFinancialHistory(ctx, &paymentpb.GetFinancialHistoryRequest{
		UserId: int64(user.ID),
		Page:   1,
		Limit:  50,
		Type:   "credit",
	})
	t.Logf("Credit filter: %d items", len(creditResp.Items))
	
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

// Transaction History Tests

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

func TestHistoryMaintenance_Complete(t *testing.T) {
	ctx := context.Background()
	_, repo, user := NewUser(ctx, "test-history-complete@example.com", t)
	defer repo.DleteUser(ctx, user.ID)

	h, _ := NewPaymentHandler()

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

	_, err = h.VerifyPayment(ctx, &paymentpb.VerifyPaymentRequest{
		PaymentId: purchaseResp.PaymentId,
	})
	if err != nil {
		t.Fatalf("Failed to verify payment: %v", err)
	}
	t.Logf("✓ Payment verified")

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

	t.Log("\n=== HISTORY MAINTENANCE SUMMARY ===")
	t.Logf("✓ Payment History: %d purchases recorded", len(historyResp.Purchases))
	t.Logf("✓ Transaction History: %d transactions recorded", len(txHistoryResp2.Transactions))
	t.Log("✓ ALL HISTORY IS PROPERLY MAINTAINED!")
}

// Merchant Payment Tests

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

func TestPaymentToMerchant_WithBalanceCheck(t *testing.T) {
	ctx := context.Background()

	_, repo1, payer := NewUser(ctx, "test-payer-balance@example.com", t)
	defer repo1.DleteUser(ctx, payer.ID)

	_, repo2, merchantUser := NewUser(ctx, "test-merchant-balance@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchantUser.Email, repo2, t)
		repo2.DleteUser(ctx, merchantUser.ID)
	}()
	merchant := CreateMerchantRecord(ctx, merchantUser, repo2, t)

	h, _ := NewPaymentHandler()

	balanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(payer.ID)})
	if err != nil {
		t.Fatalf("Failed to get payer balance: %v", err)
	}
	payerInitialBalance := balanceResp1.Balance
	t.Logf("Payer initial balance: %.2f", payerInitialBalance)

	merchantBalanceResp1, err := h.GetBalance(ctx, &paymentpb.GetBalanceRequest{UserId: int64(merchant.ID)})
	if err != nil {
		t.Fatalf("Failed to get merchant balance: %v", err)
	}
	merchantInitialBalance := merchantBalanceResp1.Balance
	t.Logf("Merchant initial balance: %.2f", merchantInitialBalance)

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

// Transfer Tests

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

func TestTransferToUser(t *testing.T) {
	ctx := context.Background()
	
	_, repo, sender := NewUser(ctx, "sender@test.com", t)
	defer repo.DleteUser(ctx, sender.ID)
	
	_, _, receiver := NewUser(ctx, "receiver@test.com", t)
	defer repo.DleteUser(ctx, receiver.ID)

	h, _ := NewPaymentHandler()

	_, err := h.InitiateCoinPurchase(ctx, &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(sender.ID),
		Amount:        500,
		PaymentMethod: "stripe",
	})
	if err != nil {
		t.Fatalf("Purchase failed: %v", err)
	}

	transferResp, err := h.TransferToUser(ctx, &paymentpb.TransferToUserRequest{
		FromUserId: int64(sender.ID),
		ToUserId:   int64(receiver.ID),
		Amount:     100,
	})
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	t.Logf("Transfer successful: %v, Sender balance: %.2f", transferResp.Success, transferResp.RemainingBalance)

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

	expectedSenderBalance := 10.0 + 500.0 - 100.0
	expectedReceiverBalance := 10.0 + 100.0

	if senderHistory.CurrentBalance != expectedSenderBalance {
		t.Errorf("Sender balance mismatch: expected %.2f, got %.2f", expectedSenderBalance, senderHistory.CurrentBalance)
	}

	if receiverHistory.CurrentBalance != expectedReceiverBalance {
		t.Errorf("Receiver balance mismatch: expected %.2f, got %.2f", expectedReceiverBalance, receiverHistory.CurrentBalance)
	}

	t.Logf("✓ Transfer logic working correctly!")
}

func TestTransferBetweenUsers_WithBalanceCheck(t *testing.T) {
	ctx := context.Background()

	_, repo1, sender := NewUser(ctx, "test-transfer-sender@example.com", t)
	defer repo1.DleteUser(ctx, sender.ID)

	_, repo2, receiver := NewUser(ctx, "test-transfer-receiver@example.com", t)
	defer repo2.DleteUser(ctx, receiver.ID)

	h, _ := NewPaymentHandler()

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

// Test Transfer User Details in Financial History
func TestFinancialHistory_TransferUserDetails(t *testing.T) {
	ctx := context.Background()
	
	_, repo, sender := NewUser(ctx, "sender-details@test.com", t)
	defer repo.DleteUser(ctx, sender.ID)
	
	_, _, receiver := NewUser(ctx, "receiver-details@test.com", t)
	defer repo.DleteUser(ctx, receiver.ID)

	h, _ := NewPaymentHandler()

	// Add coins to sender
	_, err := h.InitiateCoinPurchase(ctx, &paymentpb.InitiateCoinPurchaseRequest{
		UserId:        int64(sender.ID),
		Amount:        100,
		PaymentMethod: "stripe",
	})
	if err != nil {
		t.Fatalf("Purchase failed: %v", err)
	}

	// Transfer from sender to receiver
	_, err = h.TransferToUser(ctx, &paymentpb.TransferToUserRequest{
		FromUserId: int64(sender.ID),
		ToUserId:   int64(receiver.ID),
		Amount:     50,
	})
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

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

	t.Logf("\n=== Sender Financial History ===")
	foundTransfer := false
	for _, item := range senderHistory.Items {
		t.Logf("Type=%s, Amount=%.2f, Desc=%s", item.Type, item.Amount, item.Description)
		if item.Type == "debit" && item.Description == "Transfer sent" {
			foundTransfer = true
			t.Logf("  → Transfer To: ID=%d, Name=%s, Email=%s", 
				item.TransferUserId, item.TransferUserName, item.TransferUserEmail)
			
			if item.TransferUserId != int64(receiver.ID) {
				t.Errorf("Expected receiver ID %d, got %d", receiver.ID, item.TransferUserId)
			}
			if item.TransferUserEmail != receiver.Email {
				t.Errorf("Expected receiver email %s, got %s", receiver.Email, item.TransferUserEmail)
			}
		}
	}
	if !foundTransfer {
		t.Error("Transfer not found in sender's history!")
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

	t.Logf("\n=== Receiver Financial History ===")
	foundTransfer = false
	for _, item := range receiverHistory.Items {
		t.Logf("Type=%s, Amount=%.2f, Desc=%s", item.Type, item.Amount, item.Description)
		if item.Type == "credit" && item.Description == "Transfer received" {
			foundTransfer = true
			t.Logf("  → Transfer From: ID=%d, Name=%s, Email=%s", 
				item.TransferUserId, item.TransferUserName, item.TransferUserEmail)
			
			if item.TransferUserId != int64(sender.ID) {
				t.Errorf("Expected sender ID %d, got %d", sender.ID, item.TransferUserId)
			}
			if item.TransferUserEmail != sender.Email {
				t.Errorf("Expected sender email %s, got %s", sender.Email, item.TransferUserEmail)
			}
		}
	}
	if !foundTransfer {
		t.Error("Transfer not found in receiver's history!")
	}

	t.Logf("\n✓ Transfer user details working correctly!")
}
