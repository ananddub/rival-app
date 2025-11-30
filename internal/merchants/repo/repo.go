package repo

import (
	"context"

	"rival/config"
	"rival/connection"
	schema "rival/gen/sql"
	"rival/pkg/tb"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MerchantRepository interface {
	CreateMerchant(ctx context.Context, params schema.CreateMerchantParams) (schema.Merchant, error)
	GetMerchantByID(ctx context.Context, id int) (schema.Merchant, error)
	GetMerchantByEmail(ctx context.Context, email string) (schema.Merchant, error)
	UpdateMerchant(ctx context.Context, params schema.UpdateMerchantParams) error
	ListActiveMerchants(ctx context.Context) ([]schema.Merchant, error)
	GetMerchantsByCategory(ctx context.Context, category string) ([]schema.Merchant, error)
	GetMerchantBalance(ctx context.Context, merchantID int) (float64, error)
	GetMerchantTransactions(ctx context.Context, merchantID int, limit, offset int32) ([]schema.Transaction, error)
	GetMerchantCustomers(ctx context.Context, merchantID int, limit, offset int32) ([]schema.User, error)
	GetMerchantAddresses(ctx context.Context, merchantID int) ([]schema.MerchantAddress, error)
	CreateMerchantAddress(ctx context.Context, params schema.CreateMerchantAddressParams) (schema.MerchantAddress, error)
	CreateOffer(ctx context.Context, params schema.CreateOfferParams) (schema.Offer, error)
	GetMerchantOffers(ctx context.Context, merchantID int, limit, offset int32) ([]schema.Offer, error)
	GetOfferByID(ctx context.Context, offerID int) (schema.Offer, error)
	UpdateOffer(ctx context.Context, params schema.UpdateOfferParams) error
}

type merchantRepository struct {
	db      *pgxpool.Pool
	queries *schema.Queries
	tb      *tb.TbService
}

func NewMerchantRepository() (MerchantRepository, error) {
	cfg := config.GetConfig()

	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	tbService, err := tb.NewService()
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
	err = r.tb.CreateMerchantAccount(int(merchant.ID))
	if err != nil {
		return schema.Merchant{}, err
	}

	return merchant, nil
}

func (r *merchantRepository) GetMerchantByID(ctx context.Context, id int) (schema.Merchant, error) {
	return r.queries.GetMerchantByID(ctx, int64(id))
}

func (r *merchantRepository) GetMerchantByEmail(ctx context.Context, email string) (schema.Merchant, error) {
	return r.queries.GetMerchantByEmail(ctx, email)
}
func (r *merchantRepository) UpdateMerchant(ctx context.Context, params schema.UpdateMerchantParams) error {
	return r.queries.UpdateMerchant(ctx, params)
}

func (r *merchantRepository) ListActiveMerchants(ctx context.Context) ([]schema.Merchant, error) {
	return r.queries.ListActiveMerchants(ctx)
}

func (r *merchantRepository) GetMerchantsByCategory(ctx context.Context, category string) ([]schema.Merchant, error) {
	return r.queries.GetMerchantsByCategory(ctx, pgtype.Text{String: category, Valid: true})
}

func (r *merchantRepository) GetMerchantBalance(ctx context.Context, merchantID int) (float64, error) {
	return r.tb.GetBalance(merchantID)
}

func (r *merchantRepository) GetMerchantTransactions(ctx context.Context, merchantID int, limit, offset int32) ([]schema.Transaction, error) {
	return r.queries.GetMerchantTransactions(ctx, schema.GetMerchantTransactionsParams{
		MerchantID: pgtype.Int8{Int64: int64(merchantID), Valid: true},
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *merchantRepository) GetMerchantCustomers(ctx context.Context, merchantID int, limit, offset int32) ([]schema.User, error) {
	return r.queries.GetMerchantCustomers(ctx, schema.GetMerchantCustomersParams{
		MerchantID: pgtype.Int8{Int64: int64(merchantID), Valid: true},
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *merchantRepository) GetMerchantAddresses(ctx context.Context, merchantID int) ([]schema.MerchantAddress, error) {
	return r.queries.GetMerchantAddresses(ctx, pgtype.Int8{Int64: int64(merchantID), Valid: true})
}

func (r *merchantRepository) CreateMerchantAddress(ctx context.Context, params schema.CreateMerchantAddressParams) (schema.MerchantAddress, error) {
	return r.queries.CreateMerchantAddress(ctx, params)
}

func (r *merchantRepository) CreateOffer(ctx context.Context, params schema.CreateOfferParams) (schema.Offer, error) {
	return r.queries.CreateOffer(ctx, params)
}

func (r *merchantRepository) GetMerchantOffers(ctx context.Context, merchantID int, limit, offset int32) ([]schema.Offer, error) {
	return r.queries.GetMerchantOffers(ctx, schema.GetMerchantOffersParams{
		MerchantID: pgtype.Int8{Int64: int64(merchantID), Valid: true},
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *merchantRepository) GetOfferByID(ctx context.Context, offerID int) (schema.Offer, error) {
	return r.queries.GetOfferByID(ctx, int64(offerID))
}

func (r *merchantRepository) UpdateOffer(ctx context.Context, params schema.UpdateOfferParams) error {
	return r.queries.UpdateOffer(ctx, params)
}
