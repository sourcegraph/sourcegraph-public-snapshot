ALTER TABLE
    lsif_uploads
ADD
    COLUMN IF NOT EXISTS last_traversal_scan_at timestamp with time zone;

COMMENT ON COLUMN lsif_uploads.last_traversal_scan_at IS 'The last time this upload was known to be reachable by a non-expired index.';
