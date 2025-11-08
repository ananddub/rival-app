package coin

import (
	"context"
	"time"
)

type CoinPackage struct {
	ID         int64
	Name       string
	Coins      int64
	Price      float64
	BonusCoins int64
	IsActive   bool
	CreatedAt  time.Time
}

type CoinTransaction struct {
	ID        int64
	UserID    int64
	Coins     int64
	Type      string
	Reason    *string
	CreatedAt time.Time
}

type CoinPurchase struct {
	ID             int64
	UserID         int64
	PackageID      int64
	CoinsReceived  int64
	AmountPaid     float64
	PaymentStatus  string
	PaymentID      *string
	CreatedAt      time.Time
}

type Repository interface {
	GetUserCoins(ctx context.Context, userID int64) (int64, error)
	AddCoins(ctx context.Context, userID, coins int64) error
	DeductCoins(ctx context.Context, userID, coins int64) error
	
	CreateTransaction(ctx context.Context, tx *CoinTransaction) error
	GetTransactions(ctx context.Context, userID int64, limit, offset int) ([]*CoinTransaction, error)
	
	GetPackages(ctx context.Context) ([]*CoinPackage, error)
	CreatePurchase(ctx context.Context, purchase *CoinPurchase) (int64, error)
	UpdatePurchaseStatus(ctx context.Context, purchaseID int64, status string) error
}

type Service interface {
	GetBalance(ctx context.Context, userID int64) (int64, error)
	EarnCoins(ctx context.Context, userID int64, coins int64, reason string) error
	SpendCoins(ctx context.Context, userID int64, coins int64, reason string) error
	GetPackages(ctx context.Context) ([]*CoinPackage, error)
	PurchaseCoins(ctx context.Context, userID, packageID int64) error
}
