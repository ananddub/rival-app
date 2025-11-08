package repo

import (
	"context"
	"database/sql"

	"encore.app/config"
	"encore.app/connection"
	db "encore.app/gen"
	authInterface "encore.app/internal/interface/auth"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthRepo struct{}

func New() authInterface.Repository {
	return &AuthRepo{}
}

func (r *AuthRepo) CreateUser(ctx context.Context, user *authInterface.User) (int64, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}

	queries := db.New(conn)
	result, err := queries.CreateUser(ctx, db.CreateUserParams{
		FullName:     user.FullName,
		Email:        user.Email,
		PhoneNumber:  pgtype.Text{String: *user.PhoneNumber, Valid: user.PhoneNumber != nil},
		PasswordHash: pgtype.Text{String: user.PasswordHash, Valid: true},
		SignType:     user.SignType,
		Role:         user.Role,
	})
	if err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (r *AuthRepo) GetUserByEmail(ctx context.Context, email string) (*authInterface.User, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var phone *string
	if user.PhoneNumber.Valid {
		phone = &user.PhoneNumber.String
	}

	return &authInterface.User{
		ID:              user.ID,
		FullName:        user.FullName,
		Email:           user.Email,
		PhoneNumber:     phone,
		PasswordHash:    user.PasswordHash.String,
		IsEmailVerified: user.IsEmailVerified.Bool,
		IsPhoneVerified: user.IsPhoneVerified.Bool,
		SignType:        user.SignType,
		Role:            user.Role,
		CreatedAt:       user.CreatedAt.Time,
		UpdatedAt:       user.UpdatedAt.Time,
	}, nil
}

func (r *AuthRepo) GetUserByID(ctx context.Context, id int64) (*authInterface.User, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	user, err := queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var phone *string
	if user.PhoneNumber.Valid {
		phone = &user.PhoneNumber.String
	}

	return &authInterface.User{
		ID:              user.ID,
		FullName:        user.FullName,
		Email:           user.Email,
		PhoneNumber:     phone,
		PasswordHash:    user.PasswordHash.String,
		IsEmailVerified: user.IsEmailVerified.Bool,
		IsPhoneVerified: user.IsPhoneVerified.Bool,
		SignType:        user.SignType,
		Role:            user.Role,
		CreatedAt:       user.CreatedAt.Time,
		UpdatedAt:       user.UpdatedAt.Time,
	}, nil
}

func (r *AuthRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.UpdatePassword(ctx, db.UpdatePasswordParams{
		ID:           userID,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	})
	return err
}

func (r *AuthRepo) VerifyEmail(ctx context.Context, userID int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.UpdateEmailVerification(ctx, db.UpdateEmailVerificationParams{
		ID:              userID,
		IsEmailVerified: pgtype.Bool{Bool: true, Valid: true},
	})
	return err
}

func (r *AuthRepo) CreateToken(ctx context.Context, token *authInterface.JWTToken) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.CreateJWTToken(ctx, db.CreateJWTTokenParams{
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: pgtype.Timestamp{Time: token.ExpiresAt, Valid: true},
	})
	return err
}

func (r *AuthRepo) GetTokenByHash(ctx context.Context, tokenHash string) (*authInterface.JWTToken, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	token, err := queries.GetJWTToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	return &authInterface.JWTToken{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt.Time,
		CreatedAt: token.CreatedAt.Time,
	}, nil
}

func (r *AuthRepo) DeleteToken(ctx context.Context, tokenHash string) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	return queries.DeleteJWTToken(ctx, tokenHash)
}
