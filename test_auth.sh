#!/bin/bash

echo "Testing Auth APIs..."

echo "1. Testing Login API:"
curl -X POST http://127.0.0.1:4000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  -w "\nStatus: %{http_code}\n\n"

echo "2. Testing Signup API:"
curl -X POST http://127.0.0.1:4000/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"password123","phone":"+1234567890"}' \
  -w "\nStatus: %{http_code}\n\n"

echo "3. Testing Verify OTP API:"
curl -X POST http://127.0.0.1:4000/auth/login/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","otp":"123456"}' \
  -w "\nStatus: %{http_code}\n\n"

echo "4. Testing Forgot Password API:"
curl -X POST http://127.0.0.1:4000/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}' \
  -w "\nStatus: %{http_code}\n\n"

echo "5. Testing Reset Password API:"
curl -X POST http://127.0.0.1:4000/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","otp":"123456","new_password":"newpassword123"}' \
  -w "\nStatus: %{http_code}\n\n"
