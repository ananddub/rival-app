package coinHandler

import "context"

type TestCoinResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

//encore:api public method=GET path=/coins/test
func TestCoin(ctx context.Context) (*TestCoinResponse, error) {
	return &TestCoinResponse{
		Message: "Coin system is working properly",
		Status:  "success",
	}, nil
}

type CoinDashboardResponse struct {
	Balance           int64             `json:"balance"`
	Stats             CoinStatsResponse `json:"stats"`
	RecentTransactions []CoinTransaction `json:"recent_transactions"`
	AvailablePackages []Package         `json:"available_packages"`
}

//encore:api public method=GET path=/coins/dashboard/:userID
func GetCoinDashboard(ctx context.Context, userID int64) (*CoinDashboardResponse, error) {
	// Get balance
	balance, err := coinService.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get stats
	stats, err := coinService.GetCoinStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get recent transactions
	transactions, err := coinService.GetHistory(ctx, userID, 1)
	if err != nil {
		return nil, err
	}

	// Get packages
	packages, err := coinService.GetPackages(ctx)
	if err != nil {
		return nil, err
	}

	// Convert transactions (limit to 5)
	recentTxs := make([]CoinTransaction, 0)
	for i, tx := range transactions {
		if i >= 5 {
			break
		}
		recentTxs = append(recentTxs, CoinTransaction{
			ID:        tx.ID,
			UserID:    tx.UserID,
			Coins:     tx.Coins,
			Type:      tx.Type,
			Reason:    func() string { if tx.Reason != nil { return *tx.Reason }; return "" }(),
			CreatedAt: tx.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// Convert packages
	availablePackages := make([]Package, len(packages))
	for i, pkg := range packages {
		availablePackages[i] = Package{
			ID:         pkg.ID,
			Name:       pkg.Name,
			Coins:      pkg.Coins,
			BonusCoins: pkg.BonusCoins,
			Price:      pkg.Price,
		}
	}

	return &CoinDashboardResponse{
		Balance: balance,
		Stats: CoinStatsResponse{
			TotalCoins:        stats.TotalCoins,
			TotalEarned:       stats.TotalEarned,
			TotalSpent:        stats.TotalSpent,
			TotalTransactions: stats.TotalTransactions,
		},
		RecentTransactions: recentTxs,
		AvailablePackages:  availablePackages,
	}, nil
}
