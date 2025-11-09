package walletHandler

import (
	"context"

	"encore.app/internal/wallet/repo"
	"encore.app/internal/wallet/service"
)

var (
	walletRepo    = repo.New()
	walletService = service.New(walletRepo)
)

type WalletResponse struct {
	ID       int64   `json:"id"`
	Balance  float64 `json:"balance"`
	Coins    int64   `json:"coins"`
	Currency string  `json:"currency"`
}

//encore:api public method=GET path=/wallet/balance/:userID
func GetWallet(ctx context.Context, userID int64) (*WalletResponse, error) {
	wallet, err := walletService.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &WalletResponse{
		ID:       wallet.ID,
		Balance:  wallet.Balance,
		Coins:    wallet.Coins,
		Currency: wallet.Currency,
	}, nil
}

type AddMoneyRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type AddMoneyResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/wallet/add
func AddMoney(ctx context.Context, req *AddMoneyRequest) (*AddMoneyResponse, error) {
	if err := walletService.AddMoney(ctx, req.UserID, req.Amount); err != nil {
		return nil, err
	}
	return &AddMoneyResponse{Message: "Money added successfully"}, nil
}

type DeductMoneyRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type DeductMoneyResponse struct {
	Message string `json:"message"`
}

//encore:api public method=POST path=/wallet/deduct
func DeductMoney(ctx context.Context, req *DeductMoneyRequest) (*DeductMoneyResponse, error) {
	if err := walletService.DeductMoney(ctx, req.UserID, req.Amount); err != nil {
		return nil, err
	}
	return &DeductMoneyResponse{Message: "Money deducted successfully"}, nil
}

type Transaction struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Icon        string  `json:"icon"`
	CreatedAt   string  `json:"created_at"`
}

type GetHistoryResponse struct {
	Transactions []Transaction `json:"transactions"`
}

type CreateWalletRequest struct {
	UserID int64 `json:"user_id"`
}

type CreateWalletResponse struct {
	Message string `json:"message"`
	Wallet  WalletResponse `json:"wallet"`
}

//encore:api public method=POST path=/wallet/create
func CreateWallet(ctx context.Context, req *CreateWalletRequest) (*CreateWalletResponse, error) {
	wallet, err := walletService.CreateWallet(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	return &CreateWalletResponse{
		Message: "Wallet created successfully",
		Wallet: WalletResponse{
			ID:       wallet.ID,
			Balance:  wallet.Balance,
			Coins:    wallet.Coins,
			Currency: wallet.Currency,
		},
	}, nil
}

type TransferRequest struct {
	FromUserID int64   `json:"from_user_id"`
	ToUserID   int64   `json:"to_user_id"`
	Amount     float64 `json:"amount"`
	Title      string  `json:"title"`
	Description string `json:"description,omitempty"`
}

type TransferResponse struct {
	Message       string `json:"message"`
	TransactionID int64  `json:"transaction_id"`
}

//encore:api public method=POST path=/wallet/transfer
func Transfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	txID, err := walletService.Transfer(ctx, req.FromUserID, req.ToUserID, req.Amount, req.Title, req.Description)
	if err != nil {
		return nil, err
	}

	return &TransferResponse{
		Message:       "Transfer completed successfully",
		TransactionID: txID,
	}, nil
}

type WalletStatsResponse struct {
	TotalBalance    float64 `json:"total_balance"`
	TotalCoins      int64   `json:"total_coins"`
	TotalTransactions int64 `json:"total_transactions"`
	LastTransaction *Transaction `json:"last_transaction,omitempty"`
}

//encore:api public method=GET path=/wallet/stats/:userID
func GetWalletStats(ctx context.Context, userID int64) (*WalletStatsResponse, error) {
	stats, err := walletService.GetWalletStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := &WalletStatsResponse{
		TotalBalance:      stats.TotalBalance,
		TotalCoins:        stats.TotalCoins,
		TotalTransactions: stats.TotalTransactions,
	}

	if stats.LastTransaction != nil {
		tx := stats.LastTransaction
		desc := ""
		if tx.Description != nil {
			desc = *tx.Description
		}
		icon := ""
		if tx.Icon != nil {
			icon = *tx.Icon
		}
		response.LastTransaction = &Transaction{
			ID:          tx.ID,
			Title:       tx.Title,
			Description: desc,
			Amount:      tx.Amount,
			Type:        tx.Type,
			Icon:        icon,
			CreatedAt:   tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return response, nil
}

//encore:api public method=GET path=/wallet/history/:userID/:page
func GetHistory(ctx context.Context, userID int64, page int) (*GetHistoryResponse, error) {
	txs, err := walletService.GetHistory(ctx, userID, page)
	if err != nil {
		return nil, err
	}

	result := make([]Transaction, len(txs))
	for i, tx := range txs {
		desc := ""
		if tx.Description != nil {
			desc = *tx.Description
		}
		icon := ""
		if tx.Icon != nil {
			icon = *tx.Icon
		}
		result[i] = Transaction{
			ID:          tx.ID,
			Title:       tx.Title,
			Description: desc,
			Amount:      tx.Amount,
			Type:        tx.Type,
			Icon:        icon,
			CreatedAt:   tx.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return &GetHistoryResponse{Transactions: result}, nil
}
