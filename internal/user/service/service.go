package service

import (
	"context"
	"errors"
	"os"
	"strings"

	userInterface "encore.app/internal/interface/user"
	"encore.app/internal/auth/utils"
)

type UserService struct {
	repo userInterface.Repository
}

func New(repo userInterface.Repository) userInterface.Service {
	return &UserService{repo: repo}
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*userInterface.User, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.GetUserByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, fullName, phoneNumber, dob string) (*userInterface.User, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	if fullName != "" {
		if len(fullName) < 2 || len(fullName) > 100 {
			return nil, errors.New("full name must be between 2 and 100 characters")
		}
	}

	if phoneNumber != "" {
		if len(phoneNumber) < 10 || len(phoneNumber) > 15 {
			return nil, errors.New("invalid phone number")
		}
	}

	if dob != "" {
		// Basic date format validation (YYYY-MM-DD)
		if len(dob) != 10 || !strings.Contains(dob, "-") {
			return nil, errors.New("invalid date format. Use YYYY-MM-DD")
		}
	}

	return s.repo.UpdateProfile(ctx, userID, fullName, phoneNumber, dob)
}

func (s *UserService) UpdateProfilePhoto(ctx context.Context, userID int64, photoPath string) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	if photoPath == "" {
		return errors.New("photo path cannot be empty")
	}

	return s.repo.UpdateProfilePhoto(ctx, userID, photoPath)
}

func (s *UserService) DeleteProfilePhoto(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	// Get current user to find photo path
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Delete photo file if exists
	if user.ProfilePhoto != nil && *user.ProfilePhoto != "" {
		photoPath := "." + *user.ProfilePhoto
		if _, err := os.Stat(photoPath); err == nil {
			os.Remove(photoPath)
		}
	}

	return s.repo.DeleteProfilePhoto(ctx, userID)
}

func (s *UserService) VerifyEmail(ctx context.Context, userID int64, otp string) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	// Get user email for OTP verification
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify OTP (using auth utils)
	if !utils.VerifyOtp(user.Email, otp) {
		return errors.New("invalid OTP")
	}

	return s.repo.VerifyEmail(ctx, userID)
}

func (s *UserService) VerifyPhone(ctx context.Context, userID int64, otp string) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	// Get user phone for OTP verification
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.PhoneNumber == nil {
		return errors.New("phone number not set")
	}

	// Verify OTP (using auth utils)
	if !utils.VerifyOtp(*user.PhoneNumber, otp) {
		return errors.New("invalid OTP")
	}

	return s.repo.VerifyPhone(ctx, userID)
}
