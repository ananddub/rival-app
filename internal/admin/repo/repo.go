package repo

import (
	"context"

	schema "rival/gen/sql"
	"rival/config"
	"rival/connection"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository interface {
	GetTotalMerchants(ctx context.Context) (int32, error)
	GetActiveMerchants(ctx context.Context) (int32, error)
	GetTotalUsers(ctx context.Context) (int32, error)
	GetTotalTransactionVolume(ctx context.Context) (float64, error)
	GetPendingMerchantApprovals(ctx context.Context) (int32, error)
	GetAllMerchants(ctx context.Context, limit, offset int32) ([]schema.Merchant, error)
	GetAllUsers(ctx context.Context, limit, offset int32) ([]schema.User, error)
	GetAllTransactions(ctx context.Context, limit, offset int32) ([]schema.Transaction, error)
}

type adminRepository struct {
	db      *pgxpool.Pool
	queries *schema.Queries
}

func NewAdminRepository() (AdminRepository, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}
	queries := schema.New(db)
	
	return &adminRepository{
		db:      db,
		queries: queries,
	}, nil
}

func (r *adminRepository) GetTotalMerchants(ctx context.Context) (int32, error) {
	count, err := r.queries.CountMerchants(ctx)
	return int32(count), err
}

func (r *adminRepository) GetActiveMerchants(ctx context.Context) (int32, error) {
	count, err := r.queries.CountActiveMerchants(ctx)
	return int32(count), err
}

func (r *adminRepository) GetTotalUsers(ctx context.Context) (int32, error) {
	count, err := r.queries.CountUsers(ctx)
	return int32(count), err
}

func (r *adminRepository) GetTotalTransactionVolume(ctx context.Context) (float64, error) {
	return 0.0, nil // Implement with TigerBeetle query
}

func (r *adminRepository) GetPendingMerchantApprovals(ctx context.Context) (int32, error) {
	return 0, nil // Implement when merchant approval system is added
}

func (r *adminRepository) GetAllMerchants(ctx context.Context, limit, offset int32) ([]schema.Merchant, error) {
	return r.queries.GetAllMerchants(ctx, schema.GetAllMerchantsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *adminRepository) GetAllUsers(ctx context.Context, limit, offset int32) ([]schema.User, error) {
	return r.queries.GetAllUsers(ctx, schema.GetAllUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *adminRepository) GetAllTransactions(ctx context.Context, limit, offset int32) ([]schema.Transaction, error) {
	return r.queries.GetAllTransactions(ctx, schema.GetAllTransactionsParams{
		Limit:  limit,
		Offset: offset,
	})
}
