package repo

import (
	"context"

	"rival/config"
	"rival/connection"
	schema "rival/gen/sql"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, params schema.CreateOrderParams) (schema.Order, error)
	GetOrderByID(ctx context.Context, id int) (schema.Order, error)
	GetOrderByNumber(ctx context.Context, orderNumber string) (schema.Order, error)
	UpdateOrderStatus(ctx context.Context, params schema.UpdateOrderStatusParams) error
	GetUserOrders(ctx context.Context, userID int, limit, offset int32) ([]schema.Order, error)
	GetUserOrdersByStatus(ctx context.Context, userID int, status string, limit, offset int32) ([]schema.Order, error)
	GetMerchantOrders(ctx context.Context, merchantID int, limit, offset int32) ([]schema.Order, error)
	GetMerchantOrdersByStatus(ctx context.Context, merchantID int, status string, limit, offset int32) ([]schema.Order, error)
	CountUserOrders(ctx context.Context, userID int) (int64, error)
	CountMerchantOrders(ctx context.Context, merchantID int) (int64, error)
}

type orderRepository struct {
	db      *pgxpool.Pool
	queries *schema.Queries
}

func NewOrderRepository() (OrderRepository, error) {
	cfg := config.GetConfig()

	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	return &orderRepository{
		db:      db,
		queries: schema.New(db),
	}, nil
}

func (r *orderRepository) CreateOrder(ctx context.Context, params schema.CreateOrderParams) (schema.Order, error) {
	return r.queries.CreateOrder(ctx, params)
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id int) (schema.Order, error) {
	return r.queries.GetOrderByID(ctx, int64(id))
}

func (r *orderRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (schema.Order, error) {
	return r.queries.GetOrderByNumber(ctx, orderNumber)
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, params schema.UpdateOrderStatusParams) error {
	return r.queries.UpdateOrderStatus(ctx, params)
}

func (r *orderRepository) GetUserOrders(ctx context.Context, userID int, limit, offset int32) ([]schema.Order, error) {
	return r.queries.GetUserOrders(ctx, schema.GetUserOrdersParams{
		UserID: pgtype.Int8{Int64: int64(userID), Valid: true},
		Limit:  limit,
		Offset: offset,
	})
}

func (r *orderRepository) GetUserOrdersByStatus(ctx context.Context, userID int, status string, limit, offset int32) ([]schema.Order, error) {
	return r.queries.GetUserOrdersByStatus(ctx, schema.GetUserOrdersByStatusParams{
		UserID: pgtype.Int8{Int64: int64(userID), Valid: true},
		Status: pgtype.Text{String: status, Valid: true},
		Limit:  limit,
		Offset: offset,
	})
}

func (r *orderRepository) GetMerchantOrders(ctx context.Context, merchantID int, limit, offset int32) ([]schema.Order, error) {
	return r.queries.GetMerchantOrders(ctx, schema.GetMerchantOrdersParams{
		MerchantID: pgtype.Int8{Int64: int64(merchantID), Valid: true},
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *orderRepository) GetMerchantOrdersByStatus(ctx context.Context, merchantID int, status string, limit, offset int32) ([]schema.Order, error) {

	return r.queries.GetMerchantOrdersByStatus(ctx, schema.GetMerchantOrdersByStatusParams{
		MerchantID: pgtype.Int8{Int64: int64(merchantID), Valid: true},
		Status:     pgtype.Text{String: status, Valid: true},
		Limit:      limit,
		Offset:     offset,
	})
}

func (r *orderRepository) CountUserOrders(ctx context.Context, userID int) (int64, error) {
	return r.queries.CountUserOrders(ctx, pgtype.Int8{Int64: int64(userID), Valid: true})
}

func (r *orderRepository) CountMerchantOrders(ctx context.Context, merchantID int) (int64, error) {
	return r.queries.CountMerchantOrders(ctx, pgtype.Int8{Int64: int64(merchantID), Valid: true})
}
