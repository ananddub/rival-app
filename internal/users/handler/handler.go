package handler

import (
	"context"

	userspb "encore.app/gen/proto/proto/api"
	"encore.app/internal/users/repo"
	"encore.app/internal/users/service"
	"encore.app/internal/users/util"
)

type UserHandler struct {
	userspb.UnimplementedUserServiceServer
	service service.UserService
	pubsub  util.UserPubSubService
}

func NewUserHandler() (*UserHandler, error) {
	repository, err := repo.NewUserRepository()
	if err != nil {
		return nil, err
	}

	userService := service.NewUserService(repository)
	pubsubService := util.NewUserPubSubService()

	return &UserHandler{
		service: userService,
		pubsub:  pubsubService,
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userspb.GetUserRequest) (*userspb.GetUserResponse, error) {
	if req.UserId == "" {
		return &userspb.GetUserResponse{User: nil}, nil
	}
	return h.service.GetUser(ctx, req.UserId)
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userspb.UpdateUserRequest) (*userspb.UpdateUserResponse, error) {
	if req.UserId == "" {
		return &userspb.UpdateUserResponse{User: nil}, nil
	}

	params := service.UpdateUserParams{
		UserID:     req.UserId,
		Name:       req.Name,
		Phone:      req.Phone,
		ProfilePic: req.ProfilePic,
	}
	return h.service.UpdateUser(ctx, params)
}

func (h *UserHandler) GetUploadURL(ctx context.Context, req *userspb.GetUploadURLRequest) (*userspb.GetUploadURLResponse, error) {
	if req.UserId == "" || req.FileName == "" {
		return &userspb.GetUploadURLResponse{UploadUrl: "", FileUrl: "", ExpiresIn: 0}, nil
	}
	return h.service.GetUploadURL(ctx, req.UserId, req.FileName, req.ContentType)
}

func (h *UserHandler) UpdateCoinBalance(ctx context.Context, req *userspb.UpdateCoinBalanceRequest) (*userspb.UpdateCoinBalanceResponse, error) {
	if req.UserId == "" || req.Amount <= 0 {
		return &userspb.UpdateCoinBalanceResponse{NewBalance: 0}, nil
	}

	params := service.UpdateCoinBalanceParams{
		UserID:    req.UserId,
		Amount:    req.Amount,
		Operation: req.Operation,
	}
	return h.service.UpdateCoinBalance(ctx, params)
}

func (h *UserHandler) GetCoinBalance(ctx context.Context, req *userspb.GetCoinBalanceRequest) (*userspb.GetCoinBalanceResponse, error) {
	if req.UserId == "" {
		return &userspb.GetCoinBalanceResponse{Balance: 0}, nil
	}
	return h.service.GetCoinBalance(ctx, req.UserId)
}

func (h *UserHandler) GetTransactionHistory(ctx context.Context, req *userspb.GetTransactionHistoryRequest) (*userspb.GetTransactionHistoryResponse, error) {
	if req.UserId == "" {
		return &userspb.GetTransactionHistoryResponse{Transactions: nil, TotalCount: 0}, nil
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	return h.service.GetTransactionHistory(ctx, req.UserId, page, limit)
}

func (h *UserHandler) GetReferralCode(ctx context.Context, req *userspb.GetReferralCodeRequest) (*userspb.GetReferralCodeResponse, error) {
	if req.UserId == "" {
		return &userspb.GetReferralCodeResponse{ReferralCode: ""}, nil
	}
	return h.service.GetReferralCode(ctx, req.UserId)
}

func (h *UserHandler) ApplyReferralCode(ctx context.Context, req *userspb.ApplyReferralCodeRequest) (*userspb.ApplyReferralCodeResponse, error) {
	if req.UserId == "" || req.ReferralCode == "" {
		return &userspb.ApplyReferralCodeResponse{Success: false, Message: "User ID and referral code required", RewardAmount: 0}, nil
	}
	return h.service.ApplyReferralCode(ctx, req.UserId, req.ReferralCode)
}
