BEGIN;

ALTER TABLE discussion_threads ADD COLUMN target_repo_id bigint REFERENCES discussion_threads_target_repo(id) ON DELETE CASCADE;
UPDATE discussion_threads d SET target_repo_id=(SELECT t.id FROM discussion_threads_target_repo t WHERE thread_id=d.id);

COMMIT;
