package handler

import (
	"context"
	"rival/config"
	"rival/connection"
	merchantpb "rival/gen/proto/proto/api"
	pb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	authHandler "rival/internal/auth/handler"
	"testing"
)

// NewMerchantUser creates a test merchant user for testing
// NewMerchantUser creates a test merchant user for testing
func NewMerchantUser(ctx context.Context, email string, t *testing.T) (*pb.SignupRequest, *schema.Queries, schema.User) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to get db connection: %v", err)
	}
	repo := schema.New(db)
	
	// Delete existing user if exists
	existingUser, _ := repo.GetUserByEmail(ctx, email)
	if existingUser.ID != 0 {
		// Delete merchant record first if exists
		existingMerchant, _ := repo.GetMerchantByEmail(ctx, email)
		if existingMerchant.ID != 0 {
			repo.DeleteMerchant(ctx, existingMerchant.ID)
			t.Logf("Deleted existing merchant: %s", email)
		}
		repo.DleteUser(ctx, existingUser.ID)
		t.Logf("Deleted existing user: %s", email)
	}
	
	handler, err := authHandler.NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create auth handler: %v", err)
	}

	data := pb.SignupRequest{
		Name:     "Test Merchant User",
		Email:    email,
		Password: "password123",
		Role:     *schemapb.UserRole_USER_ROLE_MERCHANT.Enum(),
		Phone:    "12345678",
	}

	_, err = handler.Signup(ctx, &data)
	if err != nil {
		t.Fatalf("Failed to signup merchant user: %v", err)
	}

	user, err := repo.GetUserByEmail(ctx, data.Email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	ctx = context.WithValue(ctx, "user_id", user.ID)
	t.Logf("Created test merchant user: %v", user)
	return &data, repo, user
}

// NewCustomerUser creates a test customer user for testing
func NewCustomerUser(ctx context.Context, email string, t *testing.T) (*pb.SignupRequest, *schema.Queries, schema.User) {
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
		t.Logf("Deleted existing customer: %s", email)
	}
	
	handler, err := authHandler.NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create auth handler: %v", err)
	}

	data := pb.SignupRequest{
		Name:     "Test Customer User",
		Email:    email,
		Password: "password123",
		Role:     *schemapb.UserRole_USER_ROLE_CUSTOMER.Enum(),
		Phone:    "87654321",
	}

	_, err = handler.Signup(ctx, &data)
	if err != nil {
		t.Fatalf("Failed to signup customer user: %v", err)
	}

	user, err := repo.GetUserByEmail(ctx, data.Email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	ctx = context.WithValue(ctx, "user_id", user.ID)
	t.Logf("Created test customer user: %v", user)
	return &data, repo, user
}

func TestGetMerchant(t *testing.T) {
	ctx := context.Background()

	// Create a test merchant user
	_, repo, merchantUser := NewMerchantUser(ctx, "test-merchant@example.com", t)
	merchant := CreateMerchantRecord(ctx, merchantUser, repo, t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		err := repo.DleteUser(ctx, merchantUser.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting merchant
	req := &pb.GetMerchantRequest{
		MerchantId: int64(merchant.ID),
	}

	resp, err := h.GetMerchant(ctx, req)
	if err != nil {
		t.Fatalf("GetMerchant returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetMerchant returned nil response")
	}

	t.Logf("Get merchant response: %+v", resp)
}

func TestUpdateMerchant(t *testing.T) {
	ctx := context.Background()

	// Create a test merchant user
	_, repo, merchantUser := NewMerchantUser(ctx, "test-update-merchant@example.com", t)
	merchant := CreateMerchantRecord(ctx, merchantUser, repo, t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		err := repo.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test updating merchant
	req := &pb.UpdateMerchantRequest{
		MerchantId:         int64(merchant.ID),
		Name:               "Updated Business Name",
		Phone:              "9876543210",
		Category:           "Updated Category",
		DiscountPercentage: 5.0,
	}

	resp, err := h.UpdateMerchant(ctx, req)
	if err != nil {
		t.Fatalf("UpdateMerchant returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("UpdateMerchant returned nil response")
	}

	t.Logf("Update merchant response: %+v", resp)
}

func TestGetMerchantAddress(t *testing.T) {
	ctx := context.Background()

	// Create a test merchant user
	_, repo, merchant := NewMerchantUser(ctx, "test-address-merchant@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		err := repo.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting merchant address
	req := &merchantpb.GetMerchantAddressRequest{
		MerchantId: int64(merchant.ID),
	}

	resp, err := h.GetMerchantAddress(ctx, req)
	if err != nil {
		t.Fatalf("GetMerchantAddress returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetMerchantAddress returned nil response")
	}

	t.Logf("Get merchant address response: %+v", resp)
}

func TestUpdateMerchantAddress(t *testing.T) {
	ctx := context.Background()

	// Create a test merchant user
	_, repo, merchant := NewMerchantUser(ctx, "test-update-address-merchant@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		err := repo.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test updating merchant address
	req := &merchantpb.UpdateMerchantAddressRequest{
		MerchantId: int64(merchant.ID),
		Street:     "123 Test Street",
		City:       "Test City",
		State:      "Test State",
		PostalCode: "12345",
		Country:    "Test Country",
		Latitude:   40.7128,
		Longitude:  -74.0060,
	}

	resp, err := h.UpdateMerchantAddress(ctx, req)
	if err != nil {
		t.Fatalf("UpdateMerchantAddress returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("UpdateMerchantAddress returned nil response")
	}

	if resp.Address == nil {
		t.Fatalf("UpdateMerchantAddress returned nil address")
	}

	// Verify address fields
	if resp.Address.Street != req.Street {
		t.Errorf("Expected street %s, got %s", req.Street, resp.Address.Street)
	}
	if resp.Address.City != req.City {
		t.Errorf("Expected city %s, got %s", req.City, resp.Address.City)
	}

	t.Logf("Update merchant address response: %+v", resp)
}

func TestGetOrders(t *testing.T) {
	ctx := context.Background()

	// Create a test merchant user
	_, repo, merchant := NewMerchantUser(ctx, "test-orders-merchant@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		err := repo.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting orders
	req := &merchantpb.GetOrdersRequest{
		MerchantId: int64(merchant.ID),
		Page:       1,
		Limit:      10,
	}

	resp, err := h.GetOrders(ctx, req)
	if err != nil {
		t.Fatalf("GetOrders returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetOrders returned nil response")
	}

	t.Logf("Get orders response: %+v", resp)
}

func TestUpdateOrderStatus(t *testing.T) {
	ctx := context.Background()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test updating order status
	req := &merchantpb.UpdateOrderStatusRequest{
		OrderId: 123,         // Use int64 instead of string
		Status:  "confirmed", // Use string status instead of enum
		Notes:   "Order confirmed by merchant",
	}

	resp, err := h.UpdateOrderStatus(ctx, req)
	if err != nil {
		t.Fatalf("UpdateOrderStatus returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("UpdateOrderStatus returned nil response")
	}

	if resp.Order == nil {
		t.Fatalf("UpdateOrderStatus returned nil order")
	}

	// Verify order status update
	if resp.Order.Status != req.Status {
		t.Errorf("Expected status %v, got %v", req.Status, resp.Order.Status)
	}
	if resp.Order.Id != req.OrderId {
		t.Errorf("Expected order ID %d, got %d", req.OrderId, resp.Order.Id)
	}

	t.Logf("Update order status response: %+v", resp)
}

func TestGetCustomers(t *testing.T) {
	ctx := context.Background()

	// Create a test merchant user
	_, repo, merchant := NewMerchantUser(ctx, "test-customers-merchant@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		err := repo.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting customers
	req := &merchantpb.GetCustomersRequest{
		MerchantId: int64(merchant.ID),
		Page:       1,
		Limit:      10,
	}

	resp, err := h.GetCustomers(ctx, req)
	if err != nil {
		t.Fatalf("GetCustomers returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetCustomers returned nil response")
	}

	t.Logf("Get customers response: %+v", resp)
}

func TestGetPayouts(t *testing.T) {
	ctx := context.Background()

	// Create a test merchant user
	_, repo, merchant := NewMerchantUser(ctx, "test-payouts-merchant@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		err := repo.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting payouts
	req := &merchantpb.GetPayoutsRequest{
		MerchantId: int64(merchant.ID),
		Page:       1,
		Limit:      10,
	}

	resp, err := h.GetPayouts(ctx, req)
	if err != nil {
		t.Fatalf("GetPayouts returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetPayouts returned nil response")
	}

	t.Logf("Get payouts response: %+v", resp)
}

func TestCreateOffer(t *testing.T) {
	ctx := context.Background()

	// Create a test merchant user
	_, repo, merchantUser := NewMerchantUser(ctx, "test-offer-merchant@example.com", t)
	merchant := CreateMerchantRecord(ctx, merchantUser, repo, t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		err := repo.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test creating offer
	req := &merchantpb.CreateOfferRequest{
		MerchantId:         int64(merchant.ID),
		Title:              "Test Offer",
		Description:        "This is a test offer",
		DiscountPercentage: 20.0,
		MinAmount:          100.0,
		MaxDiscount:        50.0,
		ValidUntil:         1800000000,
	}

	resp, err := h.CreateOffer(ctx, req)
	if err != nil {
		t.Fatalf("CreateOffer returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("CreateOffer returned nil response")
	}

	t.Logf("Create offer response: %+v", resp)
}

func TestGetOffers(t *testing.T) {
	ctx := context.Background()

	// Create a test merchant user
	_, repo, merchant := NewMerchantUser(ctx, "test-get-offers-merchant@example.com", t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		err := repo.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewMerchantHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting offers
	req := &merchantpb.GetOffersRequest{
		MerchantId: int64(merchant.ID),
		ActiveOnly: true,
	}

	resp, err := h.GetOffers(ctx, req)
	if err != nil {
		t.Fatalf("GetOffers returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetOffers returned nil response")
	}

	t.Logf("Get offers response: %+v", resp)
	t.Logf("Get offers response: %+v", resp)
}

func TestGetDashboardStats(t *testing.T) {
	ctx := context.Background()
	_, repo, merchantUser := NewMerchantUser(ctx, "test-dashboard-merchant@example.com", t)
	merchant := CreateMerchantRecord(ctx, merchantUser, repo, t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		repo.DleteUser(ctx, merchantUser.ID)
	}()

	h, _ := NewMerchantHandler()
	req := &merchantpb.GetDashboardStatsRequest{
		MerchantId: int64(merchant.ID),
	}

	resp, err := h.GetDashboardStats(ctx, req)
	if err != nil {
		t.Fatalf("GetDashboardStats returned error: %v", err)
	}
	t.Logf("Dashboard stats: %+v", resp)
}

func TestUpdateOffer(t *testing.T) {
	ctx := context.Background()
	_, repo, merchantUser := NewMerchantUser(ctx, "test-update-offer-merchant@example.com", t)
	merchant := CreateMerchantRecord(ctx, merchantUser, repo, t)
	defer func() {
		CleanupMerchant(ctx, merchant.Email, repo, t)
		repo.DleteUser(ctx, merchantUser.ID)
	}()

	h, _ := NewMerchantHandler()

	// Create offer first
	createReq := &merchantpb.CreateOfferRequest{
		MerchantId:         int64(merchant.ID),
		Title:              "Test Offer",
		Description:        "Test description",
		DiscountPercentage: 20.0,
		MinAmount:          10.0,
		MaxDiscount:        50.0,
	}
	
	createResp, err := h.CreateOffer(ctx, createReq)
	if err != nil {
		t.Skip("Cannot test update without creating offer first")
	}

	// Update offer
	updateReq := &merchantpb.UpdateOfferRequest{
		OfferId:            createResp.Offer.Id,
		Title:              "Updated Offer",
		Description:        "Updated description",
		DiscountPercentage: 25.0,
		IsActive:           true,
	}

	resp, err := h.UpdateOffer(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateOffer failed: %v", err)
	}
	
	t.Logf("✓ Offer updated: %+v", resp.Offer)
}
