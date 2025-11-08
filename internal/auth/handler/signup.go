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
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupResponse struct {
	Token string `json:"token"`
}

//encore:api public method=POST path=/auth/signup
func Signup(ctx context.Context, req *SignupRequest) (*SignupResponse, error) {
	token, err := authService.Signup(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &SignupResponse{Token: token}, nil
}
