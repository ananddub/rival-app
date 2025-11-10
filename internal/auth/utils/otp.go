package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

var (
	otpStore = make(map[string]otpData)
	otpMutex = sync.RWMutex{}
	
	// Store pending signup data temporarily
	pendingSignups = make(map[string]*PendingSignupData)
	signupMutex = sync.RWMutex{}
)

type otpData struct {
	otp       string
	expiresAt time.Time
}

type PendingSignupData struct {
	Name        string
	PhoneNumber string
	Password    string
	CreatedAt   time.Time
}

func GenerateOTP() string {
	otp, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", otp.Int64()+100000)
}

func StoreOtp(email string) string {
	otp := GenerateOTP()
	otpMutex.Lock()
	defer otpMutex.Unlock()
	
	otpStore[email] = otpData{
		otp:       otp,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	return otp
}

func VerifyOtp(email, otp string) bool {
	// For testing, accept 123456 as valid OTP
	if otp == "123456" {
		return true
	}
	
	otpMutex.Lock()
	defer otpMutex.Unlock()
	
	data, exists := otpStore[email]
	if !exists {
		return false
	}
	
	if time.Now().After(data.expiresAt) {
		delete(otpStore, email)
		return false
	}
	
	if data.otp == otp {
		delete(otpStore, email)
		return true
	}
	
	return false
}

// StorePendingSignup stores signup data temporarily until OTP verification
func StorePendingSignup(email, name, phoneNumber, password string) {
	signupMutex.Lock()
	defer signupMutex.Unlock()
	
	pendingSignups[email] = &PendingSignupData{
		Name:        name,
		PhoneNumber: phoneNumber,
		Password:    password,
		CreatedAt:   time.Now(),
	}
}

// GetPendingSignup retrieves pending signup data
func GetPendingSignup(email string) *PendingSignupData {
	signupMutex.RLock()
	defer signupMutex.RUnlock()
	
	data, exists := pendingSignups[email]
	if !exists {
		return nil
	}
	
	// Check if data is expired (1 hour)
	if time.Since(data.CreatedAt) > time.Hour {
		delete(pendingSignups, email)
		return nil
	}
	
	return data
}

// ClearPendingSignup removes pending signup data
func ClearPendingSignup(email string) {
	signupMutex.Lock()
	defer signupMutex.Unlock()
	
	delete(pendingSignups, email)
}
