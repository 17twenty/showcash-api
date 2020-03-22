CREATE TABLE IF NOT EXISTS showcash.views (
        post_id         UUID NOT NULL,
        unique_value    TEXT NOT NULL DEFAULT '',
        viewed_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
        PRIMARY KEY(post_id, unique_value)
);

-- Get most popular
-- SELECT post_id, COUNT(*) AS counted
-- FROM   showcash.views
-- -- WHERE  month = 'May'  -- or whatever is stored in your varchar(3) column
-- GROUP  BY post_id
-- ORDER  BY counted DESC, post_id  -- to break ties in deterministic fashion
-- LIMIT  10;


-- Increase view
-- INSERT INTO showcash.views (
--     post_id,
--     unique_value
-- ) VALUES (
--     '00000000-0000-0000-0000-000000000000',
--     '32123'
-- ) ON CONFLICT (post_id, unique_value) DO
-- UPDATE SET 
--     viewed_at = NOW();

-- -- Upvote
-- INSERT INTO showcash.popularity (
--     post_id,
--     views,
--     views_last,
--     votes,
--     votes_last
-- ) VALUES (
--     '00000000-0000-0000-0000-000000000000',
--     1,
--     NOW(),
--     1,
--     NOW()
-- ) ON CONFLICT (post_id) DO
-- UPDATE SET 
--     votes = showcash.popularity.votes + 1,
--     votes_last = NOW();

-- -- Downvote
-- INSERT INTO showcash.popularity (
--     post_id,
--     views,
--     views_last,
--     votes,
--     votes_last
-- ) VALUES (
--     '00000000-0000-0000-0000-000000000000',
--     1,
--     NOW(),
--     1,
--     NOW()
-- ) ON CONFLICT (post_id) DO
-- UPDATE SET 
--     votes = showcash.popularity.votes - 1,
--     votes_last = NOW();