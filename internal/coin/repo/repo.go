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

func (r *CoinRepo) CreateTransaction(ctx context.Context, tx *coinInterface.CoinTransaction) (int64, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}

	queries := db.New(conn)
	result, err := queries.CreateCoinTransaction(ctx, db.CreateCoinTransactionParams{
		UserID: tx.UserID,
		Coins:  tx.Coins,
		Type:   tx.Type,
		Reason: pgtype.Text{String: *tx.Reason, Valid: tx.Reason != nil},
	})
	if err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (r *CoinRepo) GetCoinStats(ctx context.Context, userID int64) (*coinInterface.CoinStats, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	
	// Get current balance
	balance, err := r.GetUserCoins(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get all transactions
	allTxs, err := queries.GetAllCoinTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}

	var totalEarned, totalSpent int64
	for _, tx := range allTxs {
		if tx.Type == "earn" {
			totalEarned += tx.Coins
		} else if tx.Type == "spend" {
			totalSpent += tx.Coins
		}
	}

	return &coinInterface.CoinStats{
		TotalCoins:        balance,
		TotalEarned:       totalEarned,
		TotalSpent:        totalSpent,
		TotalTransactions: int64(len(allTxs)),
	}, nil
}

func (r *CoinRepo) TransferCoins(ctx context.Context, fromUserID, toUserID, coins int64, reason string) (int64, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return 0, err
	}

	// Start transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	queries := db.New(tx)

	// Deduct from sender
	err = queries.DeductCoins(ctx, db.DeductCoinsParams{
		UserID: fromUserID,
		Coins:  pgtype.Int8{Int64: coins, Valid: true},
	})
	if err != nil {
		return 0, err
	}

	// Add to receiver
	err = queries.AddCoins(ctx, db.AddCoinsParams{
		UserID: toUserID,
		Coins:  pgtype.Int8{Int64: coins, Valid: true},
	})
	if err != nil {
		return 0, err
	}

	// Create debit transaction for sender
	debitTx, err := queries.CreateCoinTransaction(ctx, db.CreateCoinTransactionParams{
		UserID: fromUserID,
		Coins:  coins,
		Type:   "transfer_out",
		Reason: pgtype.Text{String: reason, Valid: true},
	})
	if err != nil {
		return 0, err
	}

	// Create credit transaction for receiver
	_, err = queries.CreateCoinTransaction(ctx, db.CreateCoinTransactionParams{
		UserID: toUserID,
		Coins:  coins,
		Type:   "transfer_in",
		Reason: pgtype.Text{String: reason, Valid: true},
	})
	if err != nil {
		return 0, err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return debitTx.ID, nil
}

func (r *CoinRepo) CreatePackage(ctx context.Context, pkg *coinInterface.CoinPackage) (*coinInterface.CoinPackage, error) {
	cfg := config.GetConfig()
	conn, err := connection.GetPgConnection(&cfg.Database)
	if err != nil {
		return nil, err
	}

	queries := db.New(conn)
	result, err := queries.CreateCoinPackage(ctx, db.CreateCoinPackageParams{
		Name:       pkg.Name,
		Coins:      pkg.Coins,
		Price:      pgtype.Numeric{Int: big.NewInt(int64(pkg.Price * 100)), Valid: true},
		BonusCoins: pgtype.Int8{Int64: pkg.BonusCoins, Valid: true},
		IsActive:   pgtype.Bool{Bool: pkg.IsActive, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &coinInterface.CoinPackage{
		ID:         result.ID,
		Name:       result.Name,
		Coins:      result.Coins,
		Price:      float64(result.Price.Int.Int64()) / 100,
		BonusCoins: result.BonusCoins.Int64,
		IsActive:   result.IsActive.Bool,
		CreatedAt:  result.CreatedAt.Time,
	}, nil
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
