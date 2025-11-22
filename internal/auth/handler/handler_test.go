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
	if err != nil || usr.ID == 0 {
		t.Fatalf("User not deleted: %v", email)
	}
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
