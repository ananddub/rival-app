-- +goose Up
CREATE TABLE IF NOT EXISTS public.rewards (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL, -- daily, weekly, achievement, referral
    coins BIGINT DEFAULT 0,
    money DECIMAL(10,2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS public.user_rewards (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    reward_id BIGINT NOT NULL REFERENCES public.rewards(id) ON DELETE CASCADE,
    claimed BOOLEAN DEFAULT false,
    claimed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, reward_id)
);

CREATE TABLE IF NOT EXISTS public.daily_rewards (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    day INTEGER NOT NULL CHECK (day >= 1 AND day <= 7),
    claimed_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, day)
);

CREATE TABLE IF NOT EXISTS public.referral_rewards (
    id BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    referred_user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    coins_earned BIGINT DEFAULT 100,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(referrer_id, referred_user_id)
);

-- Insert default rewards
INSERT INTO public.rewards (title, description, type, coins, money) VALUES
('Welcome Bonus', 'Welcome to Rival! Claim your bonus', 'achievement', 50, 0),
('First Purchase', 'Complete your first coin purchase', 'achievement', 25, 0),
('Profile Complete', 'Complete your profile information', 'achievement', 30, 0),
('Email Verified', 'Verify your email address', 'achievement', 20, 0),
('Phone Verified', 'Verify your phone number', 'achievement', 20, 0),
('Referral Bonus', 'Refer a friend and earn rewards', 'referral', 100, 0);

CREATE INDEX IF NOT EXISTS idx_user_rewards_user_id ON public.user_rewards(user_id);
CREATE INDEX IF NOT EXISTS idx_daily_rewards_user_id ON public.daily_rewards(user_id);
CREATE INDEX IF NOT EXISTS idx_referral_rewards_referrer_id ON public.referral_rewards(referrer_id);

-- +goose Down
DROP TABLE IF EXISTS public.referral_rewards CASCADE;
DROP TABLE IF EXISTS public.daily_rewards CASCADE;
DROP TABLE IF EXISTS public.user_rewards CASCADE;
DROP TABLE IF EXISTS public.rewards CASCADE;
