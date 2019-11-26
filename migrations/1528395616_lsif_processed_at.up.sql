BEGIN;

ALTER TABLE lsif_dumps RENAME uploaded_at TO processed_at;
ALTER TABLE lsif_dumps ADD COLUMN uploaded_at TIMESTAMP WITH TIME ZONE;
UPDATE lsif_dumps SET uploaded_at = processed_at;
ALTER TABLE lsif_dumps ALTER COLUMN uploaded_at SET NOT NULL;
DROP INDEX lsif_dumps_uploaded_at;
CREATE INDEX lsif_dumps_uploaded_at ON lsif_dumps(uploaded_at);

COMMIT;
