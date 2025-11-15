package repo

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"encore.app/config"
	"encore.app/connection"
	schema "encore.app/gen/sql"
	"encore.app/pkg/tigerbeetle"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, params schema.CreateUserParams) (schema.User, error)
	GetUserByEmail(ctx context.Context, email string) (schema.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (schema.User, error)
	UpdateUser(ctx context.Context, params schema.UpdateUserParams) error
	UpdateUserPassword(ctx context.Context, params schema.UpdateUserPasswordParams) error
	CreateSession(ctx context.Context, params schema.CreateJWTSessionParams) error
	GetSession(ctx context.Context, tokenHash string) (schema.JwtSession, error)
	RevokeSession(ctx context.Context, tokenHash string) error
	StoreOTP(ctx context.Context, email, otp string, expiry time.Duration) error
	VerifyOTP(ctx context.Context, email, otp string) (bool, error)
}

type authRepository struct {
	db      *pgxpool.Pool
	queries *schema.Queries
	redis   *redis.Client
	tb      tigerbeetle.Service
}

func NewAuthRepository() (AuthRepository, error) {
	cfg := config.GetConfig()
	
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	
	redisClient := connection.GetRedisClient(&cfg.Redis)
	
	tbService, err := tigerbeetle.NewService()
	if err != nil {
		return nil, err
	}
	
	return &authRepository{
		db:      db,
		queries: schema.New(db),
		redis:   redisClient,
		tb:      tbService,
	}, nil
}

func (r *authRepository) CreateUser(ctx context.Context, params schema.CreateUserParams) (schema.User, error) {
	// Generate user-friendly referral code if not provided
	if !params.ReferralCode.Valid {
		code := generateUserFriendlyReferralCode(params.Name)
		params.ReferralCode = pgtype.Text{String: code, Valid: true}
	}

	// Create user in database
	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return schema.User{}, err
	}

	// Create TigerBeetle account based on user role
	userUUID, _ := dbUser.ID.Value()
	userID, _ := uuid.Parse(userUUID.(string))
	
	// Get user role for TigerBeetle account type
	var role string
	if dbUser.Role.Valid {
		role = string(dbUser.Role.UserRole)
	} else {
		role = "customer" // default
	}
	
	err = r.tb.CreateAccountByRole(userID, role)
	if err != nil {
		return schema.User{}, err
	}

	return dbUser, nil
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (schema.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *authRepository) GetUserByID(ctx context.Context, id uuid.UUID) (schema.User, error) {
	pgUUID := pgtype.UUID{}
	pgUUID.Scan(id)
	return r.queries.GetUserByID(ctx, pgUUID)
}

func (r *authRepository) UpdateUser(ctx context.Context, params schema.UpdateUserParams) error {
	return r.queries.UpdateUser(ctx, params)
}

func (r *authRepository) UpdateUserPassword(ctx context.Context, params schema.UpdateUserPasswordParams) error {
	return r.queries.UpdateUserPassword(ctx, params)
}

func (r *authRepository) CreateSession(ctx context.Context, params schema.CreateJWTSessionParams) error {
	return r.queries.CreateJWTSession(ctx, params)
}

func (r *authRepository) GetSession(ctx context.Context, tokenHash string) (schema.JwtSession, error) {
	return r.queries.GetJWTSession(ctx, tokenHash)
}

func (r *authRepository) RevokeSession(ctx context.Context, tokenHash string) error {
	return r.queries.RevokeJWTSession(ctx, tokenHash)
}

func (r *authRepository) StoreOTP(ctx context.Context, email, otp string, expiry time.Duration) error {
	key := "otp:" + email
	return r.redis.Set(ctx, key, otp, expiry).Err()
}

func (r *authRepository) VerifyOTP(ctx context.Context, email, otp string) (bool, error) {
	key := "otp:" + email
	storedOTP, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if storedOTP == otp {
		r.redis.Del(ctx, key)
		return true, nil
	}

	return false, nil
}

func generateUserFriendlyReferralCode(userName string) string {
	// Create user-friendly code: RIV + first 2 letters of name + 4 random chars
	namePrefix := strings.ToUpper(userName)
	if len(namePrefix) >= 2 {
		namePrefix = namePrefix[:2]
	} else {
		namePrefix = "US" // Default
	}
	
	// Generate 4 random alphanumeric characters
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomPart := make([]byte, 4)
	for i := range randomPart {
		randomPart[i] = chars[rand.Intn(len(chars))]
	}
	
	return "RIV" + namePrefix + string(randomPart)
}
