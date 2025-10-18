package authHandler

import (
	"context"
	"fmt"

	"encore.app/internal/auth/repo"
	"encore.app/internal/auth/service"
)

type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

type SignupResponse struct {
	Message string `json:"message"`
}

type SignupVerifyOtpRequest struct {
	Email string `json:"email"`
	Otp   string `json:"otp"`
}

type SignupVerifyOtpResponse struct {
	User User `json:"user"`
}

//encore:api public method=POST path=/auth/signup
func Signup(ctx context.Context, req *SignupRequest) (*SignupResponse, error) {
	// Check if user already exists
	_, err := authRepo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, fmt.Errorf("user already exists")
	}

	hashedPassword, err := authService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user in database directly
	user, err := authRepo.CreateUser(ctx, req.Name, req.Email, req.Phone, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Generate OTP and store in Redis
	otp := authService.GenerateOTP()
	if err := authService.StoreOTP(ctx, req.Email, otp); err != nil {
		return nil, fmt.Errorf("failed to store OTP: %v", err)
	}

	fmt.Printf("Signup OTP for %s (ID: %s): %s\n", req.Email, user.ID, otp)

	return &SignupResponse{
		Message: "User created. OTP sent to your email for verification",
	}, nil
}

//encore:api public method=POST path=/auth/signup/verify-otp
func SignupVerifyOtp(ctx context.Context, req *SignupVerifyOtpRequest) (*SignupVerifyOtpResponse, error) {
	if !authService.VerifyOTP(ctx, req.Email, req.Otp) {
		return nil, fmt.Errorf("invalid OTP")
	}

	// Get user from database
	user, err := authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Mark email as verified
	if err := authRepo.UpdateEmailVerification(ctx, user.ID, true); err != nil {
		return nil, fmt.Errorf("failed to update email verification: %v", err)
	}

	return &SignupVerifyOtpResponse{
		User: User{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			Phone:           user.Phone,
			IsEmailVerified: true,
			IsPhoneVerified: user.IsPhoneVerified,
		},
	}, nil
}

//encore:api public method=POST path=/auth/signup/resend-otp
func SignupResendOtp(ctx context.Context, req *SignupVerifyOtpRequest) (*SignupResponse, error) {
	// Check if user exists
	user, err := authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.IsEmailVerified {
		return nil, fmt.Errorf("email already verified")
	}

	// Generate new OTP and store in Redis
	otp := authService.GenerateOTP()
	if err := authService.StoreOTP(ctx, req.Email, otp); err != nil {
		return nil, fmt.Errorf("failed to store OTP: %v", err)
	}

	fmt.Printf("Resend Signup OTP for %s (ID: %s): %s\n", req.Email, user.ID, otp)

	return &SignupResponse{
		Message: "OTP resent to your email",
	}, nil
}
