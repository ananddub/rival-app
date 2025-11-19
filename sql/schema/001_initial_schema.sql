-- +goose Up
-- Create enum for user roles
CREATE TYPE  user_role AS ENUM ('customer', 'merchant', 'admin');

-- Users table
CREATE TABLE users (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    phone VARCHAR(20),
    name VARCHAR(255) NOT NULL,
    profile_pic VARCHAR(500),
    firebase_uid VARCHAR(255) UNIQUE,
    coin_balance DECIMAL(10, 2) DEFAULT 0.00,
    role user_role DEFAULT 'customer',
    referral_code VARCHAR(20) UNIQUE,
    referred_by BIGINT REFERENCES users (id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Referral rewards table
CREATE TABLE referral_rewards (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    referrer_id BIGINT REFERENCES users (id),
    referred_id BIGINT REFERENCES users (id),
    reward_amount DECIMAL(10, 2) NOT NULL,
    reward_type VARCHAR(20) DEFAULT 'signup',
    status VARCHAR(20) DEFAULT 'pending',
    credited_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Merchants/Restaurants table
CREATE TABLE merchants (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    phone VARCHAR(20),
    category VARCHAR(50) DEFAULT 'restaurant',
    discount_percentage DECIMAL(5, 2) DEFAULT 15.00,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Merchant addresses table
CREATE TABLE merchant_addresses (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    merchant_id BIGINT REFERENCES merchants (id),
    street VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) DEFAULT 'India',
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    is_primary BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Coin purchases
CREATE TABLE coin_purchases (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT REFERENCES users (id),
    amount DECIMAL(10, 2) NOT NULL,
    coins_received DECIMAL(10, 2) NOT NULL,
    payment_method VARCHAR(50),
    payment_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);

-- JWT Sessions
CREATE TABLE jwt_sessions (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT REFERENCES users (id),
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Transactions
CREATE TABLE transactions (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT REFERENCES users (id),
    merchant_id BIGINT REFERENCES merchants (id),
    coins_spent DECIMAL(10, 2) NOT NULL,
    original_amount DECIMAL(10, 2) NOT NULL,
    discount_amount DECIMAL(10, 2) NOT NULL,
    final_amount DECIMAL(10, 2) NOT NULL,
    transaction_type VARCHAR(20) DEFAULT 'payment',
    status VARCHAR(20) DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Settlement records
CREATE TABLE settlements (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    merchant_id BIGINT REFERENCES merchants (id),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_transactions INTEGER DEFAULT 0,
    total_discount_amount DECIMAL(10, 2) DEFAULT 0.00,
    settlement_amount DECIMAL(10, 2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'pending',
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Offers table
CREATE TABLE offers (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    merchant_id BIGINT REFERENCES merchants (id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    discount_percentage DECIMAL(5, 2) NOT NULL,
    min_amount DECIMAL(10, 2) DEFAULT 0.00,
    max_discount DECIMAL(10, 2),
    is_active BOOLEAN DEFAULT true,
    valid_from TIMESTAMP DEFAULT NOW(),
    valid_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Orders table
CREATE TABLE orders (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    merchant_id BIGINT REFERENCES merchants (id),
    user_id BIGINT REFERENCES users (id),
    offer_id BIGINT REFERENCES offers (id),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    items JSONB NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    discount_amount DECIMAL(10, 2) DEFAULT 0.00,
    total_amount DECIMAL(10, 2) NOT NULL,
    coins_used DECIMAL(10, 2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'pending',
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Audit logs table
CREATE TABLE audit_logs (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    actor_id BIGINT,
    actor_type VARCHAR(20) NOT NULL,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id BIGINT,
    metadata JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_email ON users (email);

CREATE INDEX idx_users_firebase_uid ON users (firebase_uid);

CREATE INDEX idx_transactions_user_id ON transactions (user_id);

CREATE INDEX idx_transactions_merchant_id ON transactions (merchant_id);

CREATE INDEX idx_transactions_created_at ON transactions (created_at);

CREATE INDEX idx_settlements_merchant_id ON settlements (merchant_id);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;

DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS offers;

DROP TABLE IF EXISTS settlements;

DROP TABLE IF EXISTS transactions;

DROP TABLE IF EXISTS jwt_sessions;

DROP TABLE IF EXISTS coin_purchases;

DROP TABLE IF EXISTS merchant_addresses;

DROP TABLE IF EXISTS merchants;

DROP TABLE IF EXISTS referral_rewards;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_role;
