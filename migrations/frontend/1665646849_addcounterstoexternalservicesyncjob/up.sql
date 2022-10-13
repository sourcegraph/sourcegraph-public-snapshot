ALTER TABLE IF EXISTS external_service_sync_jobs
ADD COLUMN repos_synced integer DEFAULT 0 NOT NULL,
ADD COLUMN repo_sync_errors integer DEFAULT 0 NOT NULL,
ADD COLUMN repos_added integer DEFAULT 0 NOT NULL,
ADD COLUMN repos_removed integer DEFAULT 0 NOT NULL,
ADD COLUMN repos_modified integer DEFAULT 0 NOT NULL,
ADD COLUMN repos_unmodified integer DEFAULT 0 NOT NULL,
ADD COLUMN repos_deleted integer DEFAULT 0 NOT NULL;
