package authHandler

import "context"

type VerifyTokenRequest struct {
	Token string `json:"token"`
}

type VerifyTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID int64  `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
}

//encore:api public method=POST path=/auth/verify
func VerifyToken(ctx context.Context, req *VerifyTokenRequest) (*VerifyTokenResponse, error) {
	user, err := authService.VerifyToken(ctx, req.Token)
	if err != nil {
		return &VerifyTokenResponse{Valid: false}, nil
	}

	return &VerifyTokenResponse{
		Valid:  true,
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}
