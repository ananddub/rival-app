package service

import (
	"context"
	"fmt"
	"time"

	merchantpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	"rival/internal/merchants/repo"
	"rival/pkg/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type MerchantService interface {
	GetMerchant(ctx context.Context, merchantID int) (*merchantpb.GetMerchantResponse, error)
	UpdateMerchant(ctx context.Context, req *merchantpb.UpdateMerchantRequest) (*merchantpb.UpdateMerchantResponse, error)
	GetMerchantAddress(ctx context.Context, merchantID int) (*merchantpb.GetMerchantAddressResponse, error)
	GetOrders(ctx context.Context, req *merchantpb.GetOrdersRequest) (*merchantpb.GetOrdersResponse, error)
	GetCustomers(ctx context.Context, req *merchantpb.GetCustomersRequest) (*merchantpb.GetCustomersResponse, error)
	GetPayouts(ctx context.Context, req *merchantpb.GetPayoutsRequest) (*merchantpb.GetPayoutsResponse, error)
	CreateOffer(ctx context.Context, req *merchantpb.CreateOfferRequest) (*merchantpb.CreateOfferResponse, error)
	GetOffers(ctx context.Context, req *merchantpb.GetOffersRequest) (*merchantpb.GetOffersResponse, error)
	UpdateOffer(ctx context.Context, req *merchantpb.UpdateOfferRequest) (*merchantpb.UpdateOfferResponse, error)
	GetDashboardStats(ctx context.Context, merchantID int) (*merchantpb.GetDashboardStatsResponse, error)
}

type merchantService struct {
	repo repo.MerchantRepository
}

func NewMerchantService(repo repo.MerchantRepository) MerchantService {
	return &merchantService{repo: repo}
}

func (s *merchantService) GetMerchant(ctx context.Context, merchantID int) (*merchantpb.GetMerchantResponse, error) {
	merchant, err := s.repo.GetMerchantByID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	return &merchantpb.GetMerchantResponse{
		Merchant: convertToProtoMerchant(merchant),
	}, nil
}

func (s *merchantService) UpdateMerchant(ctx context.Context, req *merchantpb.UpdateMerchantRequest) (*merchantpb.UpdateMerchantResponse, error) {

	updateParams := schema.UpdateMerchantParams{
		ID:                 int64(req.MerchantId),
		Name:               req.Name,
		Phone:              pgtype.Text{String: req.Phone, Valid: req.Phone != ""},
		Category:           pgtype.Text{String: req.Category, Valid: req.Category != ""},
		DiscountPercentage: utils.Float64ToNumeric(req.DiscountPercentage),
		IsActive:           pgtype.Bool{Bool: true, Valid: true},
	}

	err := s.repo.UpdateMerchant(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update merchant: %w", err)
	}

	// Get updated merchant
	merchant, err := s.repo.GetMerchantByID(ctx, int(req.MerchantId))
	if err != nil {
		return nil, fmt.Errorf("failed to get updated merchant: %w", err)
	}

	return &merchantpb.UpdateMerchantResponse{
		Merchant: convertToProtoMerchant(merchant),
	}, nil
}

func (s *merchantService) GetMerchantAddress(ctx context.Context, merchantID int) (*merchantpb.GetMerchantAddressResponse, error) {
	addresses, err := s.repo.GetMerchantAddresses(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant addresses: %w", err)
	}

	var protoAddresses []*schemapb.MerchantAddress
	for _, addr := range addresses {
		protoAddresses = append(protoAddresses, convertToProtoMerchantAddress(addr))
	}

	return &merchantpb.GetMerchantAddressResponse{
		Addresses: protoAddresses,
	}, nil
}

func (s *merchantService) GetOrders(ctx context.Context, req *merchantpb.GetOrdersRequest) (*merchantpb.GetOrdersResponse, error) {
	// Get merchant transactions as "orders"
	transactions, err := s.repo.GetMerchantTransactions(ctx, int(req.MerchantId), req.Limit, (req.Page-1)*req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant transactions: %w", err)
	}

	// Convert transactions to orders (simplified)
	var orders []*schemapb.Order
	for _, tx := range transactions {
		orders = append(orders, convertTransactionToOrder(tx))
	}

	return &merchantpb.GetOrdersResponse{
		Orders:     orders,
		TotalCount: int32(len(orders)),
	}, nil
}

func (s *merchantService) GetCustomers(ctx context.Context, req *merchantpb.GetCustomersRequest) (*merchantpb.GetCustomersResponse, error) {
	// Get merchant customers
	customers, err := s.repo.GetMerchantCustomers(ctx, int(req.MerchantId), req.Limit, (req.Page-1)*req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant customers: %w", err)
	}

	// Convert to protobuf
	var protoCustomers []*schemapb.User
	for _, customer := range customers {
		protoCustomers = append(protoCustomers, convertToProtoUser(customer))
	}

	return &merchantpb.GetCustomersResponse{
		Customers:  protoCustomers,
		TotalCount: int32(len(protoCustomers)),
	}, nil
}

func (s *merchantService) GetPayouts(ctx context.Context, req *merchantpb.GetPayoutsRequest) (*merchantpb.GetPayoutsResponse, error) {
	// Get merchant balance as pending payout
	balance, err := s.repo.GetMerchantBalance(ctx, int(req.MerchantId))
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant balance: %w", err)
	}

	// Create a settlement record for current balance
	var payouts []*schemapb.Settlement
	if balance > 0 {
		payouts = append(payouts, &schemapb.Settlement{
			Id:               req.MerchantId,
			MerchantId:       req.MerchantId,
			SettlementAmount: balance,
			Status:           "pending",
			PeriodStart:      time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
			PeriodEnd:        time.Now().Format("2006-01-02"),
			CreatedAt:        time.Now().Unix(),
		})
	}

	return &merchantpb.GetPayoutsResponse{
		Payouts:    payouts,
		TotalCount: int32(len(payouts)),
	}, nil
}

func (s *merchantService) CreateOffer(ctx context.Context, req *merchantpb.CreateOfferRequest) (*merchantpb.CreateOfferResponse, error) {

	var validUntil pgtype.Timestamp
	if req.ValidUntil != 0 {
		validUntil = pgtype.Timestamp{Time: time.Unix(req.ValidUntil, 0), Valid: true}
	}

	createParams := schema.CreateOfferParams{
		MerchantID:         pgtype.Int8{Int64: int64(req.MerchantId), Valid: true},
		Title:              req.Title,
		Description:        pgtype.Text{String: req.Description, Valid: req.Description != ""},
		DiscountPercentage: utils.Float64ToNumeric(req.DiscountPercentage),
		MinAmount:          utils.Float64ToNumeric(req.MinAmount),
		MaxDiscount:        utils.Float64ToNumeric(req.MaxDiscount),
		ValidFrom:          pgtype.Timestamp{Time: time.Now(), Valid: true},
		ValidUntil:         validUntil,
	}

	offer, err := s.repo.CreateOffer(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	return &merchantpb.CreateOfferResponse{
		Offer: convertToProtoOffer(offer),
	}, nil
}

func (s *merchantService) GetOffers(ctx context.Context, req *merchantpb.GetOffersRequest) (*merchantpb.GetOffersResponse, error) {
	// Use default pagination values since they're not in the protobuf
	limit := int32(50)
	offset := int32(0)

	offers, err := s.repo.GetMerchantOffers(ctx, int(req.MerchantId), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant offers: %w", err)
	}

	var protoOffers []*schemapb.Offer
	for _, offer := range offers {
		protoOffers = append(protoOffers, convertToProtoOffer(offer))
	}

	return &merchantpb.GetOffersResponse{
		Offers: protoOffers,
	}, nil
}

func (s *merchantService) UpdateOffer(ctx context.Context, req *merchantpb.UpdateOfferRequest) (*merchantpb.UpdateOfferResponse, error) {
	var validUntil pgtype.Timestamp
	if req.ValidUntil != 0 {
		validUntil = pgtype.Timestamp{Time: time.Unix(req.ValidUntil, 0), Valid: true}
	}

	updateParams := schema.UpdateOfferParams{
		ID:                 int64(req.OfferId),
		Title:              req.Title,
		Description:        pgtype.Text{String: req.Description, Valid: req.Description != ""},
		DiscountPercentage: utils.Float64ToNumeric(req.DiscountPercentage),
		MinAmount:          utils.Float64ToNumeric(req.MinAmount),
		MaxDiscount:        utils.Float64ToNumeric(req.MaxDiscount),
		ValidFrom:          pgtype.Timestamp{Time: time.Now(), Valid: true},
		ValidUntil:         validUntil,
		IsActive:           pgtype.Bool{Bool: req.IsActive, Valid: true},
	}

	err := s.repo.UpdateOffer(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update offer: %w", err)
	}

	// Get updated offer
	updatedOffer, err := s.repo.GetOfferByID(ctx, int(req.OfferId))
	if err != nil {
		return nil, fmt.Errorf("failed to get updated offer: %w", err)
	}

	return &merchantpb.UpdateOfferResponse{
		Offer: convertToProtoOffer(updatedOffer),
	}, nil
}

func (s *merchantService) GetDashboardStats(ctx context.Context, merchantID int) (*merchantpb.GetDashboardStatsResponse, error) {
	balance, err := s.repo.GetMerchantBalance(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant balance: %w", err)
	}

	// Get merchant transactions for stats
	transactions, err := s.repo.GetMerchantTransactions(ctx, merchantID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant transactions: %w", err)
	}

	// Calculate today's stats
	todayRevenue := calculateTodayRevenue(transactions)
	todayOrders := countTodayOrders(transactions)

	// Get customer stats
	customers, err := s.repo.GetMerchantCustomers(ctx, merchantID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant customers: %w", err)
	}
	totalCustomers := int32(len(customers))
	newCustomers := countNewCustomers(customers)

	return &merchantpb.GetDashboardStatsResponse{
		TodayRevenue:   todayRevenue,
		TodayOrders:    todayOrders,
		NewCustomers:   newCustomers,
		TotalCustomers: totalCustomers,
		PendingPayout:  balance,
	}, nil
}

// Conversion functions
func convertToProtoMerchant(merchant schema.Merchant) *schemapb.Merchant {

	return &schemapb.Merchant{
		Id:                 merchant.ID,
		Name:               merchant.Name,
		Email:              merchant.Email,
		Phone:              merchant.Phone.String,
		Category:           merchant.Category.String,
		DiscountPercentage: utils.NumericToFloat64(merchant.DiscountPercentage),
		IsActive:           merchant.IsActive.Bool,
		CreatedAt:          merchant.CreatedAt.Time.Unix(),
		UpdatedAt:          merchant.UpdatedAt.Time.Unix(),
	}
}

func convertTransactionToOrder(tx schema.Transaction) *schemapb.Order {

	return &schemapb.Order{
		Id:          tx.ID,
		UserId:      tx.UserID.Int64,
		MerchantId:  tx.MerchantID.Int64,
		TotalAmount: utils.NumericToFloat64(tx.OriginalAmount),
		Status:      tx.Status.String,
		CreatedAt:   tx.CreatedAt.Time.Unix(),
	}
}

func convertToProtoUser(user schema.User) *schemapb.User {
	return &schemapb.User{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone.String,
		Role:      schemapb.UserRole_USER_ROLE_CUSTOMER,
		CreatedAt: user.CreatedAt.Time.Unix(),
		UpdatedAt: user.UpdatedAt.Time.Unix(),
	}
}

func convertToProtoMerchantAddress(addr schema.MerchantAddress) *schemapb.MerchantAddress {

	return &schemapb.MerchantAddress{
		Id:         addr.ID,
		MerchantId: addr.MerchantID.Int64,
		Street:     addr.Street.String,
		City:       addr.City.String,
		State:      addr.State.String,
		PostalCode: addr.PostalCode.String,
		Country:    addr.Country.String,
		Latitude:   utils.NumericToFloat64(addr.Latitude),
		Longitude:  utils.NumericToFloat64(addr.Longitude),
		IsPrimary:  addr.IsPrimary.Bool,
		CreatedAt:  addr.CreatedAt.Time.Unix(),
		UpdatedAt:  addr.UpdatedAt.Time.Unix(),
	}
}

func convertToProtoOffer(offer schema.Offer) *schemapb.Offer {
	var validUntil int64
	if offer.ValidUntil.Valid {
		validUntil = offer.ValidUntil.Time.Unix()
	}

	return &schemapb.Offer{
		Id:                 offer.ID,
		MerchantId:         offer.MerchantID.Int64,
		Title:              offer.Title,
		Description:        offer.Description.String,
		DiscountPercentage: utils.NumericToFloat64(offer.DiscountPercentage),
		MinAmount:          utils.NumericToFloat64(offer.MinAmount),
		MaxDiscount:        utils.NumericToFloat64(offer.MaxDiscount),
		IsActive:           offer.IsActive.Bool,
		ValidUntil:         validUntil,
		CreatedAt:          offer.CreatedAt.Time.Unix(),
		UpdatedAt:          offer.UpdatedAt.Time.Unix(),
	}
}

// Helper functions for dashboard stats
func calculateTodayRevenue(transactions []schema.Transaction) float64 {
	today := time.Now().Truncate(24 * time.Hour)
	var revenue float64

	for _, tx := range transactions {
		if tx.CreatedAt.Time.After(today) {
			revenue += utils.NumericToFloat64(tx.CoinsSpent)
		}
	}

	return revenue
}

func countTodayOrders(transactions []schema.Transaction) int32 {
	today := time.Now().Truncate(24 * time.Hour)
	var count int32

	for _, tx := range transactions {
		if tx.CreatedAt.Time.After(today) {
			count++
		}
	}

	return count
}

func countNewCustomers(customers []schema.User) int32 {
	today := time.Now().Truncate(24 * time.Hour)
	var count int32

	for _, customer := range customers {
		if customer.CreatedAt.Time.After(today) {
			count++
		}
	}

	return count
}
