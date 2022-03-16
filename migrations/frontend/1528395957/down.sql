ALTER TABLE notebooks DROP COLUMN IF EXISTS blocks_tsvector;

ALTER TABLE notebooks ALTER COLUMN title TYPE CITEXT;

DROP INDEX IF EXISTS notebooks_title_trgm_idx;

DROP INDEX IF EXISTS notebooks_blocks_tsvector_idx;
