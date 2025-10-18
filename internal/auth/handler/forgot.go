package authHandler

import (
	"context"
	"fmt"

	"encore.app/internal/auth/repo"
	"encore.app/internal/auth/service"
)

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Otp         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/auth/forgot-password
func ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) (*ForgotPasswordResponse, error) {
	// Check if user exists
	_, err := authRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	otp := authService.GenerateOTP()
	if err := authService.StoreOTP(ctx, req.Email, otp); err != nil {
		return nil, fmt.Errorf("failed to store OTP: %v", err)
	}
	
	fmt.Printf("Reset OTP for %s: %s\n", req.Email, otp)
	
	return &ForgotPasswordResponse{
		Message: "OTP sent to your email",
	}, nil
}

//encore:api public method=POST path=/auth/reset-password
func ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*ResetPasswordResponse, error) {
	if !authService.VerifyOTP(ctx, req.Email, req.Otp) {
		return nil, fmt.Errorf("invalid OTP")
	}
	
	hashedPassword, err := authService.HashPassword(req.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	
	if err := authRepo.UpdateUserPassword(ctx, req.Email, hashedPassword); err != nil {
		return nil, fmt.Errorf("failed to update password: %v", err)
	}
	
	return &ResetPasswordResponse{
		Message: "Password reset successfully",
	}, nil
}
