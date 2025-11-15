package authHandler

import (
	"context"

	"encore.app/internal/auth/repo"
	"encore.app/internal/auth/service"
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
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

//encore:api public method=POST path=/auth/signup
func Signup(ctx context.Context, req *SignupRequest) (*SignupResponse, error) {
	tokens, user, err := authService.Signup(ctx, req.Name, req.Email, req.PhoneNumber, req.Password)
	if err != nil {
		return nil, err
	}

	return &SignupResponse{
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
