BEGIN;

ALTER TABLE lsif_dumps DROP COLUMN uploaded_at;
ALTER TABLE lsif_dumps RENAME COLUMN processed_at to uploaded_at;
CREATE INDEX lsif_dumps_uploaded_at ON lsif_dumps(uploaded_at);

COMMIT;
