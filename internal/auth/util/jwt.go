package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	schemapb "rival/gen/proto/proto/schema"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUtil interface {
	GenerateTokens(user *schemapb.User) (accessToken, refreshToken string, err error)
	ValidateToken(token string) (*TokenClaims, error)
	RefreshAccessToken(refreshToken string) (string, error)
	HashToken(token string) string
}

type TokenClaims struct {
	UserID int               `json:"user_id"`
	Email  string            `json:"email"`
	Role   schemapb.UserRole `json:"role"`
	Exp    int64             `json:"exp"`
	jwt.RegisteredClaims
}

type jwtUtil struct {
	secretKey       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTUtil(secretKey string, accessTTL, refreshTTL time.Duration) JWTUtil {
	return &jwtUtil{
		secretKey:       secretKey,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (j *jwtUtil) GenerateTokens(user *schemapb.User) (accessToken, refreshToken string, err error) {
	// Access token claims
	accessClaims := TokenClaims{
		UserID: int(user.Id),
		Email:  user.Email,
		Role:   user.Role,
		Exp:    time.Now().Add(j.accessTokenTTL).Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.Id),
		},
	}

	// Generate access token
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", "", err
	}

	// Refresh token claims
	refreshClaims := TokenClaims{
		UserID: int(user.Id),
		Email:  user.Email,
		Role:   user.Role,
		Exp:    time.Now().Add(j.refreshTokenTTL).Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.Id),
		},
	}

	// Generate refresh token
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (j *jwtUtil) ValidateToken(token string) (*TokenClaims, error) {
	tokenObj, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := tokenObj.Claims.(*TokenClaims); ok && tokenObj.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (j *jwtUtil) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Create new access token with same user info
	newClaims := TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		Exp:    time.Now().Add(j.accessTokenTTL).Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", claims.UserID),
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return accessTokenObj.SignedString([]byte(j.secretKey))
}

func (j *jwtUtil) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
