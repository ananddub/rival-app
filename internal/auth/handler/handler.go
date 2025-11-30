package handler

import (
	"context"
	"errors"
	"strings"
	"time"

	"rival/config"
	authpb "rival/gen/proto/proto/api"
	"rival/internal/auth/repo"
	"rival/internal/auth/service"
	"rival/internal/auth/util"

	"google.golang.org/grpc/metadata"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	service service.AuthService
}

func NewAuthHandler() (*AuthHandler, error) {
	// Initialize repository
	repository, err := repo.NewAuthRepository()
	if err != nil {
		return nil, err
	}

	// Initialize JWT util
	cfg := config.GetConfig()
	jwtUtil := util.NewJWTUtil(cfg.JWT.Secret,
		time.Duration(cfg.JWT.ExpiryHour)*time.Hour,
		24*time.Hour)

	// Initialize email service
	emailService := util.NewEmailService()

	// Initialize service
	authService := service.NewAuthService(repository, jwtUtil, emailService, nil)

	return &AuthHandler{
		service: authService,
	}, nil
}

func (h *AuthHandler) Signup(ctx context.Context, req *authpb.SignupRequest) (*authpb.SignupResponse, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return &authpb.SignupResponse{
			Message: "Email, password and name are required",
			OtpSent: false,
		}, nil
	}

	params := service.SignupParams{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Phone:    req.Phone,
		Role:     req.Role,
	}

	return h.service.Signup(ctx, params)
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	params := service.LoginParams{
		Email:    req.Email,
		Password: req.Password,
	}

	return h.service.Login(ctx, params)
}

func (h *AuthHandler) VerifyOTP(ctx context.Context, req *authpb.VerifyOTPRequest) (*authpb.VerifyOTPResponse, error) {
	if req.Email == "" || req.Otp == "" {
		return nil, errors.New("email and OTP are required")
	}

	params := service.VerifyOTPParams{
		Email: req.Email,
		OTP:   req.Otp,
	}

	return h.service.VerifyOTP(ctx, params)
}

func (h *AuthHandler) ResendOTP(ctx context.Context, req *authpb.ResendOTPRequest) (*authpb.ResendOTPResponse, error) {
	if req.Email == "" {
		return nil, errors.New("email is required")
	}

	return h.service.ResendOTP(ctx, req.Email)
}

func (h *AuthHandler) FirebaseLogin(ctx context.Context, req *authpb.FirebaseLoginRequest) (*authpb.FirebaseLoginResponse, error) {
	if req.FirebaseToken == "" {
		return nil, errors.New("firebase token is required")
	}

	return h.service.FirebaseLogin(ctx, req.FirebaseToken)
}

func (h *AuthHandler) ForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (*authpb.ForgotPasswordResponse, error) {
	if req.Email == "" {
		return nil, errors.New("email is required")
	}

	return h.service.ForgotPassword(ctx, req.Email)
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.ResetPasswordResponse, error) {
	if req.Email == "" || req.Otp == "" || req.NewPassword == "" {
		return nil, errors.New("email, OTP and new password are required")
	}

	params := service.ResetPasswordParams{
		Email:       req.Email,
		OTP:         req.Otp,
		NewPassword: req.NewPassword,
	}

	return h.service.ResetPassword(ctx, params)
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	return h.service.RefreshToken(ctx, req.RefreshToken)
}

func (h *AuthHandler) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	if req.Token == "" {
		return nil, errors.New("token is required")
	}

	return h.service.Logout(ctx, req.Token)
}

func (h *AuthHandler) WhoAmI(ctx context.Context, req *authpb.WhoAmIRequest) (*authpb.WhoAmIResponse, error) {
	// Extract user ID from JWT token in metadata/context
	userID := extractUserIDFromContext(ctx)
	if userID == -1 {
		return nil, errors.New("unauthenticated: invalid or missing token")
	}
	return h.service.WhoAmI(ctx, int(userID))
}

func extractUserIDFromContext(ctx context.Context) int {
	// Extract JWT token from gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return -1
	}

	// Get authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return -1
	}

	// Extract token from "Bearer <token>" format
	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return -1
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Create JWT util instance (should be injected in real implementation)
	jwtUtil := util.NewJWTUtil("your-secret-key", time.Hour, 24*time.Hour)

	// Validate and extract claims
	claims, err := jwtUtil.ValidateToken(token)
	if err != nil {
		return -1
	}

	return claims.UserID
}
