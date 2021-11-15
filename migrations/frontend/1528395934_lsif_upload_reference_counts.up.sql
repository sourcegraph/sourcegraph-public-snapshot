BEGIN;

ALTER TABLE lsif_uploads ADD COLUMN reference_count int;
COMMENT ON COLUMN lsif_uploads.reference_count IS 'The number of references to this upload data from other upload records (via lsif_references).';
COMMENT ON COLUMN lsif_uploads.num_references IS 'Deprecated in favor of reference_count.';

COMMIT;
