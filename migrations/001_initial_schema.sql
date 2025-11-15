-- +goose Up
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    name VARCHAR(255) NOT NULL,
    firebase_uid VARCHAR(255) UNIQUE,
    coin_balance DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Merchants/Restaurants table
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    category VARCHAR(50) DEFAULT 'restaurant', -- restaurant, grocery
    discount_percentage DECIMAL(5,2) DEFAULT 15.00,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Coin purchases
CREATE TABLE coin_purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    amount DECIMAL(10,2) NOT NULL,
    coins_received DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(50), -- stripe, razorpay, upi
    payment_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending', -- pending, completed, failed
    created_at TIMESTAMP DEFAULT NOW()
);

-- Transactions
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    merchant_id UUID REFERENCES merchants(id),
    coins_spent DECIMAL(10,2) NOT NULL,
    original_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) NOT NULL,
    final_amount DECIMAL(10,2) NOT NULL,
    transaction_type VARCHAR(20) DEFAULT 'payment', -- payment, refund
    status VARCHAR(20) DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Settlement records
CREATE TABLE settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_transactions INTEGER DEFAULT 0,
    total_discount_amount DECIMAL(10,2) DEFAULT 0.00,
    settlement_amount DECIMAL(10,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'pending', -- pending, paid, failed
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_merchant_id ON transactions(merchant_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_settlements_merchant_id ON settlements(merchant_id);

-- +goose Down
DROP TABLE IF EXISTS settlements;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS coin_purchases;
DROP TABLE IF EXISTS merchants;
DROP TABLE IF EXISTS users;
