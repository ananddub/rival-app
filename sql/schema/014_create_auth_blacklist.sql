-- +goose Up
-- Blacklisted tokens for logout functionality
CREATE TABLE IF NOT EXISTS public.blacklisted_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    hashed_token VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_user_id ON public.blacklisted_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_blacklisted_tokens_hash ON public.blacklisted_tokens(hashed_token);

-- +goose Down
DROP TABLE IF EXISTS public.blacklisted_tokens CASCADE;
