ALTER TABLE
    lsif_uploads
ADD
    COLUMN IF NOT EXISTS last_reconcile_at timestamp with time zone;

CREATE INDEX IF NOT EXISTS lsif_uploads_last_reconcile_at ON lsif_uploads(last_reconcile_at, id)
WHERE
    state = 'completed';
