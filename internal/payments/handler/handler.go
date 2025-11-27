package handler

import (
	"context"
	"fmt"

	paymentpb "rival/gen/proto/proto/api"
	"rival/internal/payments/repo"
	"rival/internal/payments/service"
	"rival/internal/payments/util"
)

type PaymentHandler struct {
	paymentpb.UnimplementedPaymentServiceServer
	service service.PaymentService
	pubsub  util.PaymentPubSubService
}

func NewPaymentHandler() (*PaymentHandler, error) {
	repo, err := repo.NewPaymentRepository()
	if err != nil {
		return nil, err
	}

	service := service.NewPaymentService(repo)
	pubsubService := util.NewPaymentPubSubService()

	return &PaymentHandler{
		service: service,
		pubsub:  pubsubService,
	}, nil
}

// Coin Purchase
func (h *PaymentHandler) InitiateCoinPurchase(ctx context.Context, req *paymentpb.InitiateCoinPurchaseRequest) (*paymentpb.InitiateCoinPurchaseResponse, error) {

	return h.service.InitiateCoinPurchase(ctx, req)
}

func (h *PaymentHandler) VerifyPayment(ctx context.Context, req *paymentpb.VerifyPaymentRequest) (*paymentpb.VerifyPaymentResponse, error) {
	if req.PaymentId == "" {
		return &paymentpb.VerifyPaymentResponse{Success: false}, nil
	}

	return h.service.VerifyPayment(ctx, req)
}

func (h *PaymentHandler) GetPaymentHistory(ctx context.Context, req *paymentpb.GetPaymentHistoryRequest) (*paymentpb.GetPaymentHistoryResponse, error) {
	if req.UserId == 0 {
		return &paymentpb.GetPaymentHistoryResponse{}, nil
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	return h.service.GetPaymentHistory(ctx, req)
}

func (h *PaymentHandler) RefundPayment(ctx context.Context, req *paymentpb.RefundPaymentRequest) (*paymentpb.RefundPaymentResponse, error) {
	if req.PaymentId == "" {
		return &paymentpb.RefundPaymentResponse{Success: false}, nil
	}

	return h.service.RefundPayment(ctx, req)
}

// Payment Transfers
func (h *PaymentHandler) PayToMerchant(ctx context.Context, req *paymentpb.PayToMerchantRequest) (*paymentpb.PayToMerchantResponse, error) {

	if req.Amount <= 0 {
		return &paymentpb.PayToMerchantResponse{Success: false}, nil
	}

	return h.service.PayToMerchant(ctx, req)
}

func (h *PaymentHandler) TransferToUser(ctx context.Context, req *paymentpb.TransferToUserRequest) (*paymentpb.TransferToUserResponse, error) {

	if req.Amount <= 0 {
		return &paymentpb.TransferToUserResponse{Success: false}, nil
	}

	return h.service.TransferToUser(ctx, req)
}

func (h *PaymentHandler) GetBalance(ctx context.Context, req *paymentpb.GetBalanceRequest) (*paymentpb.GetBalanceResponse, error) {
	return h.service.GetBalance(ctx, req)
}

func (h *PaymentHandler) GetTransactionHistory(ctx context.Context, req *paymentpb.GetTransactionHistoryRequest) (*paymentpb.GetTransactionHistoryResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	return h.service.GetTransactionHistory(ctx, req)
}

func (h *PaymentHandler) ProcessRefund(ctx context.Context, req *paymentpb.ProcessRefundRequest) (*paymentpb.ProcessRefundResponse, error) {
	if req.TransactionId == "" {
		return &paymentpb.ProcessRefundResponse{Success: false}, nil
	}

	if req.Amount <= 0 {
		return &paymentpb.ProcessRefundResponse{Success: false}, nil
	}

	return h.service.ProcessRefund(ctx, req)
}

func (h *PaymentHandler) GetFinancialHistory(ctx context.Context, req *paymentpb.GetFinancialHistoryRequest) (*paymentpb.GetFinancialHistoryResponse, error) {
	if req.UserId == 0 {
		return &paymentpb.GetFinancialHistoryResponse{}, nil
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	return h.service.GetFinancialHistory(ctx, req)
}

// Merchant Settlements
func (h *PaymentHandler) InitiateSettlement(ctx context.Context, req *paymentpb.InitiateSettlementRequest) (*paymentpb.InitiateSettlementResponse, error) {

	if req.Amount <= 0 {
		return &paymentpb.InitiateSettlementResponse{Success: false}, nil
	}

	return h.service.InitiateSettlement(ctx, req)
}

func (h *PaymentHandler) GetSettlements(ctx context.Context, req *paymentpb.GetSettlementsRequest) (*paymentpb.GetSettlementsResponse, error) {

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	return h.service.GetSettlements(ctx, req)
}

// Streaming methods
func (h *PaymentHandler) StreamPaymentUpdates(req *paymentpb.StreamPaymentUpdatesRequest, stream paymentpb.PaymentService_StreamPaymentUpdatesServer) error {
	ch := h.pubsub.SubscribePaymentUpdates(fmt.Sprintf("%d", req.UserId))
	defer ch.Close()

	for data := range ch.Receive() {
		if update, ok := data.(*paymentpb.StreamPaymentUpdatesResponse); ok {
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *PaymentHandler) StreamTransactionUpdates(req *paymentpb.StreamTransactionUpdatesRequest, stream paymentpb.PaymentService_StreamTransactionUpdatesServer) error {
	ch := h.pubsub.SubscribeTransactionUpdates(fmt.Sprintf("%d", req.UserId))
	defer ch.Close()

	for data := range ch.Receive() {
		if update, ok := data.(*paymentpb.StreamTransactionUpdatesResponse); ok {
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
	return nil
}
