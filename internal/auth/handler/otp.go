package authHandler

import (
	"context"
	
	"encore.app/internal/auth/utils"
)

type SendOtpRequest struct {
	Email string `json:"email"`
}

type SendOtpResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/auth/send-otp
func SendOtp(ctx context.Context, req *SendOtpRequest) (*SendOtpResponse, error) {
	otp := utils.StoreOtp(req.Email)
	// TODO: Send email with OTP
	_ = otp // Use the generated OTP
	return &SendOtpResponse{Message: "OTP sent successfully"}, nil
}

type VerifyOtpRequest struct {
	Email string `json:"email"`
	Otp   string `json:"otp"`
}

type VerifyOtpResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

//encore:api public method=POST path=/auth/verify-otp
func VerifyOtp(ctx context.Context, req *VerifyOtpRequest) (*VerifyOtpResponse, error) {
	valid, err := authService.VerifyOtp(ctx, req.Email, req.Otp)
	if err != nil {
		return nil, err
	}

	if valid {
		return &VerifyOtpResponse{Valid: true, Message: "OTP verified successfully"}, nil
	}

	return &VerifyOtpResponse{Valid: false, Message: "Invalid OTP"}, nil
}
