package authHandler

import "context"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	token, err := authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token}, nil
}
