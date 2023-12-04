DROP INDEX IF EXISTS lsif_uploads_last_reconcile_at;

ALTER TABLE
    lsif_uploads DROP COLUMN IF EXISTS last_reconcile_at;
