package handler

import (
	"context"
	"fmt"

	orderpb "rival/gen/proto/proto/api"
	"rival/internal/orders/repo"
	"rival/internal/orders/service"
	"rival/internal/orders/util"
)

type OrderHandler struct {
	orderpb.UnimplementedOrderServiceServer
	service service.OrderService
	pubsub  util.OrderPubSubService
}

func NewOrderHandler() (*OrderHandler, error) {
	repository, err := repo.NewOrderRepository()
	if err != nil {
		return nil, err
	}

	orderService := service.NewOrderService(repository)
	pubsubService := util.NewOrderPubSubService()

	return &OrderHandler{
		service: orderService,
		pubsub:  pubsubService,
	}, nil
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	if req.Subtotal <= 0 {
		return nil, fmt.Errorf("subtotal must be greater than 0")
	}

	return h.service.CreateOrder(ctx, req)
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {

	return h.service.GetOrder(ctx, req)
}

func (h *OrderHandler) GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.GetUserOrdersResponse, error) {

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	return h.service.GetUserOrders(ctx, req)
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {

	return h.service.CancelOrder(ctx, req)
}

func (h *OrderHandler) StreamOrderUpdates(req *orderpb.StreamOrderUpdatesRequest, stream orderpb.OrderService_StreamOrderUpdatesServer) error {
	ch := h.pubsub.SubscribeOrderUpdates(int(req.UserId))
	defer ch.Close()

	for data := range ch.Receive() {
		if update, ok := data.(*orderpb.StreamOrderUpdatesResponse); ok {
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
	return nil
}
