package repo

import (
	"context"
	"math/big"

	"encore.app/config"
	"encore.app/connection"
	db "encore.app/gen"
	coinInterface "encore.app/internal/interface/coin"
	"github.com/jackc/pgx/v5/pgtype"
)

type CoinRepo struct{}

func New() coinInterface.Repository {
	return &CoinRepo{}
}

func (r *CoinRepo) GetUserCoins(ctx context.Context, userID int64) (int64, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}

	queries := db.New(conn)
	coins, err := queries.GetUserCoins(ctx, userID)
	if err != nil {
		return 0, err
	}
	return coins.Int64, nil
}

func (r *CoinRepo) AddCoins(ctx context.Context, userID, coins int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	return queries.AddCoins(ctx, db.AddCoinsParams{
		UserID: userID,
		Coins:  pgtype.Int8{Int64: coins, Valid: true},
	})
}

func (r *CoinRepo) DeductCoins(ctx context.Context, userID, coins int64) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	return queries.DeductCoins(ctx, db.DeductCoinsParams{
		UserID: userID,
		Coins:  pgtype.Int8{Int64: coins, Valid: true},
	})
}

func (r *CoinRepo) CreateTransaction(ctx context.Context, tx *coinInterface.CoinTransaction) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	return queries.CreateCoinTransaction(ctx, db.CreateCoinTransactionParams{
		UserID: tx.UserID,
		Coins:  tx.Coins,
		Type:   tx.Type,
		Reason: pgtype.Text{String: *tx.Reason, Valid: tx.Reason != nil},
	})
}

func (r *CoinRepo) GetTransactions(ctx context.Context, userID int64, limit, offset int) ([]*coinInterface.CoinTransaction, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	txs, err := queries.GetCoinTransactions(ctx, db.GetCoinTransactionsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	result := make([]*coinInterface.CoinTransaction, len(txs))
	for i, tx := range txs {
		var reason *string
		if tx.Reason.Valid {
			reason = &tx.Reason.String
		}
		result[i] = &coinInterface.CoinTransaction{
			ID:        tx.ID,
			UserID:    tx.UserID,
			Coins:     tx.Coins,
			Type:      tx.Type,
			Reason:    reason,
			CreatedAt: tx.CreatedAt.Time,
		}
	}
	return result, nil
}

func (r *CoinRepo) GetPackages(ctx context.Context) ([]*coinInterface.CoinPackage, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	packages, err := queries.GetCoinPackages(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*coinInterface.CoinPackage, len(packages))
	for i, pkg := range packages {
		result[i] = &coinInterface.CoinPackage{
			ID:         pkg.ID,
			Name:       pkg.Name,
			Coins:      pkg.Coins,
			Price:      float64(pkg.Price.Int.Int64()) / 100,
			BonusCoins: pkg.BonusCoins.Int64,
			IsActive:   pkg.IsActive.Bool,
			CreatedAt:  pkg.CreatedAt.Time,
		}
	}
	return result, nil
}

func (r *CoinRepo) CreatePurchase(ctx context.Context, purchase *coinInterface.CoinPurchase) (int64, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}

	queries := db.New(conn)
	result, err := queries.CreateCoinPurchase(ctx, db.CreateCoinPurchaseParams{
		UserID:        purchase.UserID,
		PackageID:     purchase.PackageID,
		CoinsReceived: purchase.CoinsReceived,
		AmountPaid:    pgtype.Numeric{Int: big.NewInt(int64(purchase.AmountPaid * 100)), Valid: true},
		PaymentStatus: pgtype.Text{String: purchase.PaymentStatus, Valid: true},
		PaymentID:     pgtype.Text{String: *purchase.PaymentID, Valid: purchase.PaymentID != nil},
	})
	if err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (r *CoinRepo) UpdatePurchaseStatus(ctx context.Context, purchaseID int64, status string) error {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return err
	}

	queries := db.New(conn)
	return queries.UpdatePurchaseStatus(ctx, db.UpdatePurchaseStatusParams{
		ID:            purchaseID,
		PaymentStatus: pgtype.Text{String: status, Valid: true},
	})
}
