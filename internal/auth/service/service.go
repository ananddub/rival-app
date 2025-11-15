package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
	"math/big"
	
	authpb "encore.app/gen/proto/proto/api"
	schemapb "encore.app/gen/proto/proto/schema"
	schema "encore.app/gen/sql"
	"encore.app/config"
	"encore.app/connection"
	"encore.app/internal/auth/repo"
	"encore.app/internal/auth/util"
	"encore.app/pkg/referral"
	"encore.app/pkg/tigerbeetle"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type SignupParams struct {
	Email       string
	Password    string
	Name        string
	Phone       string
	Role        schemapb.UserRole
	ReferralCode string // Optional referral code
}

type LoginParams struct {
	Email    string
	Password string
}

type VerifyOTPParams struct {
	Email string
	OTP   string
}

type ResetPasswordParams struct {
	Email       string
	OTP         string
	NewPassword string
}

type AuthService interface {
	Signup(ctx context.Context, params SignupParams) (*authpb.SignupResponse, error)
	VerifyOTP(ctx context.Context, params VerifyOTPParams) (*authpb.VerifyOTPResponse, error)
	ResendOTP(ctx context.Context, email string) (*authpb.ResendOTPResponse, error)
	Login(ctx context.Context, params LoginParams) (*authpb.LoginResponse, error)
	FirebaseLogin(ctx context.Context, firebaseToken string) (*authpb.FirebaseLoginResponse, error)
	ForgotPassword(ctx context.Context, email string) (*authpb.ForgotPasswordResponse, error)
	ResetPassword(ctx context.Context, params ResetPasswordParams) (*authpb.ResetPasswordResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*authpb.RefreshTokenResponse, error)
	Logout(ctx context.Context, token string) (*authpb.LogoutResponse, error)
	WhoAmI(ctx context.Context, userID string) (*authpb.WhoAmIResponse, error)
}

type authService struct {
	repo     repo.AuthRepository
	jwt      util.JWTUtil
	email    util.EmailService
	firebase util.FirebaseService
	tb       tigerbeetle.Service
	referral *referral.Service
}

func NewAuthService(authRepo repo.AuthRepository, jwt util.JWTUtil, email util.EmailService, firebase util.FirebaseService) AuthService {
	// Initialize TigerBeetle service
	tbService, _ := tigerbeetle.NewService()
	
	// Initialize referral service
	cfg := config.GetConfig()
	db, _ := connection.GetPgConnection(&cfg.Database)
	referralService := referral.NewService(db, tbService)
	
	return &authService{
		repo:     authRepo,
		jwt:      jwt,
		email:    email,
		firebase: firebase,
		tb:       tbService,
		referral: referralService,
	}
}

func (s *authService) Signup(ctx context.Context, params SignupParams) (*authpb.SignupResponse, error) {
	// Check if user already exists
	_, err := s.repo.GetUserByEmail(ctx, params.Email)
	if err == nil {
		return &authpb.SignupResponse{
			Message: "User already exists",
			OtpSent: false,
		}, nil
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	createParams := schema.CreateUserParams{
		Email:        params.Email,
		PasswordHash: pgtype.Text{String: string(hashedPassword), Valid: true},
		Phone:        pgtype.Text{String: params.Phone, Valid: params.Phone != ""},
		Name:         params.Name,
		Role:         schema.NullUserRole{UserRole: schema.UserRole(params.Role.String()), Valid: true},
	}

	user, err := s.repo.CreateUser(ctx, createParams)
	if err != nil {
		return nil, err
	}

	// Give initial signup bonus coins (e.g., 10 coins)
	err = s.giveInitialCoins(user.ID, 10.0)
	if err != nil {
		// Log error but don't fail signup
		fmt.Printf("Failed to give initial coins: %v\n", err)
	}

	// Process referral if provided
	if params.ReferralCode != "" && s.referral != nil {
		userIDStr, _ := user.ID.Value()
		err = s.referral.ProcessReferral(ctx, params.ReferralCode, userIDStr.(string))
		if err != nil {
			// Log error but don't fail signup
			fmt.Printf("Failed to process referral: %v\n", err)
		}
	}

	// Generate and send OTP
	otp := generateOTP()
	err = s.repo.StoreOTP(ctx, params.Email, otp, 10*time.Minute)
	if err != nil {
		return nil, err
	}

	err = s.email.SendOTP(params.Email, otp)
	if err != nil {
		return nil, err
	}

	// Send welcome email
	s.email.SendWelcomeEmail(params.Email, user.Name)

	return &authpb.SignupResponse{
		Message: "OTP sent to your email",
		OtpSent: true,
	}, nil
}

func (s *authService) VerifyOTP(ctx context.Context, params VerifyOTPParams) (*authpb.VerifyOTPResponse, error) {
	// Verify OTP
	valid, err := s.repo.VerifyOTP(ctx, params.Email, params.OTP)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid OTP")
	}

	// Get user
	user, err := s.repo.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	protoUser := convertToProtoUser(user)
	accessToken, refreshToken, err := s.jwt.GenerateTokens(protoUser)
	if err != nil {
		return nil, err
	}

	// Store session
	sessionParams := schema.CreateJWTSessionParams{
		UserID:           user.ID,
		TokenHash:        s.jwt.HashToken(accessToken),
		RefreshTokenHash: pgtype.Text{String: s.jwt.HashToken(refreshToken), Valid: true},
		ExpiresAt:        pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
	}

	err = s.repo.CreateSession(ctx, sessionParams)
	if err != nil {
		return nil, err
	}

	return &authpb.VerifyOTPResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         protoUser,
		ExpiresIn:    86400, // 24 hours
	}, nil
}

func (s *authService) ResendOTP(ctx context.Context, email string) (*authpb.ResendOTPResponse, error) {
	// Check if user exists
	_, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return &authpb.ResendOTPResponse{
			Message: "User not found",
			OtpSent: false,
		}, nil
	}

	// Generate and send new OTP
	otp := generateOTP()
	err = s.repo.StoreOTP(ctx, email, otp, 10*time.Minute)
	if err != nil {
		return nil, err
	}

	err = s.email.SendOTP(email, otp)
	if err != nil {
		return nil, err
	}

	return &authpb.ResendOTPResponse{
		Message: "OTP sent to your email",
		OtpSent: true,
	}, nil
}

func (s *authService) Login(ctx context.Context, params LoginParams) (*authpb.LoginResponse, error) {
	// Get user
	user, err := s.repo.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(params.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	protoUser := convertToProtoUser(user)
	accessToken, refreshToken, err := s.jwt.GenerateTokens(protoUser)
	if err != nil {
		return nil, err
	}

	// Store session
	sessionParams := schema.CreateJWTSessionParams{
		UserID:           user.ID,
		TokenHash:        s.jwt.HashToken(accessToken),
		RefreshTokenHash: pgtype.Text{String: s.jwt.HashToken(refreshToken), Valid: true},
		ExpiresAt:        pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
	}

	err = s.repo.CreateSession(ctx, sessionParams)
	if err != nil {
		return nil, err
	}

	return &authpb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         protoUser,
		ExpiresIn:    86400,
	}, nil
}

func (s *authService) FirebaseLogin(ctx context.Context, firebaseToken string) (*authpb.FirebaseLoginResponse, error) {
	// Verify Firebase token
	firebaseUser, err := s.firebase.VerifyToken(ctx, firebaseToken)
	if err != nil {
		return nil, fmt.Errorf("invalid firebase token: %v", err)
	}

	// Check if user exists
	user, err := s.repo.GetUserByEmail(ctx, firebaseUser.Email)
	if err != nil {
		// Create new user if doesn't exist
		createParams := schema.CreateUserParams{
			Email:       firebaseUser.Email,
			Name:        firebaseUser.Name,
			ProfilePic:  pgtype.Text{String: firebaseUser.Picture, Valid: firebaseUser.Picture != ""},
			Phone:       pgtype.Text{String: firebaseUser.PhoneNumber, Valid: firebaseUser.PhoneNumber != ""},
			FirebaseUid: pgtype.Text{String: firebaseUser.UID, Valid: true},
			Role:        schema.NullUserRole{UserRole: schema.UserRoleCustomer, Valid: true},
		}

		user, err = s.repo.CreateUser(ctx, createParams)
		if err != nil {
			return nil, err
		}

		// Send welcome email
		s.email.SendWelcomeEmail(firebaseUser.Email, firebaseUser.Name)
	}

	// Generate tokens
	protoUser := convertToProtoUser(user)
	accessToken, refreshToken, err := s.jwt.GenerateTokens(protoUser)
	if err != nil {
		return nil, err
	}

	// Store session
	sessionParams := schema.CreateJWTSessionParams{
		UserID:           user.ID,
		TokenHash:        s.jwt.HashToken(accessToken),
		RefreshTokenHash: pgtype.Text{String: s.jwt.HashToken(refreshToken), Valid: true},
		ExpiresAt:        pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
	}

	err = s.repo.CreateSession(ctx, sessionParams)
	if err != nil {
		return nil, err
	}

	return &authpb.FirebaseLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         protoUser,
		ExpiresIn:    86400,
	}, nil
}

func (s *authService) ForgotPassword(ctx context.Context, email string) (*authpb.ForgotPasswordResponse, error) {
	// Check if user exists
	_, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return &authpb.ForgotPasswordResponse{
			Message: "If email exists, OTP has been sent",
			OtpSent: true,
		}, nil
	}

	// Generate and send reset OTP
	otp := generateOTP()
	err = s.repo.StoreOTP(ctx, "reset:"+email, otp, 10*time.Minute)
	if err != nil {
		return nil, err
	}

	err = s.email.SendPasswordResetEmail(email, otp)
	if err != nil {
		return nil, err
	}

	return &authpb.ForgotPasswordResponse{
		Message: "Password reset OTP sent to your email",
		OtpSent: true,
	}, nil
}

func (s *authService) ResetPassword(ctx context.Context, params ResetPasswordParams) (*authpb.ResetPasswordResponse, error) {
	// Verify reset OTP
	valid, err := s.repo.VerifyOTP(ctx, "reset:"+params.Email, params.OTP)
	if err != nil || !valid {
		return &authpb.ResetPasswordResponse{
			Message: "Invalid OTP",
			Success: false,
		}, nil
	}

	// Get user
	user, err := s.repo.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Update password
	updateParams := schema.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: pgtype.Text{String: string(hashedPassword), Valid: true},
	}

	err = s.repo.UpdateUserPassword(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	return &authpb.ResetPasswordResponse{
		Message: "Password reset successfully",
		Success: true,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*authpb.RefreshTokenResponse, error) {
	// Generate new access token
	newAccessToken, err := s.jwt.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return &authpb.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
	}, nil
}

func (s *authService) Logout(ctx context.Context, token string) (*authpb.LogoutResponse, error) {
	// Revoke session
	tokenHash := s.jwt.HashToken(token)
	err := s.repo.RevokeSession(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	return &authpb.LogoutResponse{Success: true}, nil
}

func (s *authService) WhoAmI(ctx context.Context, userID string) (*authpb.WhoAmIResponse, error) {
	// Get user by ID
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &authpb.WhoAmIResponse{
		User: convertToProtoUser(user),
	}, nil
}

func (s *authService) giveInitialCoins(userID pgtype.UUID, amount float64) error {
	// Convert pgtype.UUID to uuid.UUID
	userIDVal, _ := userID.Value()
	userUUID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		return err
	}

	// Use TigerBeetle service to add coins
	return s.tb.AddCoins(userUUID, amount)
}

func generateOTP() string {
	max := big.NewInt(999999)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

func convertToProtoUser(user schema.User) *schemapb.User {
	userID, _ := user.ID.Value()
	
	// Handle coin balance
	var coinBalance float64
	if user.CoinBalance.Valid {
		val, _ := user.CoinBalance.Value()
		if val != nil {
			coinBalance = val.(float64)
		}
	}
	
	// Handle role
	var role schemapb.UserRole
	if user.Role.Valid {
		if val, ok := schemapb.UserRole_value[string(user.Role.UserRole)]; ok {
			role = schemapb.UserRole(val)
		}
	}
	
	// Handle referred by
	var referredBy string
	if user.ReferredBy.Valid {
		referredByVal, _ := user.ReferredBy.Value()
		referredBy = referredByVal.(string)
	}
	
	return &schemapb.User{
		Id:           userID.(string),
		Email:        user.Email,
		PasswordHash: user.PasswordHash.String,
		Phone:        user.Phone.String,
		Name:         user.Name,
		ProfilePic:   user.ProfilePic.String,
		FirebaseUid:  user.FirebaseUid.String,
		CoinBalance:  coinBalance,
		Role:         role,
		ReferralCode: user.ReferralCode.String,
		ReferredBy:   referredBy,
		CreatedAt:    user.CreatedAt.Time.Unix(),
		UpdatedAt:    user.UpdatedAt.Time.Unix(),
	}
}
