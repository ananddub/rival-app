package tigerbeetle

import (
	"crypto/rand"
	"encoding/binary"
	"strings"
	"time"

	"encore.app/connection"
	tigerbeetle_go "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"github.com/google/uuid"
)

type Service interface {
	CreateUserAccount(userID uuid.UUID) error
	CreateMerchantAccount(merchantID uuid.UUID) error
	CreateAccountByRole(accountID uuid.UUID, role string) error
	GetBalance(accountID uuid.UUID) (float64, error)
	AddCoins(userID uuid.UUID, amount float64) error
	ProcessPayment(userID, merchantID uuid.UUID, amount float64) error
	Transfer(fromID, toID uuid.UUID, amount float64) error
}

type service struct {
	client tigerbeetle_go.Client
}

func NewService() (Service, error) {
	client, err := connection.NewTbClient()
	if err != nil {
		return nil, err
	}
	
	return &service{client: client}, nil
}

func (s *service) CreateUserAccount(userID uuid.UUID) error {
	accountID := uuidToUint128(userID.String())
	account := types.Account{
		ID:     accountID,
		Ledger: 1,
		Code:   1, // User account
	}
	_, err := s.client.CreateAccounts([]types.Account{account})
	return err
}

func (s *service) CreateMerchantAccount(merchantID uuid.UUID) error {
	accountID := uuidToUint128(merchantID.String())
	account := types.Account{
		ID:     accountID,
		Ledger: 1,
		Code:   2, // Merchant account
	}
	_, err := s.client.CreateAccounts([]types.Account{account})
	return err
}

func (s *service) GetBalance(accountID uuid.UUID) (float64, error) {
	id := uuidToUint128(accountID.String())
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

func (s *service) AddCoins(userID uuid.UUID, amount float64) error {
	accountID := uuidToUint128(userID.String())
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

func (s *service) ProcessPayment(userID, merchantID uuid.UUID, amount float64) error {
	userAccountID := uuidToUint128(userID.String())
	merchantAccountID := uuidToUint128(merchantID.String())
	transfer := types.Transfer{
		ID:              generateTransferID(),
		DebitAccountID:  userAccountID,
		CreditAccountID: merchantAccountID,
		Amount:          types.ToUint128(uint64(amount * 100)),
		Ledger:          1,
		Code:            4, // Payment
	}
	_, err := s.client.CreateTransfers([]types.Transfer{transfer})
	return err
}

func (s *service) Transfer(fromID, toID uuid.UUID, amount float64) error {
	fromAccountID := uuidToUint128(fromID.String())
	toAccountID := uuidToUint128(toID.String())
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

// Utility functions
func uuidToUint128(uuidStr string) types.Uint128 {
	cleanUUID := strings.ReplaceAll(uuidStr, "-", "")
	bytes := make([]byte, 16)
	for i := 0; i < 32; i += 2 {
		val := hexToByte(cleanUUID[i:i+2])
		bytes[i/2] = val
	}
	high := binary.BigEndian.Uint64(bytes[:8])
	low := binary.BigEndian.Uint64(bytes[8:])
	return types.Uint128([16]uint8{
		uint8(high >> 56), uint8(high >> 48), uint8(high >> 40), uint8(high >> 32),
		uint8(high >> 24), uint8(high >> 16), uint8(high >> 8), uint8(high),
		uint8(low >> 56), uint8(low >> 48), uint8(low >> 40), uint8(low >> 32),
		uint8(low >> 24), uint8(low >> 16), uint8(low >> 8), uint8(low),
	})
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

func hexToByte(hex string) byte {
	var result byte
	for _, char := range hex {
		result <<= 4
		if char >= '0' && char <= '9' {
			result += byte(char - '0')
		} else if char >= 'a' && char <= 'f' {
			result += byte(char - 'a' + 10)
		} else if char >= 'A' && char <= 'F' {
			result += byte(char - 'A' + 10)
		}
	}
	return result
}
func (s *service) CreateAccountByRole(accountID uuid.UUID, role string) error {
	id := uuidToUint128(accountID.String())
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
