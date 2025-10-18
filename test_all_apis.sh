#!/bin/bash

BASE_URL="http://192.168.1.5:4000"

echo "=== Testing All Auth APIs ==="

echo "1. Signup:"
curl -X POST $BASE_URL/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"password123","phone":"+1234567890"}' \
  -w "\nStatus: %{http_code}\n\n"

echo "2. Signup Verify OTP:"
curl -X POST $BASE_URL/auth/signup/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","otp":"123456"}' \
  -w "\nStatus: %{http_code}\n\n"

echo "3. Login:"
curl -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  -w "\nStatus: %{http_code}\n\n"

echo "4. Login Verify OTP:"
curl -X POST $BASE_URL/auth/login/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","otp":"123456"}' \
  -w "\nStatus: %{http_code}\n\n"

echo "5. Forgot Password:"
curl -X POST $BASE_URL/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}' \
  -w "\nStatus: %{http_code}\n\n"

echo "6. Reset Password:"
curl -X POST $BASE_URL/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","otp":"123456","new_password":"newpass123"}' \
  -w "\nStatus: %{http_code}\n\n"
