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

type CoinStats struct {
	TotalCoins        int64
	TotalEarned       int64
	TotalSpent        int64
	TotalTransactions int64
}

type Repository interface {
	GetUserCoins(ctx context.Context, userID int64) (int64, error)
	AddCoins(ctx context.Context, userID, coins int64) error
	DeductCoins(ctx context.Context, userID, coins int64) error
	
	CreateTransaction(ctx context.Context, tx *CoinTransaction) (int64, error)
	GetTransactions(ctx context.Context, userID int64, limit, offset int) ([]*CoinTransaction, error)
	GetCoinStats(ctx context.Context, userID int64) (*CoinStats, error)
	TransferCoins(ctx context.Context, fromUserID, toUserID, coins int64, reason string) (int64, error)
	
	GetPackages(ctx context.Context) ([]*CoinPackage, error)
	CreatePackage(ctx context.Context, pkg *CoinPackage) (*CoinPackage, error)
	CreatePurchase(ctx context.Context, purchase *CoinPurchase) (int64, error)
	UpdatePurchaseStatus(ctx context.Context, purchaseID int64, status string) error
}

type Service interface {
	GetBalance(ctx context.Context, userID int64) (int64, error)
	EarnCoins(ctx context.Context, userID int64, coins int64, reason string) error
	SpendCoins(ctx context.Context, userID int64, coins int64, reason string) error
	GetHistory(ctx context.Context, userID int64, page int) ([]*CoinTransaction, error)
	GetCoinStats(ctx context.Context, userID int64) (*CoinStats, error)
	TransferCoins(ctx context.Context, fromUserID, toUserID, coins int64, reason string) (int64, error)
	GetPackages(ctx context.Context) ([]*CoinPackage, error)
	CreatePackage(ctx context.Context, name string, coins, bonusCoins int64, price float64) (*CoinPackage, error)
	PurchaseCoins(ctx context.Context, userID, packageID int64) error
}
