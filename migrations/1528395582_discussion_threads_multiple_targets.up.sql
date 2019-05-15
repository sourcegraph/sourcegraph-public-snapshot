BEGIN;

-- Make discussion_threads <-> discussion_threads_target_repo a 1-to-n (not 1-to-1) relationship. The discussion_threads_target_repo DB table already has a thread_id column, so we just need to remove the discussion_threads.target_repo_id column.
ALTER TABLE discussion_threads DROP COLUMN target_repo_id;

COMMIT;
