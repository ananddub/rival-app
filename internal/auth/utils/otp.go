package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"encore.app/config"
	"encore.app/connection"
)

func GenerateOTP() string {
	otp, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", otp.Int64()+100000)
}

func StoreOTP(ctx context.Context, email, otp string) error {
	cfg := config.GetConfig()
	rdb := connection.GetRedisClient(&cfg.Redis)
	key := fmt.Sprintf("otp:%s", email)
	return rdb.Set(ctx, key, otp, 5*time.Minute).Err()
}

func VerifyOTP(ctx context.Context, email, otp string) bool {
	cfg := config.GetConfig()
	rdb := connection.GetRedisClient(&cfg.Redis)
	key := fmt.Sprintf("otp:%s", email)
	storedOtp, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	if storedOtp == otp {
		rdb.Del(ctx, key)
		return true
	}
	return false
}
