ALTER TABLE permission_sync_jobs
    ADD COLUMN IF NOT EXISTS reason TEXT,
    ADD COLUMN IF NOT EXISTS triggered_by_user_id INTEGER,
    ADD FOREIGN KEY (triggered_by_user_id) REFERENCES users(id) ON DELETE SET NULL DEFERRABLE;

COMMENT ON COLUMN permission_sync_jobs.reason IS 'Specifies why permissions sync job was triggered.';
COMMENT ON COLUMN permission_sync_jobs.triggered_by_user_id IS 'Specifies an ID of a user who triggered a sync.';
