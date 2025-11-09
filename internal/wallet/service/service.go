package service

import (
	"context"
	"errors"

	walletInterface "encore.app/internal/interface/wallet"
	"encore.app/internal/wallet/utils"
)

type WalletService struct {
	repo walletInterface.Repository
}

func New(repo walletInterface.Repository) walletInterface.Service {
	return &WalletService{repo: repo}
}

func (s *WalletService) GetBalance(ctx context.Context, userID int64) (*walletInterface.Wallet, error) {
	return s.repo.GetWallet(ctx, userID)
}

func (s *WalletService) AddMoney(ctx context.Context, userID int64, amount float64) error {
	if err := utils.ValidateUserID(userID); err != nil {
		return err
	}
	if err := utils.ValidateAmount(amount); err != nil {
		return err
	}

	amount = utils.FormatAmount(amount)

	wallet, err := s.repo.GetWallet(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateBalance(ctx, userID, amount); err != nil {
		return err
	}

	desc := "Money added to wallet"
	icon := "💰"
	tx := &walletInterface.Transaction{
		UserID:      userID,
		WalletID:    wallet.ID,
		Title:       "Money Added",
		Description: &desc,
		Amount:      amount,
		Type:        "credit",
		Icon:        &icon,
	}
	_, err = s.repo.CreateTransaction(ctx, tx)
	return err
}

func (s *WalletService) DeductMoney(ctx context.Context, userID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	wallet, err := s.repo.GetWallet(ctx, userID)
	if err != nil {
		return err
	}

	if wallet.Balance < amount {
		return errors.New("insufficient balance")
	}

	if err := s.repo.UpdateBalance(ctx, userID, -amount); err != nil {
		return err
	}

	desc := "Money deducted from wallet"
	icon := "💸"
	tx := &walletInterface.Transaction{
		UserID:      userID,
		WalletID:    wallet.ID,
		Title:       "Money Deducted",
		Description: &desc,
		Amount:      amount,
		Type:        "debit",
		Icon:        &icon,
	}
	_, err = s.repo.CreateTransaction(ctx, tx)
	return err
}

func (s *WalletService) CreateWallet(ctx context.Context, userID int64) (*walletInterface.Wallet, error) {
	return s.repo.CreateWallet(ctx, userID)
}

func (s *WalletService) GetWalletStats(ctx context.Context, userID int64) (*walletInterface.WalletStats, error) {
	return s.repo.GetWalletStats(ctx, userID)
}

func (s *WalletService) Transfer(ctx context.Context, fromUserID, toUserID int64, amount float64, title, description string) (int64, error) {
	if err := utils.ValidateUserID(fromUserID); err != nil {
		return 0, err
	}
	if err := utils.ValidateUserID(toUserID); err != nil {
		return 0, err
	}
	if err := utils.ValidateAmount(amount); err != nil {
		return 0, err
	}
	if err := utils.ValidateTransactionTitle(title); err != nil {
		return 0, err
	}
	if err := utils.ValidateDescription(description); err != nil {
		return 0, err
	}

	amount = utils.FormatAmount(amount)

	if fromUserID == toUserID {
		return 0, errors.New("cannot transfer to same user")
	}

	// Check sender balance
	fromWallet, err := s.repo.GetWallet(ctx, fromUserID)
	if err != nil {
		return 0, err
	}

	if fromWallet.Balance < amount {
		return 0, errors.New("insufficient balance")
	}

	// Check receiver wallet exists
	_, err = s.repo.GetWallet(ctx, toUserID)
	if err != nil {
		return 0, errors.New("receiver wallet not found")
	}

	return s.repo.TransferMoney(ctx, fromUserID, toUserID, amount, title, description)
}

func (s *WalletService) GetHistory(ctx context.Context, userID int64, page int) ([]*walletInterface.Transaction, error) {
	if err := utils.ValidateUserID(userID); err != nil {
		return nil, err
	}
	if err := utils.ValidatePage(page); err != nil {
		return nil, err
	}

	limit := 20
	offset := (page - 1) * limit
	return s.repo.GetTransactions(ctx, userID, limit, offset)
}
