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

//encore:api public method=POST path=/coins/purchase
func Purchase(ctx context.Context, req *PurchaseRequest) (*PurchaseResponse, error) {
	if err := coinService.PurchaseCoins(ctx, req.UserID, req.PackageID); err != nil {
		return nil, err
	}
	return &PurchaseResponse{Message: "Coins purchased successfully"}, nil
}
