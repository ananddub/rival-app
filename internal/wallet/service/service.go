package service

import (
	"context"
	"errors"

	walletInterface "encore.app/internal/interface/wallet"
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
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

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
	return s.repo.CreateTransaction(ctx, tx)
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
	return s.repo.CreateTransaction(ctx, tx)
}

func (s *WalletService) GetHistory(ctx context.Context, userID int64, page int) ([]*walletInterface.Transaction, error) {
	limit := 20
	offset := (page - 1) * limit
	return s.repo.GetTransactions(ctx, userID, limit, offset)
}
