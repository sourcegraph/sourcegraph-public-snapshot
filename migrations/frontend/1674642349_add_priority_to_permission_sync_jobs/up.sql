ALTER TABLE permission_sync_jobs ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 0;
COMMENT ON COLUMN permission_sync_jobs.priority IS 'Specifies numeric priority for the permissions sync job.';

ALTER TABLE permission_sync_jobs DROP COLUMN IF EXISTS high_priority;

-- this index is used as a last resort if deduplication logic fails to work.
-- we should not enqueue more that one high priority immediate sync job (process_after IS NULL) for given repo/user.
CREATE UNIQUE INDEX IF NOT EXISTS permission_sync_jobs_unique ON permission_sync_jobs
    USING btree (priority, user_id, repository_id, cancel, process_after)
    WHERE (state = 'queued');
