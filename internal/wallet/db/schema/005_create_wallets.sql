-- +goose Up
CREATE TABLE IF NOT EXISTS public.wallets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES public.users(id) ON DELETE CASCADE,
    balance DECIMAL(15,2) DEFAULT 0.00,
    coins BIGINT DEFAULT 0,
    currency TEXT DEFAULT 'INR',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS public.transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    wallet_id BIGINT NOT NULL REFERENCES public.wallets(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    amount DECIMAL(15,2) NOT NULL,
    type VARCHAR(20) NOT NULL,
    icon TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON public.wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON public.transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON public.transactions(wallet_id);

-- +goose Down
DROP TABLE IF EXISTS public.transactions CASCADE;
DROP TABLE IF EXISTS public.wallets CASCADE;
