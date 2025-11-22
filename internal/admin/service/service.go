package service

import (
	"context"

	adminpb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	"rival/internal/admin/repo"
	"rival/pkg/utils"
)

type AdminService interface {
	GetDashboardStats(ctx context.Context) (*adminpb.GetAdminDashboardStatsResponse, error)
	GetAllMerchants(ctx context.Context, page, limit int32) (*adminpb.GetAllMerchantsResponse, error)
	GetAllUsers(ctx context.Context, page, limit int32) (*adminpb.GetAllUsersResponse, error)
	GetAllTransactions(ctx context.Context, page, limit int32) (*adminpb.GetAllTransactionsResponse, error)
}

type adminService struct {
	repo repo.AdminRepository
}

func NewAdminService(repo repo.AdminRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) GetDashboardStats(ctx context.Context) (*adminpb.GetAdminDashboardStatsResponse, error) {
	totalMerchants, _ := s.repo.GetTotalMerchants(ctx)
	activeMerchants, _ := s.repo.GetActiveMerchants(ctx)
	totalUsers, _ := s.repo.GetTotalUsers(ctx)
	totalVolume, _ := s.repo.GetTotalTransactionVolume(ctx)
	pendingApprovals, _ := s.repo.GetPendingMerchantApprovals(ctx)

	return &adminpb.GetAdminDashboardStatsResponse{
		TotalMerchants:           totalMerchants,
		ActiveMerchants:          activeMerchants,
		TotalUsers:               totalUsers,
		TotalTransactionVolume:   totalVolume,
		PendingMerchantApprovals: pendingApprovals,
	}, nil
}

func (s *adminService) GetAllMerchants(ctx context.Context, page, limit int32) (*adminpb.GetAllMerchantsResponse, error) {
	offset := (page - 1) * limit
	merchants, err := s.repo.GetAllMerchants(ctx, limit, offset)
	if err != nil {
		return &adminpb.GetAllMerchantsResponse{}, err
	}

	var protoMerchants []*schemapb.Merchant
	for _, merchant := range merchants {
		protoMerchants = append(protoMerchants, convertToProtoMerchant(merchant))
	}

	return &adminpb.GetAllMerchantsResponse{
		Merchants:  protoMerchants,
		TotalCount: int32(len(protoMerchants)),
	}, nil
}

func (s *adminService) GetAllUsers(ctx context.Context, page, limit int32) (*adminpb.GetAllUsersResponse, error) {
	offset := (page - 1) * limit
	users, err := s.repo.GetAllUsers(ctx, limit, offset)
	if err != nil {
		return &adminpb.GetAllUsersResponse{}, err
	}

	var protoUsers []*schemapb.User
	for _, user := range users {
		protoUsers = append(protoUsers, convertToProtoUser(user))
	}

	return &adminpb.GetAllUsersResponse{
		Users:      protoUsers,
		TotalCount: int32(len(protoUsers)),
	}, nil
}

func (s *adminService) GetAllTransactions(ctx context.Context, page, limit int32) (*adminpb.GetAllTransactionsResponse, error) {
	offset := (page - 1) * limit
	transactions, err := s.repo.GetAllTransactions(ctx, limit, offset)
	if err != nil {
		return &adminpb.GetAllTransactionsResponse{}, err
	}

	var protoTransactions []*schemapb.Transaction
	for _, tx := range transactions {
		protoTransactions = append(protoTransactions, convertToProtoTransaction(tx))
	}

	return &adminpb.GetAllTransactionsResponse{
		Transactions: protoTransactions,
		TotalCount:   int32(len(protoTransactions)),
	}, nil
}

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
	}
}

func convertToProtoUser(user schema.User) *schemapb.User {
	return &schemapb.User{
		Id:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Phone:        user.Phone.String,
		ProfilePic:   user.ProfilePic.String,
		CoinBalance:  utils.NumericToFloat64(user.CoinBalance),
		ReferralCode: user.ReferralCode.String,
		CreatedAt:    user.CreatedAt.Time.Unix(),
	}
}

func convertToProtoTransaction(tx schema.Transaction) *schemapb.Transaction {

	return &schemapb.Transaction{
		Id:              tx.ID,
		UserId:          tx.UserID.Int64,
		MerchantId:      tx.MerchantID.Int64,
		CoinsSpent:      utils.NumericToFloat64(tx.CoinsSpent),
		OriginalAmount:  utils.NumericToFloat64(tx.OriginalAmount),
		DiscountAmount:  utils.NumericToFloat64(tx.DiscountAmount),
		FinalAmount:     utils.NumericToFloat64(tx.FinalAmount),
		TransactionType: tx.TransactionType.String,
		Status:          tx.Status.String,
		CreatedAt:       tx.CreatedAt.Time.Unix(),
	}
}
