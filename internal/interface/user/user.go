package user

import (
	"context"
	"time"
)

type User struct {
	ID              int64
	FullName        string
	Email           string
	PhoneNumber     *string
	PasswordHash    string
	ProfilePhoto    *string
	DOB             *time.Time
	IsPhoneVerified bool
	IsEmailVerified bool
	SignType        string
	Role            string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Repository interface {
	GetUserByID(ctx context.Context, id int64) (*User, error)
	UpdateProfile(ctx context.Context, userID int64, fullName, phoneNumber, dob string) (*User, error)
	UpdateProfilePhoto(ctx context.Context, userID int64, photoPath string) error
	DeleteProfilePhoto(ctx context.Context, userID int64) error
	VerifyEmail(ctx context.Context, userID int64) error
	VerifyPhone(ctx context.Context, userID int64) error
}

type Service interface {
	GetProfile(ctx context.Context, userID int64) (*User, error)
	UpdateProfile(ctx context.Context, userID int64, fullName, phoneNumber, dob string) (*User, error)
	UpdateProfilePhoto(ctx context.Context, userID int64, photoPath string) error
	DeleteProfilePhoto(ctx context.Context, userID int64) error
	VerifyEmail(ctx context.Context, userID int64, otp string) error
	VerifyPhone(ctx context.Context, userID int64, otp string) error
}
