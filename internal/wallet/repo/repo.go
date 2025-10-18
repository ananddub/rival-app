package walletRepo

import (
	"context"

	"encore.app/config"
	"encore.app/connection"
	wallet_gen "encore.app/internal/wallet/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

type Wallet struct {
	ID       int64
	UserID   int64
	Balance  float64
	Coins    int64
	Currency string
}

type Transaction struct {
	ID          int64
	UserID      int64
	WalletID    int64
	Title       string
	Description string
	Amount      float64
	Type        string
	Icon        string
}

func CreateWallet(ctx context.Context, userID int64, balance float64, coins int64, currency string) (*Wallet, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := wallet_gen.New(db)
	
	wallet, err := queries.CreateWallet(ctx, wallet_gen.CreateWalletParams{
		UserID:   userID,
		Balance:  pgtype.Numeric{Int: nil, Exp: 0, NaN: false, InfinityModifier: 0, Valid: true},
		Coins:    pgtype.Int8{Int64: coins, Valid: true},
		Currency: pgtype.Text{String: currency, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &Wallet{
		ID:       wallet.ID,
		UserID:   wallet.UserID,
		Balance:  balance,
		Coins:    wallet.Coins.Int64,
		Currency: wallet.Currency.String,
	}, nil
}

func GetWalletByUserID(ctx context.Context, userID int64) (*Wallet, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := wallet_gen.New(db)
	
	wallet, err := queries.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		ID:       wallet.ID,
		UserID:   wallet.UserID,
		Balance:  0.0,
		Coins:    wallet.Coins.Int64,
		Currency: wallet.Currency.String,
	}, nil
}

func CreateTransaction(ctx context.Context, userID, walletID int64, title, description string, amount float64, transactionType, icon string) (*Transaction, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := wallet_gen.New(db)
	
	transaction, err := queries.CreateTransaction(ctx, wallet_gen.CreateTransactionParams{
		UserID:      userID,
		WalletID:    walletID,
		Title:       title,
		Description: pgtype.Text{String: description, Valid: true},
		Amount:      pgtype.Numeric{Int: nil, Exp: 0, NaN: false, InfinityModifier: 0, Valid: true},
		Type:        transactionType,
		Icon:        pgtype.Text{String: icon, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &Transaction{
		ID:          transaction.ID,
		UserID:      transaction.UserID,
		WalletID:    transaction.WalletID,
		Title:       transaction.Title,
		Description: transaction.Description.String,
		Amount:      amount,
		Type:        transaction.Type,
		Icon:        transaction.Icon.String,
	}, nil
}

func GetTransactionsByUserID(ctx context.Context, userID int64) ([]*Transaction, error) {
	cfg := config.GetConfig()
	db, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := wallet_gen.New(db)
	
	transactions, err := queries.GetTransactionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []*Transaction
	for _, transaction := range transactions {
		result = append(result, &Transaction{
			ID:          transaction.ID,
			UserID:      transaction.UserID,
			WalletID:    transaction.WalletID,
			Title:       transaction.Title,
			Description: transaction.Description.String,
			Amount:      0.0,
			Type:        transaction.Type,
			Icon:        transaction.Icon.String,
		})
	}

	return result, nil
}
