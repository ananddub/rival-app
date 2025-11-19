package repo

import (
	"context"
	"time"

	"rival/config"
	"rival/connection"
	schema "rival/gen/sql"
	"rival/pkg/tb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

type UserRepository interface {
	GetUserProfile(ctx context.Context, userID uuid.UUID) (schema.User, error)
	UpdateUserProfile(ctx context.Context, params schema.UpdateUserProfileParams) error
	GetCoinBalance(ctx context.Context, userID uuid.UUID) (float64, error)
	AddCoins(ctx context.Context, userID uuid.UUID, amount float64) error
	GetUserByReferralCode(ctx context.Context, referralCode string) (schema.User, error)
	GetUserTransactions(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]schema.Transaction, error)
	GetUserCoinPurchases(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]schema.CoinPurchase, error)
	GetUserReferralRewards(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]schema.ReferralReward, error)
	CreateReferralReward(ctx context.Context, params schema.CreateReferralRewardParams) error
	UpdateReferralRewardStatus(ctx context.Context, params schema.UpdateReferralRewardStatusParams) error
	GenerateUploadURL(ctx context.Context, userID, fileName, contentType string) (uploadURL, fileURL string, err error)
	GenerateViewURL(ctx context.Context, userID, fileName string) (string, error)
}

type userRepository struct {
	db      *pgxpool.Pool
	queries *schema.Queries
	redis   *redis.Client
	minio   *minio.Client
	tb      *tb.TbService
}

func NewUserRepository() (UserRepository, error) {
	cfg := config.GetConfig()

	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	redisClient := connection.GetRedisClient(&cfg.Redis)

	minioClient, err := connection.NewMinioClient()
	if err != nil {
		return nil, err
	}

	tbService, err := tb.NewService()
	if err != nil {
		return nil, err
	}

	return &userRepository{
		db:      db,
		queries: schema.New(db),
		redis:   redisClient,
		minio:   minioClient,
		tb:      tbService,
	}, nil
}

func (r *userRepository) GetUserProfile(ctx context.Context, userID uuid.UUID) (schema.User, error) {
	pgUUID := pgtype.UUID{}
	pgUUID.Scan(userID)
	return r.queries.GetUserProfile(ctx, pgUUID)
}

func (r *userRepository) UpdateUserProfile(ctx context.Context, params schema.UpdateUserProfileParams) error {
	return r.queries.UpdateUserProfile(ctx, params)
}

func (r *userRepository) GetCoinBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	return r.tb.GetBalance(userID)
}

func (r *userRepository) AddCoins(ctx context.Context, userID uuid.UUID, amount float64) error {
	return r.tb.AddCoins(userID, amount)
}

func (r *userRepository) GetUserByReferralCode(ctx context.Context, referralCode string) (schema.User, error) {
	return r.queries.GetUserByReferralCode(ctx, pgtype.Text{String: referralCode, Valid: true})
}

func (r *userRepository) GetUserTransactions(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]schema.Transaction, error) {
	pgUUID := pgtype.UUID{}
	pgUUID.Scan(userID)
	return r.queries.GetUserTransactions(ctx, schema.GetUserTransactionsParams{
		UserID: pgUUID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *userRepository) GetUserCoinPurchases(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]schema.CoinPurchase, error) {
	pgUUID := pgtype.UUID{}
	pgUUID.Scan(userID)
	return r.queries.GetUserCoinPurchases(ctx, schema.GetUserCoinPurchasesParams{
		UserID: pgUUID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *userRepository) GetUserReferralRewards(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]schema.ReferralReward, error) {
	pgUUID := pgtype.UUID{}
	pgUUID.Scan(userID)
	return r.queries.GetUserReferralRewards(ctx, schema.GetUserReferralRewardsParams{
		ReferrerID: pgUUID,
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *userRepository) CreateReferralReward(ctx context.Context, params schema.CreateReferralRewardParams) error {
	_, err := r.queries.CreateReferralReward(ctx, params)
	return err
}

func (r *userRepository) UpdateReferralRewardStatus(ctx context.Context, params schema.UpdateReferralRewardStatusParams) error {
	return r.queries.UpdateReferralRewardStatus(ctx, params)
}

func (r *userRepository) GenerateUploadURL(ctx context.Context, userID, fileName, contentType string) (uploadURL, fileURL string, err error) {
	cfg := config.GetConfig()
	objectName := "profiles/" + userID + "/" + fileName

	// Generate presigned URL for upload
	url, err := r.minio.PresignedPutObject(ctx, cfg.S3.BucketName, objectName, time.Hour)
	if err != nil {
		return "", "", err
	}

	uploadURL = url.String()
	fileURL = "https://" + cfg.S3.Endpoint + "/" + cfg.S3.BucketName + "/" + objectName

	return uploadURL, fileURL, nil
}
func (r *userRepository) GenerateViewURL(ctx context.Context, userID, fileName string) (string, error) {
	cfg := config.GetConfig()
	objectName := "profiles/" + userID + "/" + fileName

	// Generate presigned URL for viewing (GET request)
	url, err := r.minio.PresignedGetObject(ctx, cfg.S3.BucketName, objectName, time.Hour*24, nil)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}
