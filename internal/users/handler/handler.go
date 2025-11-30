package handler

import (
	"context"

	userspb "rival/gen/proto/proto/api"
	"rival/internal/users/repo"
	"rival/internal/users/service"
)

type UserHandler struct {
	userspb.UnimplementedUserServiceServer
	service service.UserService
}

func NewUserHandler() (*UserHandler, error) {
	repository, err := repo.NewUserRepository()
	if err != nil {
		return nil, err
	}

	userService := service.NewUserService(repository)

	return &UserHandler{
		service: userService,
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userspb.GetUserRequest) (*userspb.GetUserResponse, error) {
	if req.UserId == 0 {
		return &userspb.GetUserResponse{User: nil}, nil
	}

	return h.service.GetUser(ctx, int(req.UserId))
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userspb.UpdateUserRequest) (*userspb.UpdateUserResponse, error) {
	if req.UserId == 0 {
		return &userspb.UpdateUserResponse{User: nil}, nil
	}

	params := service.UpdateUserParams{
		UserID:     int(req.UserId),
		Name:       req.Name,
		Phone:      req.Phone,
		ProfilePic: req.ProfilePic,
	}
	return h.service.UpdateUser(ctx, params)
}

func (h *UserHandler) GetUploadURL(ctx context.Context, req *userspb.GetUploadURLRequest) (*userspb.GetUploadURLResponse, error) {
	if req.UserId == 0 || req.FileName == "" {
		return &userspb.GetUploadURLResponse{UploadUrl: "", FileUrl: "", ExpiresIn: 0}, nil
	}
	return h.service.GetUploadURL(ctx, int(req.UserId), req.FileName, req.ContentType)
}

func (h *UserHandler) UpdateCoinBalance(ctx context.Context, req *userspb.UpdateCoinBalanceRequest) (*userspb.UpdateCoinBalanceResponse, error) {
	if req.UserId == 0 || req.Amount <= 0 {
		return &userspb.UpdateCoinBalanceResponse{NewBalance: 0}, nil
	}

	params := service.UpdateCoinBalanceParams{
		UserID:    int(req.UserId),
		Amount:    req.Amount,
		Operation: req.Operation,
	}
	return h.service.UpdateCoinBalance(ctx, params)
}

func (h *UserHandler) GetCoinBalance(ctx context.Context, req *userspb.GetCoinBalanceRequest) (*userspb.GetCoinBalanceResponse, error) {
	if req.UserId == 0 {
		return &userspb.GetCoinBalanceResponse{Balance: 0}, nil
	}
	return h.service.GetCoinBalance(ctx, int(req.UserId))
}

func (h *UserHandler) GetTransactionHistory(ctx context.Context, req *userspb.GetTransactionHistoryRequest) (*userspb.GetTransactionHistoryResponse, error) {
	if req.UserId == 0 {
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

	return h.service.GetTransactionHistory(ctx, int(req.UserId), page, limit)
}

func (h *UserHandler) GetReferralCode(ctx context.Context, req *userspb.GetReferralCodeRequest) (*userspb.GetReferralCodeResponse, error) {
	if req.UserId == 0 {
		return &userspb.GetReferralCodeResponse{ReferralCode: ""}, nil
	}
	return h.service.GetReferralCode(ctx, int(req.UserId))
}

func (h *UserHandler) ApplyReferralCode(ctx context.Context, req *userspb.ApplyReferralCodeRequest) (*userspb.ApplyReferralCodeResponse, error) {
	if req.UserId == 0 || req.ReferralCode == "" {
		return &userspb.ApplyReferralCodeResponse{Success: false, Message: "User ID and referral code required", RewardAmount: 0}, nil
	}
	return h.service.ApplyReferralCode(ctx, int(req.UserId), req.ReferralCode)
}
