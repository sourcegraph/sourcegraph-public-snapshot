ALTER TABLE
    lsif_uploads
ADD
    COLUMN IF NOT EXISTS last_referenced_scan_at timestamp with time zone;

COMMENT ON COLUMN lsif_uploads.last_referenced_scan_at IS 'The last time this upload was known to be referenced by another (possibly expired) index.';
