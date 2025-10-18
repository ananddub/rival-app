drop table jwt_tokens;
CREATE TABLE jwt_tokens
(
    id bigserial primary key,
    user_id bigint not null references public.users on delete cascade,
    token_hash varchar(255) not null,
    expires_at timestamp    not null,
    created_at timestamp default now()
);

CREATE INDEX idx_jwt_user_id ON public.jwt_tokens (user_id);
CREATE INDEX idx_jwt_token_hash ON public.jwt_tokens (token_hash);
