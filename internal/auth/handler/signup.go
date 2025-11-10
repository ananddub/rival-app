package authHandler

import (
	"context"
	"errors"

	"encore.app/internal/auth/repo"
	"encore.app/internal/auth/service"
	"encore.app/internal/auth/utils"
)

var (
	authRepo    = repo.New()
	authService = service.New(authRepo)
)

type SignupRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type SignupResponse struct {
	Message string `json:"message"`
}

type UserInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

//encore:api public method=POST path=/auth/signup
func Signup(ctx context.Context, req *SignupRequest) (*SignupResponse, error) {
	// Validate input
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if user already exists
	existing, err := authRepo.GetUserByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, errors.New("user already exists")
	}

	// Store signup data temporarily
	utils.StorePendingSignup(req.Email, req.Name, req.PhoneNumber, req.Password)
	
	// Send OTP
	utils.StoreOtp(req.Email)
	
	return &SignupResponse{
		Message: "OTP sent to your email. Please verify to complete registration.",
	}, nil
}
