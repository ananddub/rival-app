package wallet

import (
	"context"
	"time"
)

type Wallet struct {
	ID        int64
	UserID    int64
	Balance   float64
	Coins     int64
	Currency  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID          int64
	UserID      int64
	WalletID    int64
	Title       string
	Description *string
	Amount      float64
	Type        string
	Icon        *string
	CreatedAt   time.Time
}

type WalletStats struct {
	TotalBalance      float64
	TotalCoins        int64
	TotalTransactions int64
	LastTransaction   *Transaction
}

type Repository interface {
	GetWallet(ctx context.Context, userID int64) (*Wallet, error)
	CreateWallet(ctx context.Context, userID int64) (*Wallet, error)
	UpdateBalance(ctx context.Context, userID int64, amount float64) error
	
	CreateTransaction(ctx context.Context, tx *Transaction) (int64, error)
	GetTransactions(ctx context.Context, userID int64, limit, offset int) ([]*Transaction, error)
	GetWalletStats(ctx context.Context, userID int64) (*WalletStats, error)
	TransferMoney(ctx context.Context, fromUserID, toUserID int64, amount float64, title, description string) (int64, error)
}

type Service interface {
	GetBalance(ctx context.Context, userID int64) (*Wallet, error)
	CreateWallet(ctx context.Context, userID int64) (*Wallet, error)
	AddMoney(ctx context.Context, userID int64, amount float64) error
	DeductMoney(ctx context.Context, userID int64, amount float64) error
	GetHistory(ctx context.Context, userID int64, page int) ([]*Transaction, error)
	GetWalletStats(ctx context.Context, userID int64) (*WalletStats, error)
	Transfer(ctx context.Context, fromUserID, toUserID int64, amount float64, title, description string) (int64, error)
}
