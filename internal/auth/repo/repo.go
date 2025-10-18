package authRepo

import (
	"context"
	"database/sql"
	"strconv"

	"encore.app/config"
	"encore.app/connection"
	auth_gen "encore.app/internal/auth/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID              string
	Name            string
	Email           string
	Phone           string
	Password        string
	IsEmailVerified bool
	IsPhoneVerified bool
}

func GetUserByEmail(ctx context.Context, email string) (*User, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	
	queries := auth_gen.New(db)

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &User{
		ID:              strconv.FormatInt(user.ID, 10),
		Name:            user.FullName,
		Email:           user.Email,
		Phone:           user.PhoneNumber.String,
		Password:        user.PasswordHash.String,
		IsEmailVerified: user.IsEmailVerified.Bool,
		IsPhoneVerified: user.IsPhoneVerified.Bool,
	}, nil
}

func CreateUser(ctx context.Context, name, email, phone, hashedPassword string) (*User, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	
	queries := auth_gen.New(db)

	user, err := queries.CreateUser(ctx, auth_gen.CreateUserParams{
		FullName:     name,
		Email:        email,
		PhoneNumber:  pgtype.Text{String: phone, Valid: true},
		PasswordHash: pgtype.Text{String: hashedPassword, Valid: true},
		SignType:     auth_gen.SignTypeEmail,
		Role:         auth_gen.RoleTypeUser,
	})
	if err != nil {
		return nil, err
	}

	return &User{
		ID:              strconv.FormatInt(user.ID, 10),
		Name:            user.FullName,
		Email:           user.Email,
		Phone:           user.PhoneNumber.String,
		Password:        user.PasswordHash.String,
		IsEmailVerified: user.IsEmailVerified.Bool,
		IsPhoneVerified: user.IsPhoneVerified.Bool,
	}, nil
}

func UpdateUserPassword(ctx context.Context, email, hashedPassword string) error {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}
	
	queries := auth_gen.New(db)

	// First get user by email to get ID
	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	_, err = queries.UpdatePassword(ctx, auth_gen.UpdatePasswordParams{
		ID:           user.ID,
		PasswordHash: pgtype.Text{String: hashedPassword, Valid: true},
	})
	
	return err
}

func UpdateEmailVerification(ctx context.Context, userID string, verified bool) error {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}
	
	queries := auth_gen.New(db)
	
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return err
	}

	_, err = queries.UpdateEmailVerification(ctx, auth_gen.UpdateEmailVerificationParams{
		ID:               id,
		IsEmailVerified:  pgtype.Bool{Bool: verified, Valid: true},
	})
	
	return err
}

func UpdatePhoneVerification(ctx context.Context, userID string, verified bool) error {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}
	
	queries := auth_gen.New(db)
	
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return err
	}

	_, err = queries.UpdatePhoneVerification(ctx, auth_gen.UpdatePhoneVerificationParams{
		ID:               id,
		IsPhoneVerified:  pgtype.Bool{Bool: verified, Valid: true},
	})
	
	return err
}
