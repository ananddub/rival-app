package service

import (
	"context"
	"errors"

	coinInterface "encore.app/internal/interface/coin"
)

type CoinService struct {
	repo coinInterface.Repository
}

func New(repo coinInterface.Repository) coinInterface.Service {
	return &CoinService{repo: repo}
}

func (s *CoinService) GetBalance(ctx context.Context, userID int64) (int64, error) {
	return s.repo.GetUserCoins(ctx, userID)
}

func (s *CoinService) GetHistory(ctx context.Context, userID int64, page int) ([]*coinInterface.CoinTransaction, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	if page < 1 {
		return nil, errors.New("page must be positive")
	}

	limit := 20
	offset := (page - 1) * limit
	return s.repo.GetTransactions(ctx, userID, limit, offset)
}

func (s *CoinService) GetCoinStats(ctx context.Context, userID int64) (*coinInterface.CoinStats, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.repo.GetCoinStats(ctx, userID)
}

func (s *CoinService) TransferCoins(ctx context.Context, fromUserID, toUserID, coins int64, reason string) (int64, error) {
	if fromUserID <= 0 || toUserID <= 0 {
		return 0, errors.New("invalid user ID")
	}
	if coins <= 0 {
		return 0, errors.New("coins must be positive")
	}
	if fromUserID == toUserID {
		return 0, errors.New("cannot transfer to same user")
	}

	// Check sender balance
	balance, err := s.repo.GetUserCoins(ctx, fromUserID)
	if err != nil {
		return 0, err
	}
	if balance < coins {
		return 0, errors.New("insufficient coins")
	}

	return s.repo.TransferCoins(ctx, fromUserID, toUserID, coins, reason)
}

func (s *CoinService) CreatePackage(ctx context.Context, name string, coins, bonusCoins int64, price float64) (*coinInterface.CoinPackage, error) {
	if name == "" {
		return nil, errors.New("package name cannot be empty")
	}
	if coins <= 0 {
		return nil, errors.New("coins must be positive")
	}
	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	pkg := &coinInterface.CoinPackage{
		Name:       name,
		Coins:      coins,
		BonusCoins: bonusCoins,
		Price:      price,
		IsActive:   true,
	}

	return s.repo.CreatePackage(ctx, pkg)
}

func (s *CoinService) EarnCoins(ctx context.Context, userID int64, coins int64, reason string) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if coins <= 0 {
		return errors.New("coins must be positive")
	}

	if err := s.repo.AddCoins(ctx, userID, coins); err != nil {
		return err
	}

	tx := &coinInterface.CoinTransaction{
		UserID: userID,
		Coins:  coins,
		Type:   "earn",
		Reason: &reason,
	}
	_, err := s.repo.CreateTransaction(ctx, tx)
	return err
}

func (s *CoinService) SpendCoins(ctx context.Context, userID int64, coins int64, reason string) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if coins <= 0 {
		return errors.New("coins must be positive")
	}

	balance, err := s.repo.GetUserCoins(ctx, userID)
	if err != nil {
		return err
	}
	if balance < coins {
		return errors.New("insufficient coins")
	}

	if err := s.repo.DeductCoins(ctx, userID, coins); err != nil {
		return err
	}

	tx := &coinInterface.CoinTransaction{
		UserID: userID,
		Coins:  coins,
		Type:   "spend",
		Reason: &reason,
	}
	_, err = s.repo.CreateTransaction(ctx, tx)
	return err
}

func (s *CoinService) GetPackages(ctx context.Context) ([]*coinInterface.CoinPackage, error) {
	return s.repo.GetPackages(ctx)
}

func (s *CoinService) PurchaseCoins(ctx context.Context, userID, packageID int64) error {
	packages, err := s.repo.GetPackages(ctx)
	if err != nil {
		return err
	}

	var selectedPkg *coinInterface.CoinPackage
	for _, pkg := range packages {
		if pkg.ID == packageID {
			selectedPkg = pkg
			break
		}
	}
	if selectedPkg == nil {
		return errors.New("package not found")
	}

	totalCoins := selectedPkg.Coins + selectedPkg.BonusCoins

	purchase := &coinInterface.CoinPurchase{
		UserID:        userID,
		PackageID:     packageID,
		CoinsReceived: totalCoins,
		AmountPaid:    selectedPkg.Price,
		PaymentStatus: "pending",
	}

	purchaseID, err := s.repo.CreatePurchase(ctx, purchase)
	if err != nil {
		return err
	}

	// TODO: Integrate payment gateway
	// For now, mark as completed and add coins
	if err := s.repo.UpdatePurchaseStatus(ctx, purchaseID, "completed"); err != nil {
		return err
	}

	return s.repo.AddCoins(ctx, userID, totalCoins)
}
