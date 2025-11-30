package handler

import (
	"context"
	"math/big"
	schema "rival/gen/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

// CreateMerchantRecord creates merchant record in merchants table
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

// CleanupMerchant deletes merchant record by email
func CleanupMerchant(ctx context.Context, email string, repo *schema.Queries, t *testing.T) {
	merchant, err := repo.GetMerchantByEmail(ctx, email)
	if err != nil {
		t.Logf("Merchant not found for cleanup: %v", err)
		return
	}
	// Delete merchant record (offers, addresses will cascade)
	err = repo.DeleteMerchant(ctx, merchant.ID)
	if err != nil {
		t.Logf("Failed to cleanup merchant: %v", err)
	}
}
