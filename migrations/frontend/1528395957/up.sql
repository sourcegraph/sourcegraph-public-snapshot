-- We add a generated column that will contain a tsvector representation of strings contained in the blocks array (extracted with the jsonb_to_tsvector expression).
ALTER TABLE notebooks ADD COLUMN IF NOT EXISTS blocks_tsvector TSVECTOR GENERATED ALWAYS AS (jsonb_to_tsvector('english', blocks, '["string"]')) STORED;

-- Postgres does not support trigram indices on a CITEXT column. We have to revert it back to a regular TEXT column to apply a trigram index.
ALTER TABLE notebooks ALTER COLUMN title TYPE TEXT;

CREATE INDEX IF NOT EXISTS notebooks_title_trgm_idx ON notebooks USING GIN (title gin_trgm_ops);

-- TSVECTOR columns do not support a gin_trgm_ops index, so we omit it here.
CREATE INDEX IF NOT EXISTS notebooks_blocks_tsvector_idx ON notebooks USING GIN (blocks_tsvector);
