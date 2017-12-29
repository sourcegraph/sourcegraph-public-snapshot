BEGIN;
ALTER TABLE comments RENAME COLUMN author_user_id TO author_user_id_old;
ALTER TABLE comments ADD COLUMN author_user_id integer REFERENCES users (id) ON DELETE RESTRICT;
UPDATE comments SET author_user_id=(SELECT users.id FROM users WHERE users.auth_id=comments.author_user_id_old);
CREATE TABLE comments_bkup_1514545501 AS (SELECT * FROM comments WHERE author_user_id IS NULL);
DELETE FROM comments WHERE author_user_id IS NULL;
ALTER TABLE comments ALTER COLUMN author_user_id SET NOT NULL;
ALTER TABLE comments DROP COLUMN author_user_id_old;
COMMIT;
