package handler

import (
	"context"
	"testing"

	"rival/config"
	"rival/connection"
	authpb "rival/gen/proto/proto/api"
	paymentpb "rival/gen/proto/proto/api"
	schema "rival/gen/sql"
	schemapb "rival/gen/proto/proto/schema"
	paymentHandler "rival/internal/payments/handler"
)

func TestSignupWithInitialCoins(t *testing.T) {
	ctx := context.Background()

	// Setup
	email := "test-signup-coins@example.com"
	cfg := config.GetConfig()
	db, _ := connection.GetPgConnection(&cfg.Database)
	queries := schema.New(db)

	// Cleanup first
	existingUser, _ := queries.GetUserByEmail(ctx, email)
	if existingUser.ID != 0 {
		queries.DleteUser(ctx, existingUser.ID)
		t.Logf("Cleaned up existing user")
	}

	// Create auth handler
	authH, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create auth handler: %v", err)
	}

	// Signup
	signupReq := &authpb.SignupRequest{
		Name:     "Test Signup Coins",
		Email:    email,
		Password: "password123",
		Role:     *schemapb.UserRole_USER_ROLE_CUSTOMER.Enum(),
		Phone:    "1234567890",
	}

	signupResp, err := authH.Signup(ctx, signupReq)
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}
	t.Logf("Signup response: %s", signupResp.Message)

	// Get user to get ID
	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	t.Logf("User created with ID: %d", user.ID)

	// Create payment handler
	paymentH, err := paymentHandler.NewPaymentHandler()
	if err != nil {
		t.Fatalf("Failed to create payment handler: %v", err)
	}

	// Check balance
	balanceResp, err := paymentH.GetBalance(ctx, &paymentpb.GetBalanceRequest{
		UserId: user.ID,
	})
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}
	t.Logf("Initial balance: %.2f", balanceResp.Balance)

	if balanceResp.Balance != 10.0 {
		t.Fatalf("Expected initial balance 10.0, got %.2f", balanceResp.Balance)
	}
	t.Logf("✓ Initial balance verified: 10.0 coins")

	// Check payment history
	historyResp, err := paymentH.GetPaymentHistory(ctx, &paymentpb.GetPaymentHistoryRequest{
		UserId: user.ID,
		Page:   1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Failed to get payment history: %v", err)
	}

	t.Logf("Payment history: %d purchases", len(historyResp.Purchases))

	if len(historyResp.Purchases) == 0 {
		t.Fatalf("Expected 1 purchase (signup bonus) in history, got 0")
	}

	signupPurchase := historyResp.Purchases[0]
	t.Logf("Signup bonus purchase: Amount=%.2f, Coins=%.2f, Method=%s, Status=%s",
		signupPurchase.Amount, signupPurchase.CoinsReceived, signupPurchase.PaymentMethod, signupPurchase.Status)

	if signupPurchase.PaymentMethod != "signup_bonus" {
		t.Fatalf("Expected payment method 'signup_bonus', got '%s'", signupPurchase.PaymentMethod)
	}

	if signupPurchase.Status != "completed" {
		t.Fatalf("Expected status 'completed', got '%s'", signupPurchase.Status)
	}

	if signupPurchase.Amount != 10.0 {
		t.Fatalf("Expected amount 10.0, got %.2f", signupPurchase.Amount)
	}

	t.Logf("✓ Signup bonus purchase record verified")
	t.Logf("✓✓✓ All checks passed! User gets 10 coins on signup with proper history.")

	// Cleanup
	queries.DleteUser(ctx, user.ID)
}
