package authHandler

import (
	"context"
	"fmt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	tokens, user, err := authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	fmt.Println(user)
	return &LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User: UserInfo{
			ID:          user.ID,
			Name:        user.FullName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
		},
	}, nil
}
