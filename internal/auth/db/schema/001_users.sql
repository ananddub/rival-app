drop table if exists public.users CASCADE;

CREATE TYPE sign_type AS ENUM ('email', 'google', 'github');
CREATE TYPE role_type AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS public.users
(
    id              bigserial              primary key,
    full_name       varchar(255)           not null,
    email           varchar(255)           not null unique,
    password_hash   varchar(255),
    dob             timestamp,
    is_phone_verified boolean DEFAULT false,
     is_email_verified boolean DEFAULT false,
    sign_type       sign_type   default    'email'::sign_type not null,
    role            role_type   default    'user'::role_type  not null,
    created_at      timestamp   default    now(),
    updated_at      timestamp   default    now()
);