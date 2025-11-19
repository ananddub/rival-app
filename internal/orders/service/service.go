package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	orderpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	"rival/internal/orders/repo"
	"rival/pkg/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type OrderService interface {
	CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error)
	GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error)
	GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.GetUserOrdersResponse, error)
	CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error)
}

type orderService struct {
	repo repo.OrderRepository
}

func NewOrderService(repo repo.OrderRepository) OrderService {
	return &orderService{repo: repo}
}

func (s *orderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	// Generate unique order number
	orderNumber := generateOrderNumber()

	// Calculate discount and total
	discountAmount := req.Subtotal * 0.15 // Default 15% discount
	totalAmount := req.Subtotal - discountAmount

	createParams := schema.CreateOrderParams{
		MerchantID:     pgtype.Int8{Int64: int64(req.MerchantId), Valid: true},
		UserID:         pgtype.Int8{Int64: int64(req.UserId), Valid: true},
		OfferID:        pgtype.Int8{Int64: int64(req.OfferId), Valid: true},
		OrderNumber:    orderNumber,
		Items:          []byte(req.Items),
		Subtotal:       utils.Float64ToNumeric(req.Subtotal),
		DiscountAmount: utils.Float64ToNumeric(discountAmount),
		TotalAmount:    utils.Float64ToNumeric(totalAmount),
		CoinsUsed:      utils.Float64ToNumeric(req.CoinsUsed),
		Status:         pgtype.Text{String: "pending", Valid: true},
		Notes:          pgtype.Text{String: req.Notes, Valid: req.Notes != ""},
	}

	order, err := s.repo.CreateOrder(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &orderpb.CreateOrderResponse{
		Order: convertToProtoOrder(order),
	}, nil
}

func (s *orderService) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {

	order, err := s.repo.GetOrderByID(ctx, int(req.OrderId))
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &orderpb.GetOrderResponse{
		Order: convertToProtoOrder(order),
	}, nil
}

func (s *orderService) GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.GetUserOrdersResponse, error) {
	userID := int(req.UserId)
	var orders []schema.Order
	if req.Status != "" {
		orders, _ = s.repo.GetUserOrdersByStatus(ctx, userID, req.Status, req.Limit, (req.Page-1)*req.Limit)
	} else {
		orders, _ = s.repo.GetUserOrders(ctx, userID, req.Limit, (req.Page-1)*req.Limit)
	}

	var protoOrders []*schemapb.Order
	for _, order := range orders {
		protoOrders = append(protoOrders, convertToProtoOrder(order))
	}

	return &orderpb.GetUserOrdersResponse{
		Orders:     protoOrders,
		TotalCount: int32(len(protoOrders)),
	}, nil
}

func (s *orderService) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {

	err := s.repo.UpdateOrderStatus(ctx, schema.UpdateOrderStatusParams{
		ID:     int64(req.OrderId),
		Status: pgtype.Text{String: "cancelled", Valid: true},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	return &orderpb.CancelOrderResponse{
		Success: true,
	}, nil
}

// Helper functions
func generateOrderNumber() string {
	timestamp := time.Now().Unix()
	return "ORD" + strconv.FormatInt(timestamp, 10)
}

func convertToProtoOrder(order schema.Order) *schemapb.Order {
	userID, _ := order.UserID.Value()
	merchantID, _ := order.MerchantID.Value()
	orderID := order.ID

	return &schemapb.Order{
		Id:             orderID,
		UserId:         userID.(int64),
		MerchantId:     merchantID.(int64),
		OfferId:        order.OfferID.Int64,
		OrderNumber:    order.OrderNumber,
		Items:          string(order.Items),
		Subtotal:       utils.NumericToFloat64(order.Subtotal),
		DiscountAmount: utils.NumericToFloat64(order.DiscountAmount),
		TotalAmount:    utils.NumericToFloat64(order.TotalAmount),
		CoinsUsed:      utils.NumericToFloat64(order.CoinsUsed),
		Status:         order.Status.String,
		Notes:          order.Notes.String,
		CreatedAt:      order.CreatedAt.Time.Unix(),
		UpdatedAt:      order.UpdatedAt.Time.Unix(),
	}
}
