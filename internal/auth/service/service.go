package service

import (
	"context"
	"errors"
	"time"

	"encore.app/internal/auth/utils"
	authInterface "encore.app/internal/interface/auth"
)

type AuthService struct {
	repo authInterface.Repository
}

func New(repo authInterface.Repository) authInterface.Service {
	return &AuthService{repo: repo}
}

func (s *AuthService) Signup(ctx context.Context, fullName, email, phoneNumber, password string) (*authInterface.TokenPair, *authInterface.User, error) {
	if err := utils.ValidateEmail(email); err != nil {
		return nil, nil, err
	}
	if err := utils.ValidatePassword(password); err != nil {
		return nil, nil, err
	}

	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, errors.New("user already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	user := &authInterface.User{
		FullName:     fullName,
		Email:        email,
		PhoneNumber:  phoneNumber,
		PasswordHash: hashedPassword,
		SignType:     "email",
		Role:         "user",
	}

	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	user.ID = userID
	tokens, err := s.generateTokenPair(ctx, userID, email)
	if err != nil {
		return nil, nil, err
	}

	return tokens, user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*authInterface.TokenPair, *authInterface.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.New("invalid credentials")
	}

	if err := utils.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	tokens, err := s.generateTokenPair(ctx, user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return tokens, user, nil
}

func (s *AuthService) LoginOrCreate(ctx context.Context, email, password, name, phoneNumber string) (*authInterface.TokenPair, *authInterface.User, error) {
	// First try to login existing user
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	// If user exists, verify password and login
	if user != nil {
		if err := utils.VerifyPassword(user.PasswordHash, password); err != nil {
			return nil, nil, errors.New("invalid credentials")
		}

		tokens, err := s.generateTokenPair(ctx, user.ID, user.Email)
		if err != nil {
			return nil, nil, err
		}
		return tokens, user, nil
	}

	// If user doesn't exist, create new user (from signup data)
	if name == "" {
		return nil, nil, errors.New("invalid credentials")
	}

	if err := utils.ValidateEmail(email); err != nil {
		return nil, nil, err
	}
	if err := utils.ValidatePassword(password); err != nil {
		return nil, nil, err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	newUser := &authInterface.User{
		FullName:     name,
		Email:        email,
		PhoneNumber:  phoneNumber,
		PasswordHash: hashedPassword,
		SignType:     "email",
		Role:         "user",
	}

	userID, err := s.repo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, nil, err
	}

	newUser.ID = userID
	tokens, err := s.generateTokenPair(ctx, userID, email)
	if err != nil {
		return nil, nil, err
	}

	return tokens, newUser, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*authInterface.TokenPair, error) {
	// Verify refresh token
	storedToken, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if time.Now().After(storedToken.ExpiresAt) {
		s.repo.DeleteRefreshToken(ctx, refreshToken)
		return nil, errors.New("refresh token expired")
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	tokens, err := s.generateTokenPair(ctx, user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token
	s.repo.DeleteRefreshToken(ctx, refreshToken)

	return tokens, nil
}

func (s *AuthService) generateTokenPair(ctx context.Context, userID int64, email string) (*authInterface.TokenPair, error) {
	// Generate access token (15 minutes)
	accessToken, err := utils.GenerateAccessToken(userID, email)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (7 days)
	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store access token
	tokenHash := utils.HashToken(accessToken)
	jwtToken := &authInterface.JWTToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.CreateToken(ctx, jwtToken); err != nil {
		return nil, err
	}

	// Store refresh token
	refreshTokenRecord := &authInterface.RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.repo.CreateRefreshToken(ctx, refreshTokenRecord); err != nil {
		return nil, err
	}

	return &authInterface.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, tokenString string) (*authInterface.User, error) {
	userID, _, err := utils.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	tokenHash := utils.HashToken(token)
	return s.repo.DeleteToken(ctx, tokenHash)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	otp := utils.StoreOtp(email)
	// TODO: Send email with OTP
	_ = otp // Use the generated OTP
	return nil
}

func (s *AuthService) generateToken(ctx context.Context, userID int64, email string) (string, error) {
	tokenString, err := utils.GenerateToken(userID, email)
	if err != nil {
		return "", err
	}

	tokenHash := utils.HashToken(tokenString)
	jwtToken := &authInterface.JWTToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.repo.CreateToken(ctx, jwtToken); err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) VerifyOtp(ctx context.Context, email, otp string) (bool, error) {
	return utils.VerifyOtp(email, otp), nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, newPassword, otp string) error {
	if !utils.VerifyOtp(email, otp) {
		return errors.New("invalid OTP")
	}

	if err := utils.ValidatePassword(newPassword); err != nil {
		return err
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, user.ID, hashedPassword)
}
