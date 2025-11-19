package referral

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	schema "rival/gen/sql"
	"rival/pkg/tb"
	"rival/pkg/utils"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	ReferrerBonus float64 // Bonus for person who referred
	RefereeBonus  float64 // Bonus for new user
	CodeLength    int     // Length of referral code
	CodePrefix    string  // Prefix for codes (e.g., "RIV")
	MaxRewards    int     // Max rewards per referrer
	ExpiryDays    int     // Days before reward expires
}

type Service struct {
	db      *pgxpool.Pool
	queries *schema.Queries
	tb      *tb.TbService
	config  Config
}

func NewService(db *pgxpool.Pool, tb *tb.TbService) *Service {
	return &Service{
		db:      db,
		queries: schema.New(db),
		tb:      tb,
		config: Config{
			ReferrerBonus: 5.0,   // $5 for referrer
			RefereeBonus:  10.0,  // $10 for new user
			CodeLength:    6,     // 6 character codes
			CodePrefix:    "RIV", // RIVAL prefix
			MaxRewards:    50,    // Max 50 referrals per user
			ExpiryDays:    30,    // 30 days to claim
		},
	}
}

func (s *Service) GenerateReferralCode(userName string) string {
	// Create user-friendly code: PREFIX + first 2 letters of name + 4 random chars
	namePrefix := strings.ToUpper(userName)
	if len(namePrefix) >= 2 {
		namePrefix = namePrefix[:2]
	} else {
		namePrefix = "US" // Default
	}

	// Generate 4 random alphanumeric characters
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomPart := make([]byte, 4)
	for i := range randomPart {
		randomPart[i] = chars[rand.Intn(len(chars))]
	}

	return s.config.CodePrefix + namePrefix + string(randomPart)
}

func (s *Service) ProcessReferral(ctx context.Context, referrerCode string, newUserID int) error {

	// Find referrer by code
	referrer, err := s.queries.GetUserByReferralCode(ctx, pgtype.Text{String: referrerCode, Valid: true})
	if err != nil {
		return fmt.Errorf("invalid referral code: %v", err)
	}

	// Check if referrer has reached max rewards
	stats, err := s.queries.GetUserReferralStats(ctx, pgtype.Int8{Int64: referrer.ID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to get referral stats: %v", err)
	}

	if int(stats.TotalReferrals) >= s.config.MaxRewards {
		return fmt.Errorf("referrer has reached maximum referral limit")
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	_, err = qtx.CreateReferralReward(ctx, schema.CreateReferralRewardParams{
		ReferrerID:   pgtype.Int8{Int64: referrer.ID, Valid: true},
		RewardAmount: utils.Float64ToNumeric(s.config.ReferrerBonus),
		RewardType:   pgtype.Text{String: "signup", Valid: true},
		Status:       pgtype.Text{String: "pending", Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create referral reward: %v", err)
	}

	// Give bonus to new user immediately
	err = (*s.tb).AddCoins(newUserID, s.config.RefereeBonus)
	if err != nil {
		return fmt.Errorf("failed to add bonus to new user: %v", err)
	}
	err = (*s.tb).AddCoins(int(referrer.ID), s.config.ReferrerBonus)
	if err != nil {
		return fmt.Errorf("failed to add bonus to referrer: %v", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (s *Service) ValidateReferralCode(ctx context.Context, code string) (bool, error) {
	if len(code) < 6 {
		return false, fmt.Errorf("referral code too short")
	}

	if !strings.HasPrefix(code, s.config.CodePrefix) {
		return false, fmt.Errorf("invalid referral code format")
	}

	// Check if code exists
	_, err := s.queries.GetUserByReferralCode(ctx, pgtype.Text{String: code, Valid: true})
	if err != nil {
		return false, fmt.Errorf("referral code not found")
	}

	return true, nil
}

func (s *Service) GetReferralStats(ctx context.Context, userID int) (*ReferralStats, error) {
	stats, err := s.queries.GetUserReferralStats(ctx, pgtype.Int8{Int64: int64(userID), Valid: true})
	if err != nil {
		return nil, err
	}

	return &ReferralStats{
		TotalReferrals: int(stats.TotalReferrals),
		TotalEarned:    utils.NumericToFloat64(stats.TotalEarned.(pgtype.Numeric)),
		MaxRewards:     s.config.MaxRewards,
		RemainingSlots: s.config.MaxRewards - int(stats.TotalReferrals),
	}, nil
}

type ReferralStats struct {
	TotalReferrals int     `json:"total_referrals"`
	TotalEarned    float64 `json:"total_earned"`
	MaxRewards     int     `json:"max_rewards"`
	RemainingSlots int     `json:"remaining_slots"`
}
