package authHandler

import "context"

type LogoutRequest struct {
	Token string `json:"token"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/auth/logout
func Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	if err := authService.Logout(ctx, req.Token); err != nil {
		return nil, err
	}

	return &LogoutResponse{Message: "Logged out successfully"}, nil
}
