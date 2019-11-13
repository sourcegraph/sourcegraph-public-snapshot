BEGIN;

ALTER TABLE lsif_dumps DROP COLUMN uploaded_at;
ALTER TABLE lsif_dumps RENAME COLUMN processed_at to uploaded_at;

COMMIT;
