CREATE SCHEMA IF NOT EXISTS showcash;

-- Functions!
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.last_updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TABLE IF NOT EXISTS showcash.post (
    user_id         UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000', -- This is blank for v1
    
    -- Extracted
    id              UUID PRIMARY KEY,
    title           TEXT NOT NULL DEFAULT '',
    imageuri        TEXT NOT NULL DEFAULT '',
    date            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    -- Note: We extract the itemList seperately
    -- End Extracted
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS showcash.item (
    post_id       UUID NOT NULL,
    -- Extracted
    id            BIGINT NOT NULL DEFAULT 0,
    title         TEXT NOT NULL DEFAULT '',
    description   TEXT NOT NULL DEFAULT '',
    link          TEXT NOT NULL DEFAULT '',
    "left" BIGINT NOT NULL DEFAULT 0,
    top BIGINT NOT NULL DEFAULT 0,
    -- End Extracted
    PRIMARY KEY(post_id, id),
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
