package handler

import (
	"context"
	"time"

	merchantpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	"rival/internal/merchants/repo"
	"rival/internal/merchants/service"
	"rival/internal/merchants/util"
)

type MerchantHandler struct {
	merchantpb.UnimplementedMerchantServiceServer
	service service.MerchantService
	pubsub  util.MerchantPubSubService
}

func NewMerchantHandler() (*MerchantHandler, error) {
	repository, err := repo.NewMerchantRepository()
	if err != nil {
		return nil, err
	}

	merchantService := service.NewMerchantService(repository)
	pubsubService := util.NewMerchantPubSubService()

	return &MerchantHandler{
		service: merchantService,
		pubsub:  pubsubService,
	}, nil
}

func (h *MerchantHandler) GetMerchant(ctx context.Context, req *merchantpb.GetMerchantRequest) (*merchantpb.GetMerchantResponse, error) {

	return h.service.GetMerchant(ctx, int(req.MerchantId))
}

func (h *MerchantHandler) UpdateMerchant(ctx context.Context, req *merchantpb.UpdateMerchantRequest) (*merchantpb.UpdateMerchantResponse, error) {

	return h.service.UpdateMerchant(ctx, req)
}

func (h *MerchantHandler) GetMerchantAddress(ctx context.Context, req *merchantpb.GetMerchantAddressRequest) (*merchantpb.GetMerchantAddressResponse, error) {

	return h.service.GetMerchantAddress(ctx, int(req.MerchantId))
}
func (h *MerchantHandler) UpdateMerchantAddress(ctx context.Context, req *merchantpb.UpdateMerchantAddressRequest) (*merchantpb.UpdateMerchantAddressResponse, error) {
	// Create address record (simplified)
	address := &schemapb.MerchantAddress{
		Id:         req.MerchantId,
		MerchantId: req.MerchantId,
		Street:     req.Street,
		City:       req.City,
		State:      req.State,
		PostalCode: req.PostalCode,
		Country:    req.Country,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}

	return &merchantpb.UpdateMerchantAddressResponse{Address: address}, nil
}

func (h *MerchantHandler) GetOrders(ctx context.Context, req *merchantpb.GetOrdersRequest) (*merchantpb.GetOrdersResponse, error) {

	return h.service.GetOrders(ctx, req)
}

func (h *MerchantHandler) UpdateOrderStatus(ctx context.Context, req *merchantpb.UpdateOrderStatusRequest) (*merchantpb.UpdateOrderStatusResponse, error) {
	// Update order status (simplified)
	order := &schemapb.Order{
		Id:        req.OrderId,
		Status:    req.Status,
		UpdatedAt: time.Now().Unix(),
	}

	return &merchantpb.UpdateOrderStatusResponse{Order: order}, nil
}

func (h *MerchantHandler) GetCustomers(ctx context.Context, req *merchantpb.GetCustomersRequest) (*merchantpb.GetCustomersResponse, error) {

	return h.service.GetCustomers(ctx, req)
}

func (h *MerchantHandler) GetPayouts(ctx context.Context, req *merchantpb.GetPayoutsRequest) (*merchantpb.GetPayoutsResponse, error) {

	return h.service.GetPayouts(ctx, req)
}

func (h *MerchantHandler) CreateOffer(ctx context.Context, req *merchantpb.CreateOfferRequest) (*merchantpb.CreateOfferResponse, error) {

	return h.service.CreateOffer(ctx, req)
}

func (h *MerchantHandler) GetOffers(ctx context.Context, req *merchantpb.GetOffersRequest) (*merchantpb.GetOffersResponse, error) {

	return h.service.GetOffers(ctx, req)
}

func (h *MerchantHandler) UpdateOffer(ctx context.Context, req *merchantpb.UpdateOfferRequest) (*merchantpb.UpdateOfferResponse, error) {

	return h.service.UpdateOffer(ctx, req)
}

func (h *MerchantHandler) GetDashboardStats(ctx context.Context, req *merchantpb.GetDashboardStatsRequest) (*merchantpb.GetDashboardStatsResponse, error) {

	return h.service.GetDashboardStats(ctx, int(req.MerchantId))
}

func (h *MerchantHandler) StreamOrders(req *merchantpb.StreamOrdersRequest, stream merchantpb.MerchantService_StreamOrdersServer) error {
	ch := h.pubsub.SubscribeOrderUpdates(int(req.MerchantId))
	defer ch.Close()

	for data := range ch.Receive() {
		if update, ok := data.(*merchantpb.StreamOrdersResponse); ok {
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *MerchantHandler) StreamNotifications(req *merchantpb.StreamNotificationsRequest, stream merchantpb.MerchantService_StreamNotificationsServer) error {
	ch := h.pubsub.SubscribeMerchantNotifications(int(req.MerchantId))
	defer ch.Close()

	for data := range ch.Receive() {
		if notification, ok := data.(*merchantpb.StreamNotificationsResponse); ok {
			if err := stream.Send(notification); err != nil {
				return err
			}
		}
	}
	return nil
}
