DROP TABLE IF EXISTS search_context_stars;

DROP TABLE IF EXISTS search_context_default;

DELETE FROM search_contexts WHERE autodefined = true;

ALTER TABLE IF EXISTS search_contexts
    DROP COLUMN IF EXISTS autodefined;
