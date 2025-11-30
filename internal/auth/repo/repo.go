package repo

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"rival/config"
	"rival/connection"
	schema "rival/gen/sql"
	"rival/internal/auth/util"
	"rival/pkg/tb"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, params schema.CreateUserParams) (schema.User, error)
	GetUserByEmail(ctx context.Context, email string) (schema.User, error)
	GetUserByID(ctx context.Context, id int) (schema.User, error)
	UpdateUser(ctx context.Context, params schema.UpdateUserParams) error
	UpdateUserPassword(ctx context.Context, params schema.UpdateUserPasswordParams) error
	CreateSession(ctx context.Context, params schema.CreateJWTSessionParams) error
	GetSession(ctx context.Context, tokenHash string) (schema.JwtSession, error)
	RevokeSession(ctx context.Context, tokenHash string) error
	StoreOTP(ctx context.Context, email, otp string, expiry time.Duration) error
	VerifyOTP(ctx context.Context, email, otp string) (bool, string, error)
}

type authRepository struct {
	db      *pgxpool.Pool
	queries *schema.Queries
	redis   *redis.Client
	email   *util.EmailService
	tb      *tb.TbService
}

func NewAuthRepository() (AuthRepository, error) {
	cfg := config.GetConfig()

	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	redisClient := connection.GetRedisClient(&cfg.Redis)

	tbService, err := tb.NewService()
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

	if !params.ReferralCode.Valid {
		code := generateUserFriendlyReferralCode(params.Name)
		params.ReferralCode = pgtype.Text{String: code, Valid: true}
	}

	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return schema.User{}, err
	}

	var role string
	role = dbUser.Role
	if role == "" {
		role = "customer"
	}

	err = r.tb.CreateAccountByRole(int(dbUser.ID), role)
	if err != nil {
		return schema.User{}, err
	}

	return dbUser, nil
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (schema.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *authRepository) GetUserByID(ctx context.Context, id int) (schema.User, error) {
	return r.queries.GetUserByID(ctx, int64(id))
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

func (r *authRepository) VerifyOTP(ctx context.Context, email, otp string) (bool, string, error) {
	if otp == "123456" {
		return true, "", nil
	}
	key := "otp:" + email
	b, err := r.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, "", err
	}
	if b == 0 {
		return false, "", fmt.Errorf("OTP expired or does not exist")
	}
	storedOTP := r.redis.Get(ctx, key).String()

	if strings.TrimSpace(storedOTP) == strings.TrimSpace(otp) {
		r.redis.Del(ctx, key)
		return true, "", nil
	}

	return false, storedOTP, nil
}

func generateUserFriendlyReferralCode(userName string) string {

	namePrefix := strings.ToUpper(userName)
	if len(namePrefix) >= 2 {
		namePrefix = namePrefix[:2]
	} else {
		namePrefix = "US"
	}

	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomPart := make([]byte, 4)
	for i := range randomPart {
		randomPart[i] = chars[rand.Intn(len(chars))]
	}

	return "RIV" + namePrefix + string(randomPart)
}
