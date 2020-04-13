CREATE TABLE IF NOT EXISTS showcash.posttag (
    tag_id              BIGINT NOT NULL,
    post_id             UUID NOT NULL,
    PRIMARY KEY(post_id, tag_id)
);

CREATE TABLE IF NOT EXISTS showcash.tag (
    tag_id            BIGSERIAL PRIMARY KEY,
    -- Extracted
    tag               TEXT NOT NULL DEFAULT '' UNIQUE
    -- End Extracted
);
