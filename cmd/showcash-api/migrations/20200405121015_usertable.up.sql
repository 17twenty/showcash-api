CREATE TABLE IF NOT EXISTS showcash.user (
    -- Extracted
    user_id             UUID PRIMARY KEY NOT NULL,
    username            TEXT NOT NULL DEFAULT '' UNIQUE,
    realname            TEXT NOT NULL DEFAULT '',
    location            TEXT NOT NULL DEFAULT '',
    profile_uri         TEXT NOT NULL DEFAULT '',
    bio                 TEXT NOT NULL DEFAULT '',
    social_1            TEXT NOT NULL DEFAULT '', -- instagram
    social_2            TEXT NOT NULL DEFAULT '', -- facebook
    social_3            TEXT NOT NULL DEFAULT '', -- twitter

    email_address       TEXT NOT NULL DEFAULT '' UNIQUE,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    -- End Extracted
    password            TEXT NOT NULL,
    shadow_banned       BOOLEAN NOT NULL DEFAULT FALSE
);