ALTER TABLE IF EXISTS external_service_sync_jobs
ADD COLUMN IF NOT EXISTS repos_synced integer DEFAULT 0 NOT NULL,
ADD COLUMN IF NOT EXISTS repo_sync_errors integer DEFAULT 0 NOT NULL,
ADD COLUMN IF NOT EXISTS repos_added integer DEFAULT 0 NOT NULL,
ADD COLUMN IF NOT EXISTS repos_deleted integer DEFAULT 0 NOT NULL,
ADD COLUMN IF NOT EXISTS repos_modified integer DEFAULT 0 NOT NULL,
ADD COLUMN IF NOT EXISTS repos_unmodified integer DEFAULT 0 NOT NULL;

COMMENT ON COLUMN external_service_sync_jobs.repos_synced IS 'The number of repos synced during this sync job.';
COMMENT ON COLUMN external_service_sync_jobs.repo_sync_errors IS 'The number of times an error occurred syncing a repo during this sync job.';
COMMENT ON COLUMN external_service_sync_jobs.repos_added IS 'The number of new repos discovered during this sync job.';
COMMENT ON COLUMN external_service_sync_jobs.repos_deleted IS 'The number of repos deleted as a result of this sync job.';
COMMENT ON COLUMN external_service_sync_jobs.repos_modified IS 'The number of existing repos whose metadata has changed during this sync job.';
COMMENT ON COLUMN external_service_sync_jobs.repos_unmodified IS 'The number of existing repos whose metadata did not change during this sync job.';
