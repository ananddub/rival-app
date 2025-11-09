package repo

import (
	"context"
	"time"

	"encore.app/config"
	"encore.app/connection"
	db "encore.app/gen"
	userInterface "encore.app/internal/interface/user"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepo struct{}

func New() userInterface.Repository {
	return &UserRepo{}
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*userInterface.User, error) {
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

	var phone, photo *string
	var dob *time.Time

	if user.PhoneNumber.Valid {
		phone = &user.PhoneNumber.String
	}
	if user.ProfilePhoto.Valid {
		photo = &user.ProfilePhoto.String
	}
	if user.Dob.Valid {
		dob = &user.Dob.Time
	}

	return &userInterface.User{
		ID:              user.ID,
		FullName:        user.FullName,
		Email:           user.Email,
		PhoneNumber:     phone,
		PasswordHash:    user.PasswordHash.String,
		ProfilePhoto:    photo,
		DOB:             dob,
		IsPhoneVerified: user.IsPhoneVerified.Bool,
		IsEmailVerified: user.IsEmailVerified.Bool,
		SignType:        user.SignType,
		Role:            user.Role,
		CreatedAt:       user.CreatedAt.Time,
		UpdatedAt:       user.UpdatedAt.Time,
	}, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, userID int64, fullName, phoneNumber, dob string) (*userInterface.User, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)

	// Get current user
	currentUser, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Use current values if new ones are empty
	if fullName == "" {
		fullName = currentUser.FullName
	}

	var phoneParam pgtype.Text
	if phoneNumber != "" {
		phoneParam = pgtype.Text{String: phoneNumber, Valid: true}
	} else if currentUser.PhoneNumber != nil {
		phoneParam = pgtype.Text{String: *currentUser.PhoneNumber, Valid: true}
	}

	var dobParam pgtype.Timestamp
	if dob != "" {
		dobTime, err := time.Parse("2006-01-02", dob)
		if err != nil {
			return nil, err
		}
		dobParam = pgtype.Timestamp{Time: dobTime, Valid: true}
	} else if currentUser.DOB != nil {
		dobParam = pgtype.Timestamp{Time: *currentUser.DOB, Valid: true}
	}

	user, err := queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:          userID,
		FullName:    fullName,
		PhoneNumber: phoneParam,
		Dob:         dobParam,
	})
	if err != nil {
		return nil, err
	}

	var phone, photo *string
	var dobPtr *time.Time

	if user.PhoneNumber.Valid {
		phone = &user.PhoneNumber.String
	}
	if user.ProfilePhoto.Valid {
		photo = &user.ProfilePhoto.String
	}
	if user.Dob.Valid {
		dobPtr = &user.Dob.Time
	}

	return &userInterface.User{
		ID:              user.ID,
		FullName:        user.FullName,
		Email:           user.Email,
		PhoneNumber:     phone,
		PasswordHash:    user.PasswordHash.String,
		ProfilePhoto:    photo,
		DOB:             dobPtr,
		IsPhoneVerified: user.IsPhoneVerified.Bool,
		IsEmailVerified: user.IsEmailVerified.Bool,
		SignType:        user.SignType,
		Role:            user.Role,
		CreatedAt:       user.CreatedAt.Time,
		UpdatedAt:       user.UpdatedAt.Time,
	}, nil
}

func (r *UserRepo) UpdateProfilePhoto(ctx context.Context, userID int64, photoPath string) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.UpdateProfilePhoto(ctx, db.UpdateProfilePhotoParams{
		ID:           userID,
		ProfilePhoto: pgtype.Text{String: photoPath, Valid: true},
	})
	return err
}

func (r *UserRepo) DeleteProfilePhoto(ctx context.Context, userID int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.UpdateProfilePhoto(ctx, db.UpdateProfilePhotoParams{
		ID:           userID,
		ProfilePhoto: pgtype.Text{Valid: false},
	})
	return err
}

func (r *UserRepo) VerifyEmail(ctx context.Context, userID int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.VerifyEmail(ctx, userID)
	return err
}

func (r *UserRepo) VerifyPhone(ctx context.Context, userID int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.VerifyPhone(ctx, userID)
	return err
}
