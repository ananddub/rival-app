package tb

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"rival/connection"

	tigerbeetle_go "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type Service interface {
	CreateUserAccount(userID int) error
	CreateMerchantAccount(merchantID int) error
	CreateAccountByRole(accountID int, role string) error
	GetBalance(accountID int) (float64, error)
	AddCoins(userID int, amount float64) error
	GetUser(userID int) (*[]types.Account, error)
	ProcessPayment(userID, merchantID int, amount float64) error
	Transfer(fromID, toID int, amount float64) error
	Close()
}

type TbService struct {
	client tigerbeetle_go.Client
}

func NewService() (*TbService, error) {
	client, err := connection.NewTbClient()
	if err != nil {
		return nil, err
	}

	return &TbService{client: client}, nil
}

func (s *TbService) GetUser(userID int) (*[]types.Account, error) {
	accounts, err := s.client.LookupAccounts([]types.Uint128{types.ToUint128(uint64(userID))})
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, nil
	}

	return &accounts, nil
}

func (s *TbService) CreateUserAccount(userID int) error {
	accountID := types.ToUint128(uint64(userID))
	account := types.Account{
		ID:     accountID,
		Ledger: 1,
		Code:   1, // User account
	}
	_, err := s.client.CreateAccounts([]types.Account{account})
	return err
}

func (s *TbService) CreateMerchantAccount(merchantID int) error {
	accountID := types.ToUint128(uint64(merchantID))
	account := types.Account{
		ID:     accountID,
		Ledger: 1,
		Code:   2, // Merchant account
	}
	_, err := s.client.CreateAccounts([]types.Account{account})
	return err
}

func (s *TbService) GetBalance(accountID int) (float64, error) {
	id := types.ToUint128(uint64(accountID))
	accounts, err := s.client.LookupAccounts([]types.Uint128{id})
	if err != nil {
		return 0, err
	}
	if len(accounts) == 0 {
		return 0, nil
	}
	balanceUint64 := uint64(accounts[0].CreditsPosted[0])
	return float64(balanceUint64) / 100, nil
}

func (s *TbService) AddCoins(userID int, amount float64) error {
	accountID := types.ToUint128(uint64(userID))
	transfer := types.Transfer{
		ID:              generateTransferID(),
		CreditAccountID: accountID,
		DebitAccountID:  types.ToUint128(1), // System account
		Amount:          types.ToUint128(uint64(amount * 100)),
		Ledger:          1,
		Code:            1, // Add coins
	}
	_, err := s.client.CreateTransfers([]types.Transfer{transfer})
	return err
}

func (s *TbService) ProcessPayment(userID, merchantID int, amount float64) error {
	userAccountID := types.ToUint128(uint64(userID))
	merchantAccountID := types.ToUint128(uint64(merchantID))
	transfer := types.Transfer{
		ID:              generateTransferID(),
		DebitAccountID:  userAccountID,
		CreditAccountID: merchantAccountID,
		Amount:          types.ToUint128(uint64(amount * 100)),
		Ledger:          1,
		Code:            2, // Payment to merchant
	}
	_, err := s.client.CreateTransfers([]types.Transfer{transfer})
	return err
}

func (s *TbService) Transfer(fromID, toID int, amount float64) error {
	fromAccountID := types.ToUint128(uint64(fromID))
	toAccountID := types.ToUint128(uint64(toID))
	transfer := types.Transfer{
		ID:              generateTransferID(),
		DebitAccountID:  fromAccountID,
		CreditAccountID: toAccountID,
		Amount:          types.ToUint128(uint64(amount * 100)),
		Ledger:          1,
		Code:            3, // Transfer
	}
	_, err := s.client.CreateTransfers([]types.Transfer{transfer})
	return err
}

func generateTransferID() types.Uint128 {
	now := time.Now().UnixNano()
	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	high := uint64(now)
	low := binary.BigEndian.Uint64(randBytes)
	return types.Uint128([16]uint8{
		uint8(high >> 56), uint8(high >> 48), uint8(high >> 40), uint8(high >> 32),
		uint8(high >> 24), uint8(high >> 16), uint8(high >> 8), uint8(high),
		uint8(low >> 56), uint8(low >> 48), uint8(low >> 40), uint8(low >> 32),
		uint8(low >> 24), uint8(low >> 16), uint8(low >> 8), uint8(low),
	})
}

func (s *TbService) CreateAccountByRole(accountID int, role string) error {
	id := types.ToUint128(uint64(accountID))
	var account types.Account

	switch role {
	case "customer":
		account = types.Account{
			ID:     id,
			Ledger: 1,
			Code:   1, // Customer account
		}
	case "merchant":
		account = types.Account{
			ID:     id,
			Ledger: 1,
			Code:   2, // Merchant account
		}
	case "admin":
		account = types.Account{
			ID:     id,
			Ledger: 1,
			Code:   3, // Admin account
		}
	default:
		// Default to customer
		account = types.Account{
			ID:     id,
			Ledger: 1,
			Code:   1, // Customer account
		}
	}

	_, err := s.client.CreateAccounts([]types.Account{account})
	return err
}

func (s *TbService) Close() {
	s.client.Close()
}
