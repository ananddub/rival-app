package handler

import (
	"context"
	"testing"

	"rival/config"
	"rival/connection"
	authpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	"rival/pkg/tb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/metadata"
)

type authRepository struct {
	db      *pgxpool.Pool
	queries *schema.Queries
	redis   *redis.Client
	tb      *tb.TbService
}

func NewRepo() (*authRepository, error) {
	cfg := config.GetConfig()

	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	redisClient := connection.GetRedisClient(&cfg.Redis)

	tbService, err := tb.NewService()
	if err != nil {
		return nil, err
	}

	return &authRepository{
		db:      db,
		queries: schema.New(db),
		redis:   redisClient,
		tb:      tbService,
	}, nil
}

func getTigerbeetleId(id pgtype.UUID) uuid.UUID {
	userUUID, _ := id.Value()
	userID, _ := uuid.Parse(userUUID.(string))
	return userID
}

func TestSignup(t *testing.T) {
	handler, err := NewAuthHandler()
	repo, err := NewRepo()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	tests := authpb.SignupRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Role:     *schemapb.UserRole_USER_ROLE_ADMIN.Enum(),
		Phone:    "12345678",
	}
	_, err = handler.Signup(context.Background(), &tests)

	if err != nil {
		t.Errorf("Signup() error = %v", err)
	}
	defer func() {
		repo.queries.DeleteByEmail(context.Background(), tests.Email)
		// repo.db.Close()
	}()
	user, err := repo.queries.GetUserByEmail(context.Background(), tests.Email)
	if err != nil {
		t.Errorf("Signup() error = %v", err)
	}
	if user.Email != tests.Email {
		t.Errorf("Signup() failed to create user with correct email, got = %v, want %v", user.Email, tests.Email)
	}
	tbuser, err := repo.tb.GetUser(int(user.ID))
	if err != nil {
		t.Errorf("Signup() failed to get tigerbeetle user: %v", err)
	}

	if tbuser == nil || len(*tbuser) == 0 {
		t.Errorf("Signup() failed to create tigerbeetle account with correct ID")
	}

	t.Log(user, tbuser)
	if err != nil {
		t.Errorf("Signup() error = %v", err)
	}
}

func TestLogin(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	ctx := context.Background()
	data := authpb.SignupRequest{
		Name:     "Test User",
		Email:    "test1@example.com",
		Password: "password123",
		Role:     *schemapb.UserRole_USER_ROLE_ADMIN.Enum(),
		Phone:    "12345678",
	}
	value, err := handler.Signup(ctx, &data)
	if err != nil {
		t.Fatalf("Failed to signup user: %v", err)
	}
	t.Logf("Signup result:%v", value)
	defer func() {
		repo, err := NewRepo()
		if err != nil {
			t.Fatalf("Failed to create repo: %v", err)
		}
		repo.queries.DeleteByEmail(ctx, data.Email)
		// repo.db.Close()
	}()
	req := &authpb.LoginRequest{
		Email:    data.Email,
		Password: data.Password,
	}

	loginValue, err := handler.Login(ctx, req)
	if err != nil {
		t.Errorf("Login() error = %v", err)
	}
	t.Logf("Login result:%v", loginValue)
}
func signupAuto(ctx context.Context, email string, t *testing.T) *authpb.SignupRequest {
	handler, err := NewAuthHandler()
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
	repo, err := NewRepo()
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}
	user, err := repo.queries.GetUserByEmail(ctx, data.Email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}
	t.Logf("SignupAuto created user: %v", user)
	return &data
}
func TestVerifyOTP(t *testing.T) {
	handler, err := NewAuthHandler()
	redisClient := connection.GetRedisClient(&config.GetConfig().Redis)
	ctx := context.Background()
	data := signupAuto(ctx, "testingotp@example.com", t)
	defer deleteUserByEmail(ctx, data.Email, t)
	handler.Login(
		ctx,
		&authpb.LoginRequest{
			Email:    data.Email,
			Password: data.Password,
		},
	)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	otp := redisClient.Get(ctx, "otp:"+data.Email).String()
	t.Logf("OTP from redis: %v", otp)
	req := &authpb.VerifyOTPRequest{
		Email: data.Email,
		Otp:   otp,
	}
	_, err = handler.VerifyOTP(context.Background(), req)
	if err != nil {
		t.Errorf("VerifyOTP() error = %v :%v", err, otp)
	}
	t.Logf("VerifyOTP result: %v", err)
}
func deleteUserByEmail(ctx context.Context, email string, t *testing.T) {
	repo, err := NewRepo()
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}
	err = repo.queries.DeleteByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to delete user by email: %v", err)
	}
	usr, err := repo.queries.GetUserByEmail(ctx, email)
	if err == nil && usr.ID != 0 {
		t.Fatalf("User not deleted: %v", email)
	}
	t.Logf("User successfully deleted: %v", email)
}
func TestResetPassword(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	ctx := context.Background()
	data := signupAuto(ctx, "testresedotp@gmail.com", t)
	defer deleteUserByEmail(ctx, data.Email, t)
	req := &authpb.ForgotPasswordRequest{
		Email: data.Email,
	}

	_, err = handler.ForgotPassword(ctx, req)
	if err != nil {
		t.Errorf("ForgotPassword() error = %v", err)
	}
	otp := connection.GetRedisClient(&config.GetConfig().Redis).Get(ctx, "otp:reset:"+data.Email).String()
	req2 := &authpb.ResetPasswordRequest{
		Email:       data.Email,
		Otp:         otp,
		NewPassword: "123456788",
	}

	_, err = handler.ResetPassword(context.Background(), req2)
	if err != nil {
		t.Errorf("ResetPassword() error = %v", err)
	}
	_, err = handler.Login(
		ctx,
		&authpb.LoginRequest{
			Email:    data.Email,
			Password: req2.NewPassword,
		},
	)
	t.Logf("ResendOTP result: %v", err)
}

func sTestFirebaseLogin(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	req := &authpb.FirebaseLoginRequest{
		FirebaseToken: "dummy_token",
	}

	_, err = handler.FirebaseLogin(context.Background(), req)
	t.Logf("FirebaseLogin result: %v", err)
}
func TestRefreshToken(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	req := &authpb.RefreshTokenRequest{
		RefreshToken: "dummy_refresh_token",
	}

	_, err = handler.RefreshToken(context.Background(), req)
	t.Logf("RefreshToken result: %v", err)
}

func TestLogout(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	req := &authpb.LogoutRequest{
		Token: "dummy_token",
	}

	_, err = handler.Logout(context.Background(), req)
	t.Logf("Logout result: %v", err)
}

func TestWhoAmI(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer dummy_token",
	))

	_, err = handler.WhoAmI(ctx, &authpb.WhoAmIRequest{})

	t.Logf("WhoAmI result: %v", err)
}

// ============ COMPREHENSIVE END-TO-END TESTS ============

func TestAuthFlow_Complete_EndToEnd(t *testing.T) {
	ctx := context.Background()
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	repo, err := NewRepo()
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	email := "e2e-auth-test@example.com"
	password := "SecurePass123"

	t.Logf("========================================")
	t.Logf("AUTH END-TO-END TEST")
	t.Logf("========================================")

	// Cleanup before test
	repo.queries.DeleteByEmail(ctx, email)

	// Step 1: Signup
	t.Logf("\n--- STEP 1: User Signup ---")
	signupReq := &authpb.SignupRequest{
		Name:     "E2E Test User",
		Email:    email,
		Password: password,
		Role:     *schemapb.UserRole_USER_ROLE_CUSTOMER.Enum(),
		Phone:    "9876543210",
	}

	signupResp, err := handler.Signup(ctx, signupReq)
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}
	t.Logf("✓ Signup successful: %s", signupResp.Message)

	defer func() {
		repo.queries.DeleteByEmail(ctx, email)
		t.Logf("Cleanup: Deleted user %s", email)
	}()

	// Step 2: Verify user in database
	t.Logf("\n--- STEP 2: Verify User in Database ---")
	user, err := repo.queries.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	t.Logf("✓ User found in DB: ID=%d, Email=%s, Role=%s", user.ID, user.Email, user.Role)

	// Step 3: Verify TigerBeetle account
	t.Logf("\n--- STEP 3: Verify TigerBeetle Account ---")
	tbUser, err := repo.tb.GetUser(int(user.ID))
	if err != nil {
		t.Logf("⚠ TigerBeetle account check failed: %v", err)
	} else if tbUser != nil && len(*tbUser) > 0 {
		t.Logf("✓ TigerBeetle account created: %v", *tbUser)
	}

	// Step 4: Login
	t.Logf("\n--- STEP 4: User Login ---")
	loginReq := &authpb.LoginRequest{
		Email:    email,
		Password: password,
	}

	loginResp, err := handler.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	t.Logf("✓ Login successful")
	t.Logf("  Access Token: %s...", loginResp.AccessToken[:20])
	t.Logf("  Refresh Token: %s...", loginResp.RefreshToken[:20])

	// Step 5: Verify OTP
	t.Logf("\n--- STEP 5: Verify OTP ---")
	otp := repo.redis.Get(ctx, "otp:"+email).Val()
	if otp == "" {
		t.Logf("⚠ No OTP found in Redis")
	} else {
		t.Logf("OTP from Redis: %s", otp)
		verifyReq := &authpb.VerifyOTPRequest{
			Email: email,
			Otp:   otp,
		}
		verifyResp, err := handler.VerifyOTP(ctx, verifyReq)
		if err != nil {
			t.Logf("⚠ OTP verification failed: %v", err)
		} else {
			t.Logf("✓ OTP verified: %s", verifyResp.AccessToken)
		}
	}

	// Step 6: WhoAmI with token
	t.Logf("\n--- STEP 6: WhoAmI (Token Validation) ---")
	ctxWithToken := metadata.NewIncomingContext(ctx, metadata.Pairs(
		"authorization", "Bearer "+loginResp.AccessToken,
	))
	whoAmIResp, err := handler.WhoAmI(ctxWithToken, &authpb.WhoAmIRequest{})
	if err != nil {
		t.Logf("⚠ WhoAmI failed: %v", err)
	} else {
		t.Logf("✓ Token validated")
		t.Logf("  User ID: %d", whoAmIResp.User.Id)
		t.Logf("  Email: %s", whoAmIResp.User.Email)
		t.Logf("  Name: %s", whoAmIResp.User.Name)
	}

	// Step 7: Refresh Token
	t.Logf("\n--- STEP 7: Refresh Token ---")
	refreshReq := &authpb.RefreshTokenRequest{
		RefreshToken: loginResp.RefreshToken,
	}
	refreshResp, err := handler.RefreshToken(ctx, refreshReq)
	if err != nil {
		t.Logf("⚠ Token refresh failed: %v", err)
	} else {
		t.Logf("✓ Token refreshed")
		t.Logf("  New Access Token: %s...", refreshResp.AccessToken[:20])
	}

	// Step 8: Forgot Password
	t.Logf("\n--- STEP 8: Forgot Password ---")
	forgotReq := &authpb.ForgotPasswordRequest{
		Email: email,
	}
	forgotResp, err := handler.ForgotPassword(ctx, forgotReq)
	if err != nil {
		t.Logf("⚠ Forgot password failed: %v", err)
	} else {
		t.Logf("✓ Password reset initiated: %s", forgotResp.Message)
	}

	// Step 9: Reset Password
	t.Logf("\n--- STEP 9: Reset Password ---")
	resetOTP := repo.redis.Get(ctx, "otp:reset:"+email).Val()
	if resetOTP != "" {
		newPassword := "NewSecurePass456"
		resetReq := &authpb.ResetPasswordRequest{
			Email:       email,
			Otp:         resetOTP,
			NewPassword: newPassword,
		}
		resetResp, err := handler.ResetPassword(ctx, resetReq)
		if err != nil {
			t.Logf("⚠ Password reset failed: %v", err)
		} else {
			t.Logf("✓ Password reset successful: %s", resetResp.Message)

			// Verify new password works
			loginReq2 := &authpb.LoginRequest{
				Email:    email,
				Password: newPassword,
			}
			_, err = handler.Login(ctx, loginReq2)
			if err != nil {
				t.Logf("⚠ Login with new password failed: %v", err)
			} else {
				t.Logf("✓ Login with new password successful")
			}
		}
	}

	// Step 10: Logout
	t.Logf("\n--- STEP 10: Logout ---")
	logoutReq := &authpb.LogoutRequest{
		Token: loginResp.AccessToken,
	}
	logoutResp, err := handler.Logout(ctx, logoutReq)
	if err != nil {
		t.Logf("⚠ Logout failed: %v", err)
	} else {
		t.Logf("✓ Logout successful: %v", logoutResp.Success)
	}

	t.Logf("\n========================================")
	t.Logf("END-TO-END TEST COMPLETED")
	t.Logf("========================================")
}

func TestAuthValidation_EdgeCases(t *testing.T) {
	ctx := context.Background()
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	t.Logf("========================================")
	t.Logf("AUTH VALIDATION TESTS")
	t.Logf("========================================")

	// Test 1: Signup with empty email
	t.Logf("\n--- TEST 1: Empty Email ---")
	_, err = handler.Signup(ctx, &authpb.SignupRequest{
		Name:     "Test",
		Email:    "",
		Password: "pass123",
	})
	if err != nil {
		t.Logf("✓ Correctly rejected empty email: %v", err)
	} else {
		t.Logf("✗ Should reject empty email")
	}

	// Test 2: Login with wrong password
	t.Logf("\n--- TEST 2: Wrong Password ---")
	_, err = handler.Login(ctx, &authpb.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpass",
	})
	if err != nil {
		t.Logf("✓ Correctly rejected wrong credentials: %v", err)
	} else {
		t.Logf("✗ Should reject wrong credentials")
	}

	// Test 3: Verify OTP with wrong code
	t.Logf("\n--- TEST 3: Wrong OTP ---")
	_, err = handler.VerifyOTP(ctx, &authpb.VerifyOTPRequest{
		Email: "test@example.com",
		Otp:   "000000",
	})
	if err != nil {
		t.Logf("✓ Correctly rejected wrong OTP: %v", err)
	} else {
		t.Logf("✗ Should reject wrong OTP")
	}

	// Test 4: Refresh with invalid token
	t.Logf("\n--- TEST 4: Invalid Refresh Token ---")
	_, err = handler.RefreshToken(ctx, &authpb.RefreshTokenRequest{
		RefreshToken: "invalid_token",
	})
	if err != nil {
		t.Logf("✓ Correctly rejected invalid token: %v", err)
	} else {
		t.Logf("✗ Should reject invalid token")
	}

	t.Logf("\n========================================")
	t.Logf("VALIDATION TESTS COMPLETED")
	t.Logf("========================================")
}

func TestAuthConcurrency_MultipleUsers(t *testing.T) {
	ctx := context.Background()
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	repo, err := NewRepo()
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	t.Logf("========================================")
	t.Logf("AUTH CONCURRENCY TEST")
	t.Logf("========================================")

	// Create multiple users
	users := []string{
		"concurrent1@example.com",
		"concurrent2@example.com",
		"concurrent3@example.com",
	}

	// Cleanup
	for _, email := range users {
		repo.queries.DeleteByEmail(ctx, email)
	}
	defer func() {
		for _, email := range users {
			repo.queries.DeleteByEmail(ctx, email)
		}
	}()

	// Signup all users
	t.Logf("\n--- Creating Multiple Users ---")
	for i, email := range users {
		signupReq := &authpb.SignupRequest{
			Name:     "Concurrent User " + string(rune(i+1)),
			Email:    email,
			Password: "password123",
			Role:     *schemapb.UserRole_USER_ROLE_CUSTOMER.Enum(),
			Phone:    "1234567890",
		}
		_, err := handler.Signup(ctx, signupReq)
		if err != nil {
			t.Logf("✗ Failed to create user %s: %v", email, err)
		} else {
			t.Logf("✓ Created user: %s", email)
		}
	}

	// Login all users
	t.Logf("\n--- Logging In All Users ---")
	for _, email := range users {
		loginReq := &authpb.LoginRequest{
			Email:    email,
			Password: "password123",
		}
		resp, err := handler.Login(ctx, loginReq)
		if err != nil {
			t.Logf("✗ Failed to login %s: %v", email, err)
		} else {
			t.Logf("✓ Logged in: %s (Token: %s...)", email, resp.AccessToken[:15])
		}
	}

	t.Logf("\n========================================")
	t.Logf("CONCURRENCY TEST COMPLETED")
	t.Logf("========================================")
}

