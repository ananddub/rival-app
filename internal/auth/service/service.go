package authService

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"encore.app/config"
	"encore.app/connection"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var rdb *redis.Client
var jwtSecret = []byte("your-secret-key") // TODO: Move to config

func init() {
	cfg := config.GetConfig()
	rdb = connection.GetRedisClient(&cfg.Redis)
}

func GenerateOTP() string {
	otp, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", otp.Int64()+100000)
}

func StoreOTP(ctx context.Context, email, otp string) error {
	key := fmt.Sprintf("otp:%s", email)
	return rdb.Set(ctx, key, otp, 5*time.Minute).Err()
}

func VerifyOTP(ctx context.Context, email, otp string) bool {
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

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func GenerateAccessToken(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
