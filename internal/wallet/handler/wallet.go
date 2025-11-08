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
