package repo

import (
	"context"
	"math/big"

	"encore.app/config"
	"encore.app/connection"
	db "encore.app/gen"
	walletInterface "encore.app/internal/interface/wallet"
	"github.com/jackc/pgx/v5/pgtype"
)

type WalletRepo struct{}

func New() walletInterface.Repository {
	return &WalletRepo{}
}

func (r *WalletRepo) GetWallet(ctx context.Context, userID int64) (*walletInterface.Wallet, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	wallet, err := queries.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	balance, _ := wallet.Balance.Float64Value()
	return &walletInterface.Wallet{
		ID:        wallet.ID,
		UserID:    wallet.UserID,
		Balance:   balance.Float64,
		Coins:     wallet.Coins.Int64,
		Currency:  wallet.Currency.String,
		CreatedAt: wallet.CreatedAt.Time,
		UpdatedAt: wallet.UpdatedAt.Time,
	}, nil
}

func (r *WalletRepo) CreateWallet(ctx context.Context, userID int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.CreateWallet(ctx, db.CreateWalletParams{
		UserID:   userID,
		Balance:  pgtype.Numeric{Int: big.NewInt(0), Valid: true},
		Coins:    pgtype.Int8{Int64: 0, Valid: true},
		Currency: pgtype.Text{String: "INR", Valid: true},
	})
	return err
}

func (r *WalletRepo) UpdateBalance(ctx context.Context, userID int64, amount float64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	wallet, err := r.GetWallet(ctx, userID)
	if err != nil {
		return err
	}

	newBalance := wallet.Balance + amount
	queries := db.New(conn)
	_, err = queries.UpdateBalance(ctx, db.UpdateBalanceParams{
		UserID:  userID,
		Balance: pgtype.Numeric{Int: big.NewInt(int64(newBalance * 100)), Valid: true},
	})
	return err
}

func (r *WalletRepo) CreateTransaction(ctx context.Context, tx *walletInterface.Transaction) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	_, err = queries.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:      tx.UserID,
		WalletID:    tx.WalletID,
		Title:       tx.Title,
		Description: pgtype.Text{String: *tx.Description, Valid: tx.Description != nil},
		Amount:      pgtype.Numeric{Int: big.NewInt(int64(tx.Amount * 100)), Valid: true},
		Type:        tx.Type,
		Icon:        pgtype.Text{String: *tx.Icon, Valid: tx.Icon != nil},
	})
	return err
}

func (r *WalletRepo) GetTransactions(ctx context.Context, userID int64, limit, offset int) ([]*walletInterface.Transaction, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	txs, err := queries.GetTransactionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	end := offset + limit
	if end > len(txs) {
		end = len(txs)
	}
	if offset >= len(txs) {
		return []*walletInterface.Transaction{}, nil
	}

	result := make([]*walletInterface.Transaction, 0)
	for i := offset; i < end; i++ {
		tx := txs[i]
		var desc, icon *string
		if tx.Description.Valid {
			desc = &tx.Description.String
		}
		if tx.Icon.Valid {
			icon = &tx.Icon.String
		}
		amount, _ := tx.Amount.Float64Value()
		result = append(result, &walletInterface.Transaction{
			ID:          tx.ID,
			UserID:      tx.UserID,
			WalletID:    tx.WalletID,
			Title:       tx.Title,
			Description: desc,
			Amount:      amount.Float64,
			Type:        tx.Type,
			Icon:        icon,
			CreatedAt:   tx.CreatedAt.Time,
		})
	}
	return result, nil
}
