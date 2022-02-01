BEGIN;

-- Drop new column
ALTER TABLE lsif_uploads DROP COLUMN reference_count;

-- Restore old comment on deprecated column
COMMENT ON COLUMN lsif_uploads.num_references IS 'The number of references to this upload data from other upload records (via lsif_references).';

COMMIT;
