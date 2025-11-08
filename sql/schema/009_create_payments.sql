-- +goose Up
CREATE TABLE IF NOT EXISTS public.payments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    order_id TEXT NOT NULL UNIQUE,
    payment_id TEXT,
    amount DECIMAL(10,2) NOT NULL,
    currency TEXT DEFAULT 'INR',
    status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50),
    payment_gateway VARCHAR(50),
    payment_type VARCHAR(20) NOT NULL,
    reference_id BIGINT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON public.payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON public.payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON public.payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_type ON public.payments(payment_type);

-- +goose Down
DROP TABLE IF EXISTS public.payments CASCADE;
