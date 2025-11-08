-- +goose Up
CREATE TABLE IF NOT EXISTS public.coin_packages (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    coins BIGINT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    bonus_coins BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS public.coin_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    coins BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL,
    reason TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS public.coin_purchases (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    package_id BIGINT NOT NULL REFERENCES public.coin_packages(id),
    coins_received BIGINT NOT NULL,
    amount_paid DECIMAL(10,2) NOT NULL,
    payment_status VARCHAR(20) DEFAULT 'pending',
    payment_id TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coin_transactions_user_id ON public.coin_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_coin_purchases_user_id ON public.coin_purchases(user_id);

-- +goose Down
DROP TABLE IF EXISTS public.coin_purchases CASCADE;
DROP TABLE IF EXISTS public.coin_transactions CASCADE;
DROP TABLE IF EXISTS public.coin_packages CASCADE;
