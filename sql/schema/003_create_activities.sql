-- +goose Up
CREATE TABLE IF NOT EXISTS public.activities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    details TEXT NOT NULL,
    category TEXT NOT NULL,
    icon TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activities_user_id ON public.activities(user_id);
CREATE INDEX IF NOT EXISTS idx_activities_category ON public.activities(category);

-- +goose Down
DROP TABLE IF EXISTS public.activities CASCADE;
