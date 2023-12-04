ALTER TABLE permission_sync_jobs ADD COLUMN IF NOT EXISTS cancellation_reason TEXT;

COMMENT ON COLUMN permission_sync_jobs.cancellation_reason IS 'Specifies why permissions sync job was cancelled.';
