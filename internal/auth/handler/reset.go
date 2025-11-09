package authHandler

import "context"

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
	Otp         string `json:"otp"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/auth/reset-password
func ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*ResetPasswordResponse, error) {
	if err := authService.ResetPassword(ctx, req.Email, req.NewPassword, req.Otp); err != nil {
		return nil, err
	}

	return &ResetPasswordResponse{Message: "Password reset successfully"}, nil
}
