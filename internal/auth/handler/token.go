package authHandler

import "context"

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

//encore:api public method=POST path=/auth/refresh-token
func RefreshToken(ctx context.Context) (*RefreshTokenResponse, error) {
	return &RefreshTokenResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}, nil
}
