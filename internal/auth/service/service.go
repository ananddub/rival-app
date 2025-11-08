package service

import (
	"context"
	"errors"
	"time"

	authInterface "encore.app/internal/interface/auth"
	"encore.app/internal/auth/utils"
)

type AuthService struct {
	repo authInterface.Repository
}

func New(repo authInterface.Repository) authInterface.Service {
	return &AuthService{repo: repo}
}

func (s *AuthService) Signup(ctx context.Context, fullName, email, password string) (string, error) {
	if err := utils.ValidateEmail(email); err != nil {
		return "", err
	}
	if err := utils.ValidatePassword(password); err != nil {
		return "", err
	}

	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return "", errors.New("user already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return "", err
	}

	user := &authInterface.User{
		FullName:     fullName,
		Email:        email,
		PasswordHash: hashedPassword,
		SignType:     "email",
		Role:         "user",
	}

	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	return s.generateToken(ctx, userID, email)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	if err := utils.VerifyPassword(user.PasswordHash, password); err != nil {
		return "", errors.New("invalid credentials")
	}

	return s.generateToken(ctx, user.ID, user.Email)
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
	// TODO: Send email with reset link
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
