ALTER TABLE IF EXISTS external_service_sync_jobs
DROP COLUMN IF EXISTS repos_synced,
DROP COLUMN IF EXISTS repo_sync_errors,
DROP COLUMN IF EXISTS repos_added,
DROP COLUMN IF EXISTS repos_deleted,
DROP COLUMN IF EXISTS repos_modified,
DROP COLUMN IF EXISTS repos_unmodified;
