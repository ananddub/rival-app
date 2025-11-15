package repo

import (
	"context"

	"encore.app/config"
	"encore.app/connection"
	schema "encore.app/gen/sql"
	"encore.app/pkg/tigerbeetle"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MerchantRepository interface {
	CreateMerchant(ctx context.Context, params schema.CreateMerchantParams) (schema.Merchant, error)
	GetMerchantByID(ctx context.Context, id uuid.UUID) (schema.Merchant, error)
	GetMerchantByEmail(ctx context.Context, email string) (schema.Merchant, error)
}

type merchantRepository struct {
	db      *pgxpool.Pool
	queries *schema.Queries
	tb      tigerbeetle.Service
}

func NewMerchantRepository() (MerchantRepository, error) {
	cfg := config.GetConfig()
	
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	
	tbService, err := tigerbeetle.NewService()
	if err != nil {
		return nil, err
	}
	
	return &merchantRepository{
		db:      db,
		queries: schema.New(db),
		tb:      tbService,
	}, nil
}

func (r *merchantRepository) CreateMerchant(ctx context.Context, params schema.CreateMerchantParams) (schema.Merchant, error) {
	// Create merchant in database
	merchant, err := r.queries.CreateMerchant(ctx, params)
	if err != nil {
		return schema.Merchant{}, err
	}

	// Create TigerBeetle account for merchant
	merchantUUID, _ := merchant.ID.Value()
	merchantID, _ := uuid.Parse(merchantUUID.(string))
	
	err = r.tb.CreateMerchantAccount(merchantID)
	if err != nil {
		return schema.Merchant{}, err
	}

	return merchant, nil
}

func (r *merchantRepository) GetMerchantByID(ctx context.Context, id uuid.UUID) (schema.Merchant, error) {
	pgUUID := pgtype.UUID{}
	pgUUID.Scan(id)
	return r.queries.GetMerchantByID(ctx, pgUUID)
}

func (r *merchantRepository) GetMerchantByEmail(ctx context.Context, email string) (schema.Merchant, error) {
	return r.queries.GetMerchantByEmail(ctx, email)
}
