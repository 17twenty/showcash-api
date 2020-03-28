CREATE TABLE IF NOT EXISTS showcash.comments (
    post_id         UUID NOT NULL,
    -- Extracted
    id              UUID PRIMARY KEY NOT NULL,
    date            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    comment         TEXT NOT NULL DEFAULT '',
    username        TEXT NOT NULL DEFAULT '',
    user_id         UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    -- End Extracted
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
