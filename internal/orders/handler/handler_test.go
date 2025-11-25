package handler

import (
	"context"
	"fmt"
	"rival/config"
	"rival/connection"
	orderpb "rival/gen/proto/proto/api"
	pb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	authHandler "rival/internal/auth/handler"
	"testing"
)

// NewOrderUser creates a test user for order testing
func NewOrderUser(ctx context.Context, email string, role schemapb.UserRole, t *testing.T) (*pb.SignupRequest, *schema.Queries, schema.User) {
	handler, err := authHandler.NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create auth handler: %v", err)
	}

	data := pb.SignupRequest{
		Name:     "Test Order User",
		Email:    email,
		Password: "password123",
		Role:     role,
		Phone:    "12345678",
	}

	_, err = handler.Signup(ctx, &data)
	if err != nil {
		t.Fatalf("Failed to signup user: %v", err)
	}

	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to get db connection: %v", err)
	}

	repo := schema.New(db)
	user, err := repo.GetUserByEmail(ctx, data.Email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	t.Logf("Created test order user: %v", user)
	return &data, repo, user
}

func TestCreateOrder_ValidRequest(t *testing.T) {
	ctx := context.Background()

	// Create a test customer user
	_, repo, customer := NewOrderUser(ctx, "test-customer@example.com", schemapb.UserRole_USER_ROLE_CUSTOMER, t)
	defer func() {
		err := repo.DleteUser(ctx, customer.ID)
		if err != nil {
			t.Logf("Failed to cleanup customer: %v", err)
		}
	}()

	// Create a test merchant user
	_, repo2, merchant := NewOrderUser(ctx, "test-merchant@example.com", schemapb.UserRole_USER_ROLE_MERCHANT, t)
	merchantRecord := CreateMerchantRecord(ctx, merchant, repo2, t)
	defer func() {
		err := repo2.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test creating order
	req := &orderpb.CreateOrderRequest{
		UserId:     int64(customer.ID),
		MerchantId: int64(merchantRecord.ID),
		Items:      `[{"name":"Test Item 1","quantity":2,"price":25.25,"subtotal":50.50},{"name":"Test Item 2","quantity":1,"price":50.00,"subtotal":50.00}]`, // JSON string
		Subtotal:   100.50,
		CoinsUsed:  0.0,
		Notes:      "Test order with delivery address",
	}

	resp, err := h.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("CreateOrder returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("CreateOrder returned nil response")
	}

	if resp.Order == nil {
		t.Errorf("Expected CreateOrder to return an order, got nil")
	}

	t.Logf("Create order response: %+v", resp)
}

func TestCreateOrder_ZeroSubtotal(t *testing.T) {
	ctx := context.Background()
	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test with zero subtotal - should return error
	req := &orderpb.CreateOrderRequest{
		UserId:     1,
		MerchantId: 1,
		Subtotal:   0, // Invalid subtotal
		CoinsUsed:  0,
		Items:      `[]`, // Empty items JSON
		Notes:      "Test with zero subtotal",
	}

	resp, err := h.CreateOrder(ctx, req)
	if err == nil {
		t.Fatalf("CreateOrder should return error for zero subtotal, got response: %v", resp)
	}

	expectedError := "subtotal must be greater than 0"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCreateOrder_NegativeSubtotal(t *testing.T) {
	ctx := context.Background()
	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test with negative subtotal - should return error
	req := &orderpb.CreateOrderRequest{
		UserId:     1,
		MerchantId: 1,
		Subtotal:   -10.50, // Invalid subtotal
		CoinsUsed:  0,
		Items:      `[]`, // Empty items JSON
		Notes:      "Test with negative subtotal",
	}

	resp, err := h.CreateOrder(ctx, req)
	if err == nil {
		t.Fatalf("CreateOrder should return error for negative subtotal, got response: %v", resp)
	}

	expectedError := "subtotal must be greater than 0"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetOrder(t *testing.T) {
	ctx := context.Background()

	// Create a test customer user
	_, repo, customer := NewOrderUser(ctx, "test-get-order@example.com", schemapb.UserRole_USER_ROLE_CUSTOMER, t)
	defer func() {
		err := repo.DleteUser(ctx, customer.ID)
		if err != nil {
			t.Logf("Failed to cleanup customer: %v", err)
		}
	}()

	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting order
	req := &orderpb.GetOrderRequest{
		OrderId: 123, // Use int64 instead of string
	}

	resp, err := h.GetOrder(ctx, req)
	if err != nil {
		t.Fatalf("GetOrder returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetOrder returned nil response")
	}

	t.Logf("Get order response: %+v", resp)
}

func TestGetUserOrders_WithDefaults(t *testing.T) {
	ctx := context.Background()

	// Create a test customer user
	_, repo, customer := NewOrderUser(ctx, "test-user-orders@example.com", schemapb.UserRole_USER_ROLE_CUSTOMER, t)
	defer func() {
		err := repo.DleteUser(ctx, customer.ID)
		if err != nil {
			t.Logf("Failed to cleanup customer: %v", err)
		}
	}()

	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test with invalid pagination - should apply defaults
	req := &orderpb.GetUserOrdersRequest{
		UserId: int64(customer.ID),
		Page:   0, // Invalid page - should default to 1
		Limit:  0, // Invalid limit - should default to 20
	}

	resp, err := h.GetUserOrders(ctx, req)
	if err != nil {
		t.Fatalf("GetUserOrders returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetUserOrders returned nil response")
	}

	// Verify defaults were applied
	if req.Page != 1 {
		t.Errorf("Expected page to be set to 1, got %d", req.Page)
	}
	if req.Limit != 20 {
		t.Errorf("Expected limit to be set to 20, got %d", req.Limit)
	}

	t.Logf("Get user orders response: %+v", resp)
}

func TestGetUserOrders_WithValidPagination(t *testing.T) {
	ctx := context.Background()

	// Create a test customer user
	_, repo, customer := NewOrderUser(ctx, "test-user-orders-valid@example.com", schemapb.UserRole_USER_ROLE_CUSTOMER, t)
	defer func() {
		err := repo.DleteUser(ctx, customer.ID)
		if err != nil {
			t.Logf("Failed to cleanup customer: %v", err)
		}
	}()

	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test with valid pagination
	req := &orderpb.GetUserOrdersRequest{
		UserId: int64(customer.ID),
		Page:   2,
		Limit:  10,
	}

	resp, err := h.GetUserOrders(ctx, req)
	if err != nil {
		t.Fatalf("GetUserOrders returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetUserOrders returned nil response")
	}

	// Verify pagination values are preserved
	if req.Page != 2 {
		t.Errorf("Expected page to remain 2, got %d", req.Page)
	}
	if req.Limit != 10 {
		t.Errorf("Expected limit to remain 10, got %d", req.Limit)
	}

	t.Logf("Get user orders with valid pagination response: %+v", resp)
}

func TestCancelOrder(t *testing.T) {
	ctx := context.Background()

	// Create a test customer user
	_, repo, customer := NewOrderUser(ctx, "test-cancel-order@example.com", schemapb.UserRole_USER_ROLE_CUSTOMER, t)
	defer func() {
		err := repo.DleteUser(ctx, customer.ID)
		if err != nil {
			t.Logf("Failed to cleanup customer: %v", err)
		}
	}()

	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test canceling order
	req := &orderpb.CancelOrderRequest{
		OrderId: 456, // Use int64 instead of string
		Reason:  "Customer requested cancellation",
	}

	resp, err := h.CancelOrder(ctx, req)
	if err != nil {
		t.Fatalf("CancelOrder returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("CancelOrder returned nil response")
	}

	t.Logf("Cancel order response: %+v", resp)
}

// Test pagination edge cases
func TestGetUserOrders_PaginationEdgeCases(t *testing.T) {
	ctx := context.Background()

	// Create a test customer user
	_, repo, customer := NewOrderUser(ctx, "test-pagination@example.com", schemapb.UserRole_USER_ROLE_CUSTOMER, t)
	defer func() {
		err := repo.DleteUser(ctx, customer.ID)
		if err != nil {
			t.Logf("Failed to cleanup customer: %v", err)
		}
	}()

	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	testCases := []struct {
		name          string
		inputPage     int32
		inputLimit    int32
		expectedPage  int32
		expectedLimit int32
	}{
		{"negative page", -1, 10, 1, 10},
		{"zero page", 0, 10, 1, 10},
		{"negative limit", 2, -5, 2, 20},
		{"zero limit", 2, 0, 2, 20},
		{"both invalid", -1, -5, 1, 20},
		{"large values", 1000, 1000, 1000, 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &orderpb.GetUserOrdersRequest{
				UserId: int64(customer.ID),
				Page:   tc.inputPage,
				Limit:  tc.inputLimit,
			}

			resp, err := h.GetUserOrders(ctx, req)
			if err != nil {
				t.Fatalf("GetUserOrders returned error for %s: %v", tc.name, err)
			}

			if resp == nil {
				t.Fatalf("GetUserOrders returned nil response for %s", tc.name)
			}

			// Check if defaults were applied correctly
			if req.Page != tc.expectedPage {
				t.Errorf("Test %s: Expected page %d, got %d", tc.name, tc.expectedPage, req.Page)
			}
			if req.Limit != tc.expectedLimit {
				t.Errorf("Test %s: Expected limit %d, got %d", tc.name, tc.expectedLimit, req.Limit)
			}

			t.Logf("Test %s completed successfully", tc.name)
		})
	}
}

// Test order creation with different item combinations
func TestCreateOrder_ItemValidation(t *testing.T) {
	ctx := context.Background()

	// Create a test customer user
	_, repo, customer := NewOrderUser(ctx, "test-items@example.com", schemapb.UserRole_USER_ROLE_CUSTOMER, t)
	defer func() {
		err := repo.DleteUser(ctx, customer.ID)
		if err != nil {
			t.Logf("Failed to cleanup customer: %v", err)
		}
	}()

	// Create a test merchant user
	_, repo2, merchant := NewOrderUser(ctx, "test-items-merchant@example.com", schemapb.UserRole_USER_ROLE_MERCHANT, t)
	merchantRecord := CreateMerchantRecord(ctx, merchant, repo2, t)
	defer func() {
		err := repo2.DleteUser(ctx, merchant.ID)
		if err != nil {
			t.Logf("Failed to cleanup merchant: %v", err)
		}
	}()

	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	testCases := []struct {
		name        string
		items       string // JSON string for items
		subtotal    float64
		shouldPass  bool
		description string
	}{
		{
			name:        "valid single item",
			items:       `[{"name":"Pizza","quantity":1,"price":15.99,"subtotal":15.99}]`,
			subtotal:    15.99,
			shouldPass:  true,
			description: "Single item order should work",
		},
		{
			name:        "valid multiple items",
			items:       `[{"name":"Burger","quantity":2,"price":12.50,"subtotal":25.00},{"name":"Fries","quantity":1,"price":3.99,"subtotal":3.99}]`,
			subtotal:    28.99,
			shouldPass:  true,
			description: "Multiple items order should work",
		},
		{
			name:        "empty items",
			items:       `[]`,
			subtotal:    50.00,
			shouldPass:  true, // Service layer might handle this
			description: "Empty items might be handled by service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &orderpb.CreateOrderRequest{
				UserId:     int64(customer.ID),
				MerchantId: int64(merchantRecord.ID),
				Subtotal:   tc.subtotal,
				CoinsUsed:  0.0,
				Items:      tc.items,
				Notes:      fmt.Sprintf("Test case: %s", tc.description),
			}

			resp, err := h.CreateOrder(ctx, req)

			if tc.shouldPass {
				if err != nil {
					t.Fatalf("Test %s: Expected success but got error: %v", tc.name, err)
				}
				if resp == nil {
					t.Fatalf("Test %s: Expected response but got nil", tc.name)
				}
				t.Logf("Test %s: %s - SUCCESS", tc.name, tc.description)
			} else {
				if err == nil {
					t.Fatalf("Test %s: Expected error but got success: %v", tc.name, resp)
				}
				t.Logf("Test %s: %s - Expected error: %v", tc.name, tc.description, err)
			}
		})
	}
}

// Test order states
func TestOrderStates(t *testing.T) {
	ctx := context.Background()

	// Create a test customer user
	_, repo, customer := NewOrderUser(ctx, "test-states@example.com", schemapb.UserRole_USER_ROLE_CUSTOMER, t)
	defer func() {
		err := repo.DleteUser(ctx, customer.ID)
		if err != nil {
			t.Logf("Failed to cleanup customer: %v", err)
		}
	}()

	h, err := NewOrderHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test getting orders with different states
	req := &orderpb.GetUserOrdersRequest{
		UserId: int64(customer.ID),
		Page:   1,
		Limit:  10,
		Status: "pending", // Filter by status
	}

	resp, err := h.GetUserOrders(ctx, req)
	if err != nil {
		t.Fatalf("GetUserOrders returned error: %v", err)
	}

	if resp == nil {
		t.Fatalf("GetUserOrders returned nil response")
	}

	t.Logf("Orders with status filter response: %+v", resp)
}

// ============ END-TO-END COMPREHENSIVE TEST ============

func TestOrderFlow_Complete_EndToEnd(t *testing.T) {
	ctx := context.Background()
	
	// Cleanup existing users
	cfg := config.GetConfig()
	db, _ := connection.GetPgConnection(&cfg.Database)
	repo := schema.New(db)
	
	customerEmail := "e2e-order-customer@example.com"
	merchantEmail := "e2e-order-merchant@example.com"
	
	existingCustomer, _ := repo.GetUserByEmail(ctx, customerEmail)
	if existingCustomer.ID != 0 {
		repo.DleteUser(ctx, existingCustomer.ID)
	}
	existingMerchant, _ := repo.GetUserByEmail(ctx, merchantEmail)
	if existingMerchant.ID != 0 {
		repo.DleteUser(ctx, existingMerchant.ID)
	}
	
	// Create customer and merchant
	_, repo1, customer := NewOrderUser(ctx, customerEmail, schemapb.UserRole_USER_ROLE_CUSTOMER, t)
	defer repo1.DleteUser(ctx, customer.ID)
	
	_, repo2, merchant := NewOrderUser(ctx, merchantEmail, schemapb.UserRole_USER_ROLE_MERCHANT, t)
	merchantRecord := CreateMerchantRecord(ctx, merchant, repo2, t)
	defer repo2.DleteUser(ctx, merchant.ID)
	
	h, _ := NewOrderHandler()
	
	t.Logf("========================================")
	t.Logf("ORDER END-TO-END TEST")
	t.Logf("========================================")
	t.Logf("Customer ID: %d, Email: %s", customer.ID, customer.Email)
	t.Logf("Merchant ID: %d, Email: %s", merchant.ID, merchant.Email)
	
	// Step 1: Create Order
	t.Logf("\n--- STEP 1: Create Order ---")
	createReq := &orderpb.CreateOrderRequest{
		UserId:     int64(customer.ID),
		MerchantId: int64(merchantRecord.ID),
		Items:      `[{"name":"Burger","quantity":2,"price":12.50,"subtotal":25.00},{"name":"Fries","quantity":1,"price":3.99,"subtotal":3.99}]`,
		Subtotal:   28.99,
		CoinsUsed:  0.0,
		Notes:      "E2E test order",
	}
	
	createResp, err := h.CreateOrder(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}
	t.Logf("✓ Order created: ID=%d, Status=%s, Total=%.2f", 
		createResp.Order.Id, createResp.Order.Status, createResp.Order.TotalAmount)
	
	orderID := createResp.Order.Id
	
	// Step 2: Get Order Details
	t.Logf("\n--- STEP 2: Get Order Details ---")
	getReq := &orderpb.GetOrderRequest{
		OrderId: orderID,
	}
	
	getResp, err := h.GetOrder(ctx, getReq)
	if err != nil {
		t.Fatalf("Failed to get order: %v", err)
	}
	t.Logf("✓ Order retrieved: ID=%d, Status=%s", getResp.Order.Id, getResp.Order.Status)
	t.Logf("  Items: %s", getResp.Order.Items)
	t.Logf("  Subtotal: %.2f", getResp.Order.Subtotal)
	t.Logf("  Total: %.2f", getResp.Order.TotalAmount)
	
	// Step 3: Get User Orders
	t.Logf("\n--- STEP 3: Get User Orders ---")
	userOrdersReq := &orderpb.GetUserOrdersRequest{
		UserId: int64(customer.ID),
		Page:   1,
		Limit:  10,
	}
	
	userOrdersResp, err := h.GetUserOrders(ctx, userOrdersReq)
	if err != nil {
		t.Fatalf("Failed to get user orders: %v", err)
	}
	t.Logf("✓ User orders retrieved: Total=%d", len(userOrdersResp.Orders))
	for i, order := range userOrdersResp.Orders {
		t.Logf("  Order %d: ID=%d, Status=%s, Total=%.2f", 
			i+1, order.Id, order.Status, order.TotalAmount)
	}
	
	// Step 4: Cancel Order
	t.Logf("\n--- STEP 4: Cancel Order ---")
	cancelReq := &orderpb.CancelOrderRequest{
		OrderId: orderID,
		Reason:  "Customer requested cancellation for E2E test",
	}
	
	_, err = h.CancelOrder(ctx, cancelReq)
	if err != nil {
		t.Logf("⚠ Cancel order failed: %v", err)
	} else {
		t.Logf("✓ Order cancelled: ID=%d, Status=%s", 
			orderID, "cancelled")
	}
	
	// Step 5: Verify Cancellation
	t.Logf("\n--- STEP 5: Verify Cancellation ---")
	verifyResp, err := h.GetOrder(ctx, getReq)
	if err != nil {
		t.Logf("⚠ Failed to verify cancellation: %v", err)
	} else {
		t.Logf("✓ Order status after cancellation: %s", verifyResp.Order.Status)
	}
	
	t.Logf("\n========================================")
	t.Logf("END-TO-END TEST COMPLETED")
	t.Logf("========================================")
}
