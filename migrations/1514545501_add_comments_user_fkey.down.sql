BEGIN;
ALTER TABLE comments RENAME COLUMN author_user_id TO author_user_id_new;
ALTER TABLE comments ADD COLUMN author_user_id text;
UPDATE comments SET author_user_id=(SELECT users.auth_id FROM users WHERE users.id=comments.author_user_id_new);
ALTER TABLE comments ALTER COLUMN author_user_id SET NOT NULL;
ALTER TABLE comments DROP COLUMN author_user_id_new;
INSERT INTO comments(id, thread_id, contents, created_at, updated_at, deleted_at, author_name, author_email, author_user_id)
	SELECT id, thread_id, contents, created_at, updated_at, deleted_at, author_name, author_email, author_user_id_old FROM comments_bkup_1514545501;
DROP TABLE comments_bkup_1514545501;
COMMIT;
