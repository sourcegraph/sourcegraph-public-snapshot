-- We are now using hard deletes instead of soft deletes on repos
BEGIN;

ALTER TABLE discussion_comments
DROP CONSTRAINT IF EXISTS discussion_comments_thread_id_fkey,
ADD CONSTRAINT discussion_comments_thread_id_fkey
  FOREIGN KEY (thread_id)
  REFERENCES discussion_threads(id)
  ON DELETE RESTRICT;

ALTER TABLE discussion_mail_reply_tokens
DROP CONSTRAINT IF EXISTS discussion_mail_reply_tokens_thread_id_fkey,
ADD CONSTRAINT discussion_mail_reply_tokens_thread_id_fkey
  FOREIGN KEY (thread_id)
  REFERENCES discussion_threads(id)
  ON DELETE RESTRICT;

ALTER TABLE discussion_threads
DROP CONSTRAINT IF EXISTS discussion_threads_target_repo_id_fk,
ADD CONSTRAINT discussion_threads_target_repo_id_fk
  FOREIGN KEY (target_repo_id)
  REFERENCES discussion_threads_target_repo(id)
  ON DELETE RESTRICT;

ALTER TABLE discussion_threads_target_repo
DROP CONSTRAINT discussion_threads_target_repo_repo_id_fkey,
ADD CONSTRAINT discussion_threads_target_repo_repo_id_fkey
  FOREIGN KEY (repo_id)
  REFERENCES repo(id)
  ON DELETE RESTRICT;

ALTER TABLE repo DROP CONSTRAINT IF EXISTS deleted_at_unused;

COMMIT;
