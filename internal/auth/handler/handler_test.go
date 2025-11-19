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

func TestSignup_Real(t *testing.T) {
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
	defer func() {
		repo.queries.DeleteByEmail(context.Background(), tests.Email)
		//repo.tb.Close()
		repo.db.Close()
	}()
	if tbuser == nil || len(*tbuser) == 0 {
		t.Errorf("Signup() failed to create tigerbeetle account with correct ID")
	}

	t.Log(user, tbuser)
	if err != nil {
		t.Errorf("Signup() error = %v", err)
	}
}

func TestLogin_Real(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	ctx := context.Background()
	value, err := handler.Signup(ctx,
		&authpb.SignupRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
			Role:     *schemapb.UserRole_USER_ROLE_ADMIN.Enum(),
			Phone:    "12345678",
		},
	)
	if err != nil {
		t.Fatalf("Failed to signup user: %v", err)
	}
	t.Logf("Signup result:%v", value)
	defer func() {
		repo, err := NewRepo()
		if err != nil {
			t.Fatalf("Failed to create repo: %v", err)
		}
		repo.queries.DeleteByEmail(ctx, "test@example.com")
		repo.db.Close()
	}()
	req := &authpb.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	loginValue, err := handler.Login(ctx, req)
	if err != nil {
		t.Errorf("Login() error = %v", err)
	}
	t.Logf("Login result:%v", loginValue)
}

func TestVerifyOTP_Real(t *testing.T) {
	handler, err := NewAuthHandler()
	ctx := context.Background()
	handler.Signup(
		ctx,
		&authpb.SignupRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
			Role:     *schemapb.UserRole_USER_ROLE_ADMIN.Enum(),
			Phone:    "12345678",
		},
	)
	handler.Login(
		ctx,
		&authpb.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		},
	)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	req := &authpb.VerifyOTPRequest{
		Email: "test@example.com",
		Otp:   "123456",
	}

	_, err = handler.VerifyOTP(context.Background(), req)
	t.Logf("VerifyOTP result: %v", err)
}

func TestResendOTP_Real(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	req := &authpb.ResendOTPRequest{
		Email: "test@example.com",
	}

	_, err = handler.ResendOTP(context.Background(), req)
	t.Logf("ResendOTP result: %v", err)
}

func TestFirebaseLogin_Real(t *testing.T) {
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

func TestForgotPassword_Real(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	req := &authpb.ForgotPasswordRequest{
		Email: "test@example.com",
	}

	_, err = handler.ForgotPassword(context.Background(), req)
	t.Logf("ForgotPassword result: %v", err)
}

func TestResetPassword_Real(t *testing.T) {
	handler, err := NewAuthHandler()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	req := &authpb.ResetPasswordRequest{
		Email:       "test@example.com",
		Otp:         "123456",
		NewPassword: "newpass123",
	}

	_, err = handler.ResetPassword(context.Background(), req)
	t.Logf("ResetPassword result: %v", err)
}

func TestRefreshToken_Real(t *testing.T) {
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

func TestLogout_Real(t *testing.T) {
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

func TestWhoAmI_Real(t *testing.T) {
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
