-- +goose Up
CREATE TABLE IF NOT EXISTS public.jwt_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_jwt_user_id ON public.jwt_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_jwt_token_hash ON public.jwt_tokens(token_hash);

-- +goose Down
DROP TABLE IF EXISTS public.jwt_tokens CASCADE;
