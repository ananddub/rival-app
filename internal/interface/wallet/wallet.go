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

type Repository interface {
	GetWallet(ctx context.Context, userID int64) (*Wallet, error)
	CreateWallet(ctx context.Context, userID int64) error
	UpdateBalance(ctx context.Context, userID int64, amount float64) error
	
	CreateTransaction(ctx context.Context, tx *Transaction) error
	GetTransactions(ctx context.Context, userID int64, limit, offset int) ([]*Transaction, error)
}

type Service interface {
	GetBalance(ctx context.Context, userID int64) (*Wallet, error)
	AddMoney(ctx context.Context, userID int64, amount float64) error
	DeductMoney(ctx context.Context, userID int64, amount float64) error
	GetHistory(ctx context.Context, userID int64, page int) ([]*Transaction, error)
}
