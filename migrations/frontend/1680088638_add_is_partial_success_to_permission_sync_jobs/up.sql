ALTER TABLE IF EXISTS permission_sync_jobs
    ADD COLUMN IF NOT EXISTS is_partial_success BOOLEAN DEFAULT FALSE;
