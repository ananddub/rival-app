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

func (s *CoinService) EarnCoins(ctx context.Context, userID int64, coins int64, reason string) error {
	if err := s.repo.AddCoins(ctx, userID, coins); err != nil {
		return err
	}

	tx := &coinInterface.CoinTransaction{
		UserID: userID,
		Coins:  coins,
		Type:   "earn",
		Reason: &reason,
	}
	return s.repo.CreateTransaction(ctx, tx)
}

func (s *CoinService) SpendCoins(ctx context.Context, userID int64, coins int64, reason string) error {
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
	return s.repo.CreateTransaction(ctx, tx)
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
