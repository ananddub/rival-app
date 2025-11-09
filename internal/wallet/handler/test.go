package walletHandler

import "context"

type TestWalletResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

//encore:api public method=GET path=/wallet/test
func TestWallet(ctx context.Context) (*TestWalletResponse, error) {
	return &TestWalletResponse{
		Message: "Wallet system is working properly",
		Status:  "success",
	}, nil
}

type WalletSummaryResponse struct {
	UserID            int64                `json:"user_id"`
	Wallet            WalletResponse       `json:"wallet"`
	Stats             WalletStatsResponse  `json:"stats"`
	RecentTransactions []Transaction       `json:"recent_transactions"`
}

//encore:api public method=GET path=/wallet/summary/:userID
func GetWalletSummary(ctx context.Context, userID int64) (*WalletSummaryResponse, error) {
	// Get wallet
	wallet, err := walletService.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get stats
	stats, err := walletService.GetWalletStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get recent transactions (first page)
	transactions, err := walletService.GetHistory(ctx, userID, 1)
	if err != nil {
		return nil, err
	}

	// Convert transactions
	recentTxs := make([]Transaction, 0)
	for i, tx := range transactions {
		if i >= 5 { // Only show 5 recent transactions
			break
		}
		desc := ""
		if tx.Description != nil {
			desc = *tx.Description
		}
		icon := ""
		if tx.Icon != nil {
			icon = *tx.Icon
		}
		recentTxs = append(recentTxs, Transaction{
			ID:          tx.ID,
			Title:       tx.Title,
			Description: desc,
			Amount:      tx.Amount,
			Type:        tx.Type,
			Icon:        icon,
			CreatedAt:   tx.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &WalletSummaryResponse{
		UserID: userID,
		Wallet: WalletResponse{
			ID:       wallet.ID,
			Balance:  wallet.Balance,
			Coins:    wallet.Coins,
			Currency: wallet.Currency,
		},
		Stats: WalletStatsResponse{
			TotalBalance:      stats.TotalBalance,
			TotalCoins:        stats.TotalCoins,
			TotalTransactions: stats.TotalTransactions,
		},
		RecentTransactions: recentTxs,
	}, nil
}
