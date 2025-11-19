package service

import (
	"context"
	"fmt"

	userspb "rival/gen/proto/proto/api"
	schemapb "rival/gen/proto/proto/schema"
	schema "rival/gen/sql"
	"rival/internal/users/repo"
	"rival/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UpdateUserParams struct {
	UserID     string
	Name       string
	Phone      string
	ProfilePic string
}

type UpdateCoinBalanceParams struct {
	UserID    string
	Amount    float64
	Operation string // add, subtract
}

type UserService interface {
	GetUser(ctx context.Context, userID string) (*userspb.GetUserResponse, error)
	UpdateUser(ctx context.Context, params UpdateUserParams) (*userspb.UpdateUserResponse, error)
	GetUploadURL(ctx context.Context, userID, fileName, contentType string) (*userspb.GetUploadURLResponse, error)
	UpdateCoinBalance(ctx context.Context, params UpdateCoinBalanceParams) (*userspb.UpdateCoinBalanceResponse, error)
	GetCoinBalance(ctx context.Context, userID string) (*userspb.GetCoinBalanceResponse, error)
	GetTransactionHistory(ctx context.Context, userID string, page, limit int32) (*userspb.GetTransactionHistoryResponse, error)
	GetCoinPurchaseHistory(ctx context.Context, userID string, page, limit int32) (*userspb.GetTransactionHistoryResponse, error)
	GetReferralHistory(ctx context.Context, userID string, page, limit int32) (*userspb.GetTransactionHistoryResponse, error)
	GetUserStats(ctx context.Context, userID string) (*userspb.GetUserResponse, error)
	GetReferralCode(ctx context.Context, userID string) (*userspb.GetReferralCodeResponse, error)
	ApplyReferralCode(ctx context.Context, userID, referralCode string) (*userspb.ApplyReferralCodeResponse, error)
}

type userService struct {
	repo repo.UserRepository
}

func NewUserService(repo repo.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUser(ctx context.Context, userID string) (*userspb.GetUserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	user, err := s.repo.GetUserProfile(ctx, id)
	if err != nil {
		return nil, err
	}

	protoUser := convertToProtoUser(user)
	
	// Generate signed URL for profile image if exists
	if user.ProfilePic.Valid && user.ProfilePic.String != "" {
		signedURL, err := s.generateProfileImageURL(ctx, userID, user.ProfilePic.String)
		if err == nil {
			protoUser.ProfilePic = signedURL
		}
	}

	return &userspb.GetUserResponse{
		User: protoUser,
	}, nil
}

func (s *userService) UpdateUser(ctx context.Context, params UpdateUserParams) (*userspb.UpdateUserResponse, error) {
	id, err := uuid.Parse(params.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	updateParams := schema.UpdateUserProfileParams{
		ID:         pgtype.UUID{},
		Name:       params.Name,
		Phone:      pgtype.Text{String: params.Phone, Valid: params.Phone != ""},
		ProfilePic: pgtype.Text{String: params.ProfilePic, Valid: params.ProfilePic != ""},
	}
	updateParams.ID.Scan(id)

	err = s.repo.UpdateUserProfile(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	// Get updated user
	user, err := s.repo.GetUserProfile(ctx, id)
	if err != nil {
		return nil, err
	}

	return &userspb.UpdateUserResponse{
		User: convertToProtoUser(user),
	}, nil
}

func (s *userService) GetUploadURL(ctx context.Context, userID, fileName, contentType string) (*userspb.GetUploadURLResponse, error) {
	uploadURL, fileURL, err := s.repo.GenerateUploadURL(ctx, userID, fileName, contentType)
	if err != nil {
		return nil, err
	}

	return &userspb.GetUploadURLResponse{
		UploadUrl: uploadURL,
		FileUrl:   fileURL,
		ExpiresIn: 3600, // 1 hour
	}, nil
}

func (s *userService) UpdateCoinBalance(ctx context.Context, params UpdateCoinBalanceParams) (*userspb.UpdateCoinBalanceResponse, error) {
	id, err := uuid.Parse(params.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Add coins using TigerBeetle service (only add operation supported)
	if params.Operation == "add" {
		err = s.repo.AddCoins(ctx, id, params.Amount)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("operation '%s' not supported, use payment service for spending", params.Operation)
	}

	// Get new balance
	newBalance, err := s.repo.GetCoinBalance(ctx, id)
	if err != nil {
		return nil, err
	}

	return &userspb.UpdateCoinBalanceResponse{
		NewBalance: newBalance,
	}, nil
}

func (s *userService) GetCoinBalance(ctx context.Context, userID string) (*userspb.GetCoinBalanceResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Get balance from TigerBeetle
	balance, err := s.repo.GetCoinBalance(ctx, id)
	if err != nil {
		return nil, err
	}

	return &userspb.GetCoinBalanceResponse{
		Balance: balance,
	}, nil
}

func (s *userService) GetTransactionHistory(ctx context.Context, userID string, page, limit int32) (*userspb.GetTransactionHistoryResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	offset := (page - 1) * limit
	transactions, err := s.repo.GetUserTransactions(ctx, id, limit, offset)
	if err != nil {
		return nil, err
	}

	var protoTransactions []*schemapb.Transaction
	for _, tx := range transactions {
		protoTransactions = append(protoTransactions, convertToProtoTransaction(tx))
	}

	return &userspb.GetTransactionHistoryResponse{
		Transactions: protoTransactions,
		TotalCount:   int32(len(protoTransactions)),
	}, nil
}

func (s *userService) GetReferralCode(ctx context.Context, userID string) (*userspb.GetReferralCodeResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	user, err := s.repo.GetUserProfile(ctx, id)
	if err != nil {
		return nil, err
	}

	return &userspb.GetReferralCodeResponse{
		ReferralCode: user.ReferralCode.String,
	}, nil
}

func (s *userService) ApplyReferralCode(ctx context.Context, userID, referralCode string) (*userspb.ApplyReferralCodeResponse, error) {
	// Get current user
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	currentUser, err := s.repo.GetUserProfile(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user already has a referrer
	if currentUser.ReferredBy.Valid {
		return &userspb.ApplyReferralCodeResponse{
			Success: false,
			Message: "You have already used a referral code",
		}, nil
	}

	// Find referrer by code
	referrer, err := s.repo.GetUserByReferralCode(ctx, referralCode)
	if err != nil {
		return &userspb.ApplyReferralCodeResponse{
			Success: false,
			Message: "Invalid referral code",
		}, nil
	}

	// Can't refer yourself
	referrerID, _ := referrer.ID.Value()
	if referrerID.(string) == userID {
		return &userspb.ApplyReferralCodeResponse{
			Success: false,
			Message: "You cannot use your own referral code",
		}, nil
	}

	// Create referral reward for referrer (signup bonus)
	createRewardParams := schema.CreateReferralRewardParams{
		ReferrerID:   referrer.ID,
		ReferredID:   currentUser.ID,
		RewardAmount: utils.Float64ToNumeric(5.0), // $5 referral bonus
		RewardType:   pgtype.Text{String: "signup", Valid: true},
		Status:       pgtype.Text{String: "pending", Valid: true},
	}

	err = s.repo.CreateReferralReward(ctx, createRewardParams)
	if err != nil {
		return nil, err
	}

	// Update current user's referred_by field
	updateParams := schema.UpdateUserProfileParams{
		ID:         pgtype.UUID{},
		Name:       currentUser.Name,
		Phone:      currentUser.Phone,
		ProfilePic: currentUser.ProfilePic,
	}
	updateParams.ID.Scan(id)

	err = s.repo.UpdateUserProfile(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	return &userspb.ApplyReferralCodeResponse{
		Success:      true,
		Message:      "Referral code applied successfully! You and your referrer will receive rewards.",
		RewardAmount: 5.0, // $5 referral bonus
	}, nil
}

func (s *userService) GetReferralRewards(ctx context.Context, userID string, page, limit int32) (*userspb.GetReferralRewardsResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	offset := (page - 1) * limit
	rewards, err := s.repo.GetUserReferralRewards(ctx, id, limit, offset)
	if err != nil {
		return nil, err
	}

	var protoRewards []*schemapb.ReferralReward
	var totalEarned float64

	for _, reward := range rewards {
		protoReward := convertToProtoReferralReward(reward)
		protoRewards = append(protoRewards, protoReward)
		
		// Add to total if credited
		if reward.Status.String == "credited" {
			if reward.RewardAmount.Valid {
				val, _ := reward.RewardAmount.Value()
				if val != nil {
					totalEarned += val.(float64)
				}
			}
		}
	}

	return &userspb.GetReferralRewardsResponse{
		Rewards:     protoRewards,
		TotalCount:  int32(len(protoRewards)),
		TotalEarned: totalEarned,
	}, nil
}

func convertToProtoUser(user schema.User) *schemapb.User {
	userID, _ := user.ID.Value()
	
	var coinBalance float64
	if user.CoinBalance.Valid {
		val, _ := user.CoinBalance.Value()
		if val != nil {
			coinBalance = val.(float64)
		}
	}
	
	var role schemapb.UserRole
	if user.Role.Valid {
		if val, ok := schemapb.UserRole_value[string(user.Role.UserRole)]; ok {
			role = schemapb.UserRole(val)
		}
	}
	
	var referredBy string
	if user.ReferredBy.Valid {
		referredByVal, _ := user.ReferredBy.Value()
		referredBy = referredByVal.(string)
	}
	
	return &schemapb.User{
		Id:           userID.(string),
		Email:        user.Email,
		PasswordHash: user.PasswordHash.String,
		Phone:        user.Phone.String,
		Name:         user.Name,
		ProfilePic:   user.ProfilePic.String,
		FirebaseUid:  user.FirebaseUid.String,
		CoinBalance:  coinBalance,
		Role:         role,
		ReferralCode: user.ReferralCode.String,
		ReferredBy:   referredBy,
		CreatedAt:    user.CreatedAt.Time.Unix(),
		UpdatedAt:    user.UpdatedAt.Time.Unix(),
	}
}

func convertToProtoReferralReward(reward schema.ReferralReward) *schemapb.ReferralReward {
	rewardID, _ := reward.ID.Value()
	referrerID, _ := reward.ReferrerID.Value()
	referredID, _ := reward.ReferredID.Value()
	
	var rewardAmount float64
	if reward.RewardAmount.Valid {
		val, _ := reward.RewardAmount.Value()
		if val != nil {
			rewardAmount = val.(float64)
		}
	}
	
	var creditedAt int64
	if reward.CreditedAt.Valid {
		creditedAt = reward.CreditedAt.Time.Unix()
	}
	
	return &schemapb.ReferralReward{
		Id:           rewardID.(string),
		ReferrerId:   referrerID.(string),
		ReferredId:   referredID.(string),
		RewardAmount: rewardAmount,
		RewardType:   reward.RewardType.String,
		Status:       reward.Status.String,
		CreditedAt:   creditedAt,
		CreatedAt:    reward.CreatedAt.Time.Unix(),
	}
}

func convertToProtoTransaction(tx schema.Transaction) *schemapb.Transaction {
	txID, _ := tx.ID.Value()
	userID, _ := tx.UserID.Value()
	merchantID, _ := tx.MerchantID.Value()
	
	// Handle numeric fields
	var coinsSpent, originalAmount, discountAmount, finalAmount float64
	if tx.CoinsSpent.Valid {
		val, _ := tx.CoinsSpent.Value()
		if val != nil {
			coinsSpent = val.(float64)
		}
	}
	if tx.OriginalAmount.Valid {
		val, _ := tx.OriginalAmount.Value()
		if val != nil {
			originalAmount = val.(float64)
		}
	}
	if tx.DiscountAmount.Valid {
		val, _ := tx.DiscountAmount.Value()
		if val != nil {
			discountAmount = val.(float64)
		}
	}
	if tx.FinalAmount.Valid {
		val, _ := tx.FinalAmount.Value()
		if val != nil {
			finalAmount = val.(float64)
		}
	}
	
	return &schemapb.Transaction{
		Id:               txID.(string),
		UserId:           userID.(string),
		MerchantId:       merchantID.(string),
		CoinsSpent:       coinsSpent,
		OriginalAmount:   originalAmount,
		DiscountAmount:   discountAmount,
		FinalAmount:      finalAmount,
		TransactionType:  tx.TransactionType.String,
		Status:           tx.Status.String,
		CreatedAt:        tx.CreatedAt.Time.Unix(),
	}
}
func (s *userService) GetCoinPurchaseHistory(ctx context.Context, userID string, page, limit int32) (*userspb.GetTransactionHistoryResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return &userspb.GetTransactionHistoryResponse{
			Transactions: nil,
			TotalCount:   0,
		}, nil
	}

	offset := (page - 1) * limit
	purchases, err := s.repo.GetUserCoinPurchases(ctx, id, limit, offset)
	if err != nil {
		return &userspb.GetTransactionHistoryResponse{
			Transactions: nil,
			TotalCount:   0,
		}, nil
	}

	// Convert purchases to transaction format
	var protoTransactions []*schemapb.Transaction
	for _, purchase := range purchases {
		protoTransactions = append(protoTransactions, convertPurchaseToTransaction(purchase))
	}

	return &userspb.GetTransactionHistoryResponse{
		Transactions: protoTransactions,
		TotalCount:   int32(len(purchases)),
	}, nil
}

func (s *userService) GetReferralHistory(ctx context.Context, userID string, page, limit int32) (*userspb.GetTransactionHistoryResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return &userspb.GetTransactionHistoryResponse{
			Transactions: nil,
			TotalCount:   0,
		}, nil
	}

	offset := (page - 1) * limit
	rewards, err := s.repo.GetUserReferralRewards(ctx, id, limit, offset)
	if err != nil {
		return &userspb.GetTransactionHistoryResponse{
			Transactions: nil,
			TotalCount:   0,
		}, nil
	}

	// Convert rewards to transaction format
	var protoTransactions []*schemapb.Transaction
	for _, reward := range rewards {
		protoTransactions = append(protoTransactions, convertRewardToTransaction(reward))
	}

	return &userspb.GetTransactionHistoryResponse{
		Transactions: protoTransactions,
		TotalCount:   int32(len(rewards)),
	}, nil
}

func (s *userService) GetUserStats(ctx context.Context, userID string) (*userspb.GetUserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return &userspb.GetUserResponse{
			User: nil,
		}, nil
	}

	user, err := s.repo.GetUserProfile(ctx, id)
	if err != nil {
		return &userspb.GetUserResponse{
			User: nil,
		}, nil
	}

	// Get coin balance
	balance, _ := s.repo.GetCoinBalance(ctx, id)

	protoUser := convertToProtoUser(user)
	protoUser.CoinBalance = balance
	
	// Generate signed URL for profile image if exists
	if user.ProfilePic.Valid && user.ProfilePic.String != "" {
		signedURL, err := s.generateProfileImageURL(ctx, userID, user.ProfilePic.String)
		if err == nil {
			protoUser.ProfilePic = signedURL
		}
	}

	return &userspb.GetUserResponse{
		User: protoUser,
	}, nil
}
// Conversion functions
func convertToCoinPurchase(purchase schema.CoinPurchase) *schemapb.CoinPurchase {
	purchaseID, _ := purchase.ID.Value()
	userID, _ := purchase.UserID.Value()
	
	return &schemapb.CoinPurchase{
		Id:            purchaseID.(string),
		UserId:        userID.(string),
		Amount:        utils.NumericToFloat64(purchase.Amount),
		CoinsReceived: utils.NumericToFloat64(purchase.CoinsReceived),
		PaymentMethod: purchase.PaymentMethod.String,
		Status:        purchase.Status.String,
		CreatedAt:     purchase.CreatedAt.Time.Unix(),
	}
}

func convertToReferralReward(reward schema.ReferralReward) *schemapb.ReferralReward {
	rewardID, _ := reward.ID.Value()
	referrerID, _ := reward.ReferrerID.Value()
	referredID, _ := reward.ReferredID.Value()
	
	return &schemapb.ReferralReward{
		Id:           rewardID.(string),
		ReferrerId:   referrerID.(string),
		ReferredId:   referredID.(string),
		RewardAmount: utils.NumericToFloat64(reward.RewardAmount),
		RewardType:   reward.RewardType.String,
		Status:       reward.Status.String,
		CreatedAt:    reward.CreatedAt.Time.Unix(),
	}
}
func convertPurchaseToTransaction(purchase schema.CoinPurchase) *schemapb.Transaction {
	purchaseID, _ := purchase.ID.Value()
	userID, _ := purchase.UserID.Value()
	
	return &schemapb.Transaction{
		Id:              purchaseID.(string),
		UserId:          userID.(string),
		MerchantId:      "", // No merchant for coin purchases
		CoinsSpent:      utils.NumericToFloat64(purchase.Amount),
		OriginalAmount:  utils.NumericToFloat64(purchase.Amount),
		TransactionType: "coin_purchase",
		Status:          purchase.Status.String,
		CreatedAt:       purchase.CreatedAt.Time.Unix(),
	}
}

func convertRewardToTransaction(reward schema.ReferralReward) *schemapb.Transaction {
	rewardID, _ := reward.ID.Value()
	referrerID, _ := reward.ReferrerID.Value()
	
	return &schemapb.Transaction{
		Id:              rewardID.(string),
		UserId:          referrerID.(string),
		MerchantId:      "", // No merchant for referral rewards
		CoinsSpent:      0,
		OriginalAmount:  utils.NumericToFloat64(reward.RewardAmount),
		TransactionType: "referral_reward",
		Status:          reward.Status.String,
		CreatedAt:       reward.CreatedAt.Time.Unix(),
	}
}
func (s *userService) generateProfileImageURL(ctx context.Context, userID, fileName string) (string, error) {
	// Generate signed URL for viewing the profile image
	return s.repo.GenerateViewURL(ctx, userID, fileName)
}
