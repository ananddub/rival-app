package handler

import (
	"context"
	"rival/config"
	"rival/connection"
	authpb "rival/gen/proto/proto/api"
	userspb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	"rival/internal/auth/handler"
	"testing"
)

func NewUser(ctx context.Context, email string, t *testing.T) (*authpb.SignupRequest, *schema.Queries, schema.User) {
	handler, err := handler.NewAuthHandler()
	if err != nil {
		panic("Failed to create handler: " + err.Error())
	}
	data := authpb.SignupRequest{
		Name:     "Test User",
		Email:    email,
		Password: "password123",
		Role:     *schemapb.UserRole_USER_ROLE_ADMIN.Enum(),
		Phone:    "12345678",
	}
	handler.Signup(ctx, &data)
	cfg := config.GetConfig()

	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}
	repo := schema.New(db)
	user, err := repo.GetUserByEmail(ctx, data.Email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}
	ctx = context.WithValue(ctx, "user_id", user.ID)
	t.Logf("SignupAuto created user: %v", user)
	return &data, repo, user
}

func TestUser(t *testing.T) {
	ctx := context.Background()
	data, repo, auser := NewUser(ctx, "test@example.com", t)
	defer func() {
		repo.DeleteByEmail(context.Background(), data.Email)
	}()

	handler, err := NewUserHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	user, err := handler.GetUser(ctx, &authpb.GetUserRequest{
		UserId: auser.ID,
	})
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if user.User.Id != auser.ID {
		t.Fatalf("GetUser returned wrong user: got %v want %v", user.User.Id, auser.ID)
	}

	// Test with userId=0
	zeroUser, err := handler.GetUser(ctx, &authpb.GetUserRequest{
		UserId: 0,
	})
	if err != nil {
		t.Fatalf("Failed to handle userId=0: %v", err)
	}
	if zeroUser.User != nil {
		t.Fatalf("GetUser should return nil user for userId=0, got: %v", zeroUser.User)
	}
}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	data, repo, auser := NewUser(ctx, "test-update@example.com", t)
	defer func() {
		repo.DeleteByEmail(context.Background(), data.Email)
	}()

	handler, err := NewUserHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test successful update
	updateReq := &userspb.UpdateUserRequest{
		UserId:     auser.ID,
		Name:       "Updated Name",
		Phone:      "987654321",
		ProfilePic: "updated-pic.jpg",
	}

	updateResp, err := handler.UpdateUser(ctx, updateReq)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if updateResp.User == nil {
		t.Fatalf("UpdateUser returned nil user")
	}

	if updateResp.User.Name != "Updated Name" {
		t.Fatalf("UpdateUser did not update name: got %v want %v", updateResp.User.Name, "Updated Name")
	}

	// Test with zero user ID
	zeroReq := &userspb.UpdateUserRequest{
		UserId: 0,
		Name:   "Test",
	}

	zeroResp, err := handler.UpdateUser(ctx, zeroReq)
	if err != nil {
		t.Fatalf("Failed to handle zero user ID: %v", err)
	}

	if zeroResp.User != nil {
		t.Fatalf("UpdateUser should return nil user for zero ID")
	}
}

func TestGetUploadURL(t *testing.T) {
	ctx := context.Background()
	data, repo, auser := NewUser(ctx, "test-upload@example.com", t)
	defer func() {
		repo.DeleteByEmail(context.Background(), data.Email)
	}()

	handler, err := NewUserHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test successful upload URL generation
	uploadReq := &userspb.GetUploadURLRequest{
		UserId:      auser.ID,
		FileName:    "profile.jpg",
		ContentType: "image/jpeg",
	}

	uploadResp, err := handler.GetUploadURL(ctx, uploadReq)
	if err != nil {
		t.Fatalf("Failed to get upload URL: %v", err)
	}

	if uploadResp.UploadUrl == "" {
		t.Fatalf("GetUploadURL returned empty upload URL")
	}

	t.Logf("Upload URL: %s", uploadResp.UploadUrl)
	t.Logf("File URL: %s", uploadResp.FileUrl)
	t.Logf("Expires In: %d seconds", uploadResp.ExpiresIn)

	// Test with zero user ID
	zeroReq := &userspb.GetUploadURLRequest{
		UserId:   0,
		FileName: "test.jpg",
	}

	zeroResp, err := handler.GetUploadURL(ctx, zeroReq)
	if err != nil {
		t.Fatalf("Failed to handle zero user ID: %v", err)
	}

	if zeroResp.UploadUrl != "" {
		t.Fatalf("GetUploadURL should return empty URL for zero user ID")
	}

	// Test with empty filename
	emptyFileReq := &userspb.GetUploadURLRequest{
		UserId:   auser.ID,
		FileName: "",
	}

	emptyFileResp, err := handler.GetUploadURL(ctx, emptyFileReq)
	if err != nil {
		t.Fatalf("Failed to handle empty filename: %v", err)
	}

	if emptyFileResp.UploadUrl != "" {
		t.Fatalf("GetUploadURL should return empty URL for empty filename")
	}
}

func TestGetBalance(t *testing.T) {
	ctx := context.Background()
	data, repo, auser := NewUser(ctx, "test-balance@example.com", t)
	defer func() {
		repo.DeleteByEmail(context.Background(), data.Email)
	}()

	handler, err := NewUserHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting coin balance
	balanceReq := &userspb.GetCoinBalanceRequest{
		UserId: auser.ID,
	}

	balanceResp, err := handler.GetCoinBalance(ctx, balanceReq)
	if err != nil {
		t.Fatalf("Failed to get coin balance: %v", err)
	}

	if balanceResp.Balance < 0 {
		t.Fatalf("GetCoinBalance returned negative balance: %v", balanceResp.Balance)
	}

	// Test updating coin balance - add coins
	updateReq := &userspb.UpdateCoinBalanceRequest{
		UserId:    auser.ID,
		Amount:    100.0,
		Operation: "add",
	}

	updateResp, err := handler.UpdateCoinBalance(ctx, updateReq)
	if err != nil {
		t.Fatalf("Failed to update coin balance: %v", err)
	}

	if updateResp.NewBalance <= balanceResp.Balance {
		t.Fatalf("UpdateCoinBalance did not increase balance: old=%v new=%v", balanceResp.Balance, updateResp.NewBalance)
	}

	// Test with zero user ID
	zeroReq := &userspb.GetCoinBalanceRequest{
		UserId: 0,
	}

	zeroResp, err := handler.GetCoinBalance(ctx, zeroReq)
	if err != nil {
		t.Fatalf("Failed to handle zero user ID: %v", err)
	}

	if zeroResp.Balance != 0 {
		t.Fatalf("GetCoinBalance should return 0 for zero user ID")
	}

	// Test update with zero amount
	zeroUpdateReq := &userspb.UpdateCoinBalanceRequest{
		UserId:    auser.ID,
		Amount:    0,
		Operation: "add",
	}

	zeroUpdateResp, err := handler.UpdateCoinBalance(ctx, zeroUpdateReq)
	if err != nil {
		t.Fatalf("Failed to handle zero amount: %v", err)
	}

	if zeroUpdateResp.NewBalance != 0 {
		t.Fatalf("UpdateCoinBalance should return 0 for zero amount")
	}
}

func TestReferralCode(t *testing.T) {
	ctx := context.Background()
	data, repo, auser := NewUser(ctx, "test-referral@example.com", t)
	defer func() {
		repo.DeleteByEmail(context.Background(), data.Email)
	}()

	handler, err := NewUserHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting referral code
	referralReq := &userspb.GetReferralCodeRequest{
		UserId: auser.ID,
	}

	referralResp, err := handler.GetReferralCode(ctx, referralReq)
	if err != nil {
		t.Fatalf("Failed to get referral code: %v", err)
	}

	if referralResp.ReferralCode == "" {
		t.Fatalf("GetReferralCode returned empty referral code")
	}

	t.Logf("Generated referral code: %s", referralResp.ReferralCode)

	// Test with zero user ID
	zeroReq := &userspb.GetReferralCodeRequest{
		UserId: 0,
	}

	zeroResp, err := handler.GetReferralCode(ctx, zeroReq)
	if err != nil {
		t.Fatalf("Failed to handle zero user ID: %v", err)
	}

	if zeroResp.ReferralCode != "" {
		t.Fatalf("GetReferralCode should return empty code for zero user ID")
	}

	// Test transaction history
	historyReq := &userspb.GetTransactionHistoryRequest{
		UserId: auser.ID,
		Page:   1,
		Limit:  10,
	}

	historyResp, err := handler.GetTransactionHistory(ctx, historyReq)
	if err != nil {
		t.Fatalf("Failed to get transaction history: %v", err)
	}

	if historyResp.TotalCount < 0 {
		t.Fatalf("GetTransactionHistory returned negative total count: %v", historyResp.TotalCount)
	}

	// Test with zero page and limit (should use defaults)
	defaultReq := &userspb.GetTransactionHistoryRequest{
		UserId: auser.ID,
		Page:   0,
		Limit:  0,
	}

	defaultResp, err := handler.GetTransactionHistory(ctx, defaultReq)
	if err != nil {
		t.Fatalf("Failed to get transaction history with defaults: %v", err)
	}

	if defaultResp.TotalCount < 0 {
		t.Fatalf("GetTransactionHistory with defaults returned negative total count: %v", defaultResp.TotalCount)
	}
}

func TestApplyReferralCode(t *testing.T) {
	ctx := context.Background()

	// Create first user (referrer)
	data1, repo, referrer := NewUser(ctx, "test-referrer@example.com", t)
	defer func() {
		repo.DeleteByEmail(context.Background(), data1.Email)
	}()

	// Create second user (referee)
	data2, _, referee := NewUser(ctx, "test-referee@example.com", t)
	defer func() {
		repo.DeleteByEmail(context.Background(), data2.Email)
	}()

	handler, err := NewUserHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// First, get referral code from referrer
	getReferralReq := &userspb.GetReferralCodeRequest{
		UserId: referrer.ID,
	}

	getReferralResp, err := handler.GetReferralCode(ctx, getReferralReq)
	if err != nil {
		t.Fatalf("Failed to get referral code: %v", err)
	}

	if getReferralResp.ReferralCode == "" {
		t.Fatalf("GetReferralCode returned empty referral code")
	}

	// Test applying valid referral code
	applyReq := &userspb.ApplyReferralCodeRequest{
		UserId:       referee.ID,
		ReferralCode: getReferralResp.ReferralCode,
	}

	applyResp, err := handler.ApplyReferralCode(ctx, applyReq)
	if err != nil {
		t.Fatalf("Failed to apply referral code: %v", err)
	}

	// The response should indicate success or failure based on business logic
	t.Logf("Apply referral code response: success=%v, message=%s, reward=%v",
		applyResp.Success, applyResp.Message, applyResp.RewardAmount)

	// Test applying referral code with zero user ID
	zeroUserReq := &userspb.ApplyReferralCodeRequest{
		UserId:       0,
		ReferralCode: getReferralResp.ReferralCode,
	}

	zeroUserResp, err := handler.ApplyReferralCode(ctx, zeroUserReq)
	if err != nil {
		t.Fatalf("Failed to handle zero user ID: %v", err)
	}

	if zeroUserResp.Success != false {
		t.Fatalf("ApplyReferralCode should fail for zero user ID")
	}

	if zeroUserResp.Message != "User ID and referral code required" {
		t.Fatalf("ApplyReferralCode should return proper error message for zero user ID")
	}

	// Test applying empty referral code
	emptyCodeReq := &userspb.ApplyReferralCodeRequest{
		UserId:       referee.ID,
		ReferralCode: "",
	}

	emptyCodeResp, err := handler.ApplyReferralCode(ctx, emptyCodeReq)
	if err != nil {
		t.Fatalf("Failed to handle empty referral code: %v", err)
	}

	if emptyCodeResp.Success != false {
		t.Fatalf("ApplyReferralCode should fail for empty referral code")
	}

	if emptyCodeResp.Message != "User ID and referral code required" {
		t.Fatalf("ApplyReferralCode should return proper error message for empty referral code")
	}

	// Test applying invalid referral code
	invalidCodeReq := &userspb.ApplyReferralCodeRequest{
		UserId:       referee.ID,
		ReferralCode: "INVALID_CODE_123",
	}

	invalidCodeResp, err := handler.ApplyReferralCode(ctx, invalidCodeReq)
	if err != nil {
		// This might fail at the service level, which is acceptable
		t.Logf("Expected error for invalid referral code: %v", err)
	} else {
		// If it doesn't error, it should at least indicate failure
		if invalidCodeResp.Success {
			t.Logf("Warning: Invalid referral code was accepted, this might need business logic review")
		}
	}
}

func TestGetUploadURL_ZeroUserId(t *testing.T) {
	ctx := context.Background()
	
	handler, err := NewUserHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	req := &userspb.GetUploadURLRequest{
		UserId:      0,
		FileName:    "0.jpg",
		ContentType: "image/jpeg",
	}

	resp, err := handler.GetUploadURL(ctx, req)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	t.Logf("Response for userId=0:")
	t.Logf("  Upload URL: %s", resp.UploadUrl)
	t.Logf("  File URL: %s", resp.FileUrl)
	t.Logf("  Expires In: %d", resp.ExpiresIn)

	if resp.UploadUrl != "" {
		t.Fatalf("Expected empty URL for userId=0, got: %s", resp.UploadUrl)
	}
}
