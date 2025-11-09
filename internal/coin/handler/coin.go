package coinHandler

import (
	"context"

	"encore.app/internal/coin/repo"
	"encore.app/internal/coin/service"
)

var (
	coinRepo    = repo.New()
	coinService = service.New(coinRepo)
)

type GetBalanceResponse struct {
	Coins int64 `json:"coins"`
}

//encore:api public method=GET path=/coins/balance/:userID
func GetBalance(ctx context.Context, userID int64) (*GetBalanceResponse, error) {
	coins, err := coinService.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &GetBalanceResponse{Coins: coins}, nil
}

type EarnCoinsRequest struct {
	UserID int64  `json:"user_id"`
	Coins  int64  `json:"coins"`
	Reason string `json:"reason"`
}

type EarnCoinsResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/coins/earn
func EarnCoins(ctx context.Context, req *EarnCoinsRequest) (*EarnCoinsResponse, error) {
	if err := coinService.EarnCoins(ctx, req.UserID, req.Coins, req.Reason); err != nil {
		return nil, err
	}
	return &EarnCoinsResponse{Message: "Coins earned successfully"}, nil
}

type SpendCoinsRequest struct {
	UserID int64  `json:"user_id"`
	Coins  int64  `json:"coins"`
	Reason string `json:"reason"`
}

type SpendCoinsResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/coins/spend
func SpendCoins(ctx context.Context, req *SpendCoinsRequest) (*SpendCoinsResponse, error) {
	if err := coinService.SpendCoins(ctx, req.UserID, req.Coins, req.Reason); err != nil {
		return nil, err
	}
	return &SpendCoinsResponse{Message: "Coins spent successfully"}, nil
}

type Package struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Coins      int64   `json:"coins"`
	BonusCoins int64   `json:"bonus_coins"`
	Price      float64 `json:"price"`
}

type GetPackagesResponse struct {
	Packages []Package `json:"packages"`
}

//encore:api public method=GET path=/coins/packages
func GetPackages(ctx context.Context) (*GetPackagesResponse, error) {
	packages, err := coinService.GetPackages(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Package, len(packages))
	for i, pkg := range packages {
		result[i] = Package{
			ID:         pkg.ID,
			Name:       pkg.Name,
			Coins:      pkg.Coins,
			BonusCoins: pkg.BonusCoins,
			Price:      pkg.Price,
		}
	}
	return &GetPackagesResponse{Packages: result}, nil
}

type PurchaseRequest struct {
	UserID    int64 `json:"user_id"`
	PackageID int64 `json:"package_id"`
}

type PurchaseResponse struct {
	Message string `json:"message"`
}

type CoinTransaction struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Coins     int64  `json:"coins"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
}

type GetHistoryResponse struct {
	Transactions []CoinTransaction `json:"transactions"`
}

//encore:api public method=GET path=/coins/history/:userID/:page
func GetHistory(ctx context.Context, userID int64, page int) (*GetHistoryResponse, error) {
	transactions, err := coinService.GetHistory(ctx, userID, page)
	if err != nil {
		return nil, err
	}

	result := make([]CoinTransaction, len(transactions))
	for i, tx := range transactions {
		result[i] = CoinTransaction{
			ID:        tx.ID,
			UserID:    tx.UserID,
			Coins:     tx.Coins,
			Type:      tx.Type,
			Reason:    func() string { if tx.Reason != nil { return *tx.Reason }; return "" }(),
			CreatedAt: tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return &GetHistoryResponse{Transactions: result}, nil
}

type TransferCoinsRequest struct {
	FromUserID int64  `json:"from_user_id"`
	ToUserID   int64  `json:"to_user_id"`
	Coins      int64  `json:"coins"`
	Reason     string `json:"reason"`
}

type TransferCoinsResponse struct {
	Message       string `json:"message"`
	TransactionID int64  `json:"transaction_id"`
}

//encore:api public method=POST path=/coins/transfer
func TransferCoins(ctx context.Context, req *TransferCoinsRequest) (*TransferCoinsResponse, error) {
	txID, err := coinService.TransferCoins(ctx, req.FromUserID, req.ToUserID, req.Coins, req.Reason)
	if err != nil {
		return nil, err
	}

	return &TransferCoinsResponse{
		Message:       "Coins transferred successfully",
		TransactionID: txID,
	}, nil
}

type CoinStatsResponse struct {
	TotalCoins        int64 `json:"total_coins"`
	TotalEarned       int64 `json:"total_earned"`
	TotalSpent        int64 `json:"total_spent"`
	TotalTransactions int64 `json:"total_transactions"`
}

//encore:api public method=GET path=/coins/stats/:userID
func GetCoinStats(ctx context.Context, userID int64) (*CoinStatsResponse, error) {
	stats, err := coinService.GetCoinStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &CoinStatsResponse{
		TotalCoins:        stats.TotalCoins,
		TotalEarned:       stats.TotalEarned,
		TotalSpent:        stats.TotalSpent,
		TotalTransactions: stats.TotalTransactions,
	}, nil
}

type CreatePackageRequest struct {
	Name       string  `json:"name"`
	Coins      int64   `json:"coins"`
	BonusCoins int64   `json:"bonus_coins"`
	Price      float64 `json:"price"`
}

type CreatePackageResponse struct {
	Message string  `json:"message"`
	Package Package `json:"package"`
}

//encore:api public method=POST path=/coins/packages/create
func CreatePackage(ctx context.Context, req *CreatePackageRequest) (*CreatePackageResponse, error) {
	pkg, err := coinService.CreatePackage(ctx, req.Name, req.Coins, req.BonusCoins, req.Price)
	if err != nil {
		return nil, err
	}

	return &CreatePackageResponse{
		Message: "Package created successfully",
		Package: Package{
			ID:         pkg.ID,
			Name:       pkg.Name,
			Coins:      pkg.Coins,
			BonusCoins: pkg.BonusCoins,
			Price:      pkg.Price,
		},
	}, nil
}

//encore:api public method=POST path=/coins/purchase
func Purchase(ctx context.Context, req *PurchaseRequest) (*PurchaseResponse, error) {
	if err := coinService.PurchaseCoins(ctx, req.UserID, req.PackageID); err != nil {
		return nil, err
	}
	return &PurchaseResponse{Message: "Coins purchased successfully"}, nil
}
