BEGIN;

ALTER TABLE lsif_uploads ADD COLUMN num_references int;
COMMENT ON COLUMN lsif_uploads.num_references IS 'The number of references to this upload data from other upload records (via lsif_references).';

COMMIT;
