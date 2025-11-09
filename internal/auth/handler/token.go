package authHandler

import "context"

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

//encore:api public method=POST path=/auth/refresh-token
func RefreshToken(ctx context.Context, req *TokenRefreshRequest) (*RefreshTokenResponse, error) {
	tokens, err := authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
