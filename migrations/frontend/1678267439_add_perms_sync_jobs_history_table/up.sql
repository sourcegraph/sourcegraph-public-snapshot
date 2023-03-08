CREATE TABLE IF NOT EXISTS perms_sync_jobs_history(
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE NULL REFERENCES users(id) ON DELETE CASCADE,
    repo_id INT UNIQUE NULL REFERENCES repo(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Indexes for fast lookups of oldest sync jobs for user or repo
CREATE INDEX IF NOT EXISTS perms_sync_jobs_history_sorted_updated_at_user_id
    ON perms_sync_jobs_history(updated_at ASC, user_id ASC NULLS LAST)
    WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS perms_sync_jobs_history_sorted_updated_at_repo_id
    ON perms_sync_jobs_history(updated_at ASC, repo_id ASC NULLS LAST)
    WHERE repo_id IS NOT NULL;

-- Copy existing data from permission_sync_jobs to perms_sync_jobs_history
INSERT INTO perms_sync_jobs_history(user_id, updated_at)
    SELECT DISTINCT ON (user_id)
        user_id, finished_at as updated_at
    FROM permission_sync_jobs
    WHERE user_id IS NOT NULL
        AND finished_at IS NOT NULL
    ORDER BY user_id ASC, finished_at ASC
    ON CONFLICT DO NOTHING;

INSERT INTO perms_sync_jobs_history(repo_id, updated_at)
    SELECT DISTINCT ON (repository_id)
        repository_id AS repo_id, finished_at as updated_at
    FROM permission_sync_jobs
    WHERE repository_id IS NOT NULL
        AND finished_at IS NOT NULL
    ORDER BY repository_id ASC, finished_at ASC
    ON CONFLICT DO NOTHING;

-- Add triggers to delete from the table if user or repo is soft deleted
CREATE OR REPLACE FUNCTION delete_perms_sync_jobs_history_on_user_soft_delete() RETURNS trigger
	LANGUAGE plpgsql
  AS $$ BEGIN
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
    	DELETE FROM perms_sync_jobs_history WHERE user_id = OLD.id;
    END IF;
    RETURN NULL;
  END
$$;

DROP TRIGGER IF EXISTS trig_delete_perms_sync_jobs_history_on_user_soft_delete ON users;
CREATE TRIGGER trig_delete_perms_sync_jobs_history_on_user_soft_delete
    AFTER UPDATE ON users FOR EACH ROW
    EXECUTE FUNCTION delete_perms_sync_jobs_history_on_user_soft_delete();

CREATE OR REPLACE FUNCTION delete_perms_sync_jobs_history_on_repo_soft_delete() RETURNS trigger
	LANGUAGE plpgsql
  AS $$ BEGIN
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
    	DELETE FROM perms_sync_jobs_history WHERE repo_id = OLD.id;
    END IF;
    RETURN NULL;
  END
$$;

DROP TRIGGER IF EXISTS trig_delete_perms_sync_jobs_history_on_repo_soft_delete ON repo;
CREATE TRIGGER trig_delete_perms_sync_jobs_history_on_repo_soft_delete
    AFTER UPDATE ON repo FOR EACH ROW
    EXECUTE FUNCTION delete_perms_sync_jobs_history_on_repo_soft_delete();

-- Add trigger to copy the data to the table when a sync job is finished
CREATE OR REPLACE FUNCTION copy_to_sync_jobs_history_on_update() RETURNS trigger
	LANGUAGE plpgsql
  AS $$ BEGIN
    IF NEW.user_id IS NOT NULL AND NEW.finished_at IS NOT NULL AND (OLD.finished_at IS NULL OR NEW.finished_at > OLD.finished_at) THEN
    	INSERT INTO perms_sync_jobs_history(user_id, updated_at)
            VALUES (NEW.user_id, NEW.finished_at)
            ON CONFLICT(user_id) DO UPDATE SET updated_at = EXCLUDED.updated_at;
    END IF;
    IF NEW.repository_id IS NOT NULL AND NEW.finished_at IS NOT NULL AND (OLD.finished_at IS NULL OR NEW.finished_at > OLD.finished_at) THEN
    	INSERT INTO perms_sync_jobs_history(repo_id, updated_at)
            VALUES (NEW.repository_id, NEW.finished_at)
            ON CONFLICT(repo_id) DO UPDATE SET updated_at = EXCLUDED.updated_at;
    END IF;
    RETURN NULL;
  END
$$;

DROP TRIGGER IF EXISTS trig_copy_to_sync_jobs_history_on_update ON permission_sync_jobs;
CREATE TRIGGER trig_copy_to_sync_jobs_history_on_update
    AFTER UPDATE ON permission_sync_jobs FOR EACH ROW
    EXECUTE FUNCTION copy_to_sync_jobs_history_on_update();
