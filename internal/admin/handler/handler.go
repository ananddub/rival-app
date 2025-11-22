package handler

import (
	"context"

	adminpb "rival/gen/proto/proto/api"
	"rival/internal/admin/repo"
	"rival/internal/admin/service"
)

type AdminHandler struct {
	adminpb.UnimplementedAdminServiceServer
	service service.AdminService
}

func NewAdminHandler() (*AdminHandler, error) {
	repository, err := repo.NewAdminRepository()
	if err != nil {
		return nil, err
	}

	adminService := service.NewAdminService(repository)

	return &AdminHandler{
		service: adminService,
	}, nil
}

func (h *AdminHandler) GetDashboardStats(ctx context.Context, req *adminpb.GetAdminDashboardStatsRequest) (*adminpb.GetAdminDashboardStatsResponse, error) {
	return h.service.GetDashboardStats(ctx)
}

func (h *AdminHandler) GetAllMerchants(ctx context.Context, req *adminpb.GetAllMerchantsRequest) (*adminpb.GetAllMerchantsResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	return h.service.GetAllMerchants(ctx, req.Page, req.Limit)
}

func (h *AdminHandler) ApproveMerchant(ctx context.Context, req *adminpb.ApproveMerchantRequest) (*adminpb.ApproveMerchantResponse, error) {
	return &adminpb.ApproveMerchantResponse{
		Success: true,
	}, nil
}

func (h *AdminHandler) SuspendMerchant(ctx context.Context, req *adminpb.SuspendMerchantRequest) (*adminpb.SuspendMerchantResponse, error) {
	return &adminpb.SuspendMerchantResponse{
		Success: true,
	}, nil
}

func (h *AdminHandler) GetAllUsers(ctx context.Context, req *adminpb.GetAllUsersRequest) (*adminpb.GetAllUsersResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	return h.service.GetAllUsers(ctx, req.Page, req.Limit)
}

func (h *AdminHandler) SuspendUser(ctx context.Context, req *adminpb.SuspendUserRequest) (*adminpb.SuspendUserResponse, error) {
	return &adminpb.SuspendUserResponse{
		Success: true,
	}, nil
}

func (h *AdminHandler) GetAllTransactions(ctx context.Context, req *adminpb.GetAllTransactionsRequest) (*adminpb.GetAllTransactionsResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	return h.service.GetAllTransactions(ctx, req.Page, req.Limit)
}

func (h *AdminHandler) GetAuditLogs(ctx context.Context, req *adminpb.GetAuditLogsRequest) (*adminpb.GetAuditLogsResponse, error) {
	return &adminpb.GetAuditLogsResponse{
		TotalCount: 0,
	}, nil
}
