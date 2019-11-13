BEGIN;

ALTER TABLE lsif_dumps RENAME uploaded_at to processed_at;
ALTER TABLE lsif_dumps ADD COLUMN uploaded_at timestamp with time zone;
UPDATE lsif_dumps SET uploaded_at = processed_at;
ALTER TABLE lsif_dumps ALTER COLUMN uploaded_at SET NOT NULL;

COMMIT;
