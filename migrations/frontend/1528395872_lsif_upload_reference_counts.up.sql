BEGIN;

ALTER TABLE lsif_uploads ADD COLUMN num_references int;
COMMENT ON COLUMN lsif_uploads.num_references IS 'The number of references to this upload data from other upload records (via lsif_references).';

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
