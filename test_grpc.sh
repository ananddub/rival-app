#!/bin/bash

echo "=== Testing Payment APIs with grpcurl ==="
echo ""

USER_ID=3702

echo "1. Checking balance for user $USER_ID..."
grpcurl -plaintext -d "{\"user_id\": $USER_ID}" localhost:8080 rival.api.v1.PaymentService/GetBalance
echo ""

echo "2. Getting payment history..."
grpcurl -plaintext -d "{\"user_id\": $USER_ID, \"page\": 1, \"limit\": 10}" localhost:8080 rival.api.v1.PaymentService/GetPaymentHistory
echo ""

echo "3. Initiating coin purchase (50 coins)..."
grpcurl -plaintext -d "{\"user_id\": $USER_ID, \"amount\": 50, \"payment_method\": \"stripe\"}" localhost:8080 rival.api.v1.PaymentService/InitiateCoinPurchase
echo ""

echo "4. Checking updated balance..."
grpcurl -plaintext -d "{\"user_id\": $USER_ID}" localhost:8080 rival.api.v1.PaymentService/GetBalance
echo ""

echo "5. Getting updated payment history..."
grpcurl -plaintext -d "{\"user_id\": $USER_ID, \"page\": 1, \"limit\": 10}" localhost:8080 rival.api.v1.PaymentService/GetPaymentHistory
echo ""

echo "=== Test Complete ==="
