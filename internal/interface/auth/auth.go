package authHandler

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
	DOB             *time.Time
	IsPhoneVerified bool
	IsEmailVerified bool
	SignType        string
	Role            string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type JWTToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Repository interface {
	CreateUser(ctx context.Context, user *User) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	VerifyEmail(ctx context.Context, userID int64) error

	CreateToken(ctx context.Context, token *JWTToken) error
	GetTokenByHash(ctx context.Context, tokenHash string) (*JWTToken, error)
	DeleteToken(ctx context.Context, tokenHash string) error
}

type Service interface {
	Signup(ctx context.Context, fullName, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
	VerifyToken(ctx context.Context, token string) (*User, error)
	Logout(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, email string) error
}
