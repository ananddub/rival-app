package walletHandler

import (
	"context"
	"strconv"

	"encore.app/internal/wallet/repo"
)

type CreateWalletRequest struct {
	UserID   string  `json:"user_id"`
	Balance  float64 `json:"balance"`
	Coins    int64   `json:"coins"`
	Currency string  `json:"currency"`
}

type WalletResponse struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	Balance  float64 `json:"balance"`
	Coins    int64   `json:"coins"`
	Currency string  `json:"currency"`
}

type CreateTransactionRequest struct {
	UserID      string  `json:"user_id"`
	WalletID    string  `json:"wallet_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Icon        string  `json:"icon"`
}

type TransactionResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	WalletID    string  `json:"wallet_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Icon        string  `json:"icon"`
}

type GetTransactionsResponse struct {
	Transactions []*TransactionResponse `json:"transactions"`
}

//encore:api public method=POST path=/wallet
func CreateWallet(ctx context.Context, req *CreateWalletRequest) (*WalletResponse, error) {
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)
	
	wallet, err := walletRepo.CreateWallet(ctx, userID, req.Balance, req.Coins, req.Currency)
	if err != nil {
		return nil, err
	}

	return &WalletResponse{
		ID:       strconv.FormatInt(wallet.ID, 10),
		UserID:   strconv.FormatInt(wallet.UserID, 10),
		Balance:  wallet.Balance,
		Coins:    wallet.Coins,
		Currency: wallet.Currency,
	}, nil
}

//encore:api public method=GET path=/wallet/user/:userID
func GetUserWallet(ctx context.Context, userID string) (*WalletResponse, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)
	
	wallet, err := walletRepo.GetWalletByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return &WalletResponse{
		ID:       strconv.FormatInt(wallet.ID, 10),
		UserID:   strconv.FormatInt(wallet.UserID, 10),
		Balance:  wallet.Balance,
		Coins:    wallet.Coins,
		Currency: wallet.Currency,
	}, nil
}

//encore:api public method=POST path=/transaction
func CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*TransactionResponse, error) {
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)
	walletID, _ := strconv.ParseInt(req.WalletID, 10, 64)
	
	transaction, err := walletRepo.CreateTransaction(ctx, userID, walletID, req.Title, req.Description, req.Amount, req.Type, req.Icon)
	if err != nil {
		return nil, err
	}

	return &TransactionResponse{
		ID:          strconv.FormatInt(transaction.ID, 10),
		UserID:      strconv.FormatInt(transaction.UserID, 10),
		WalletID:    strconv.FormatInt(transaction.WalletID, 10),
		Title:       transaction.Title,
		Description: transaction.Description,
		Amount:      transaction.Amount,
		Type:        transaction.Type,
		Icon:        transaction.Icon,
	}, nil
}

//encore:api public method=GET path=/transaction/user/:userID
func GetUserTransactions(ctx context.Context, userID string) (*GetTransactionsResponse, error) {
	uid, _ := strconv.ParseInt(userID, 10, 64)
	
	transactions, err := walletRepo.GetTransactionsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	var result []*TransactionResponse
	for _, transaction := range transactions {
		result = append(result, &TransactionResponse{
			ID:          strconv.FormatInt(transaction.ID, 10),
			UserID:      strconv.FormatInt(transaction.UserID, 10),
			WalletID:    strconv.FormatInt(transaction.WalletID, 10),
			Title:       transaction.Title,
			Description: transaction.Description,
			Amount:      transaction.Amount,
			Type:        transaction.Type,
			Icon:        transaction.Icon,
		})
	}

	return &GetTransactionsResponse{Transactions: result}, nil
}
