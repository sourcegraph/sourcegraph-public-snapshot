BEGIN;

DROP INDEX IF EXISTS lsif_dumps_uploaded_at;
DROP INDEX IF EXISTS lsif_dumps_visible_at_tip;
ALTER TABLE lsif_dumps DROP COLUMN uploaded_at;

COMMIT;
