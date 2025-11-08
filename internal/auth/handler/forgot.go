package authHandler

import "context"

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/auth/forgot-password
func ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) (*ForgotPasswordResponse, error) {
	if err := authService.ForgotPassword(ctx, req.Email); err != nil {
		return nil, err
	}

	return &ForgotPasswordResponse{Message: "Password reset link sent to email"}, nil
}
