package authHandler

import (
	"context"
	"fmt"

	"encore.app/internal/auth/repo"
	"encore.app/internal/auth/service"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	IsEmailVerified bool   `json:"is_email_verified"`
	IsPhoneVerified bool   `json:"is_phone_verified"`
}

type LoginResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type VerifyOtpRequest struct {
	Email string `json:"email"`
	Otp   string `json:"otp"`
}

type VerifyOtpResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user from database
	user, err := authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	accessToken, err := authService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := authService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	return &LoginResponse{
		User: User{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			IsEmailVerified: user.IsEmailVerified,
			IsPhoneVerified: user.IsPhoneVerified,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

//encore:api public method=POST path=/auth/login/verify-otp
func VerifyOtp(ctx context.Context, req *VerifyOtpRequest) (*VerifyOtpResponse, error) {
	if !authService.VerifyOTP(ctx, req.Email, req.Otp) {
		return nil, fmt.Errorf("invalid OTP")
	}

	// Get user from database
	user, err := authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate tokens
	accessToken, err := authService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := authService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	return &VerifyOtpResponse{
		User: User{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			IsEmailVerified: user.IsEmailVerified,
			IsPhoneVerified: user.IsPhoneVerified,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
