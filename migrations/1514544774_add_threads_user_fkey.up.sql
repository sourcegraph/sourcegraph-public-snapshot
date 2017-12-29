BEGIN;
ALTER TABLE threads RENAME COLUMN author_user_id TO author_user_id_old;
ALTER TABLE threads ADD COLUMN author_user_id integer REFERENCES users (id) ON DELETE RESTRICT;
UPDATE threads SET author_user_id=(SELECT users.id FROM users WHERE users.auth_id=threads.author_user_id_old);
CREATE TABLE threads_bkup_1514544774 AS (SELECT * FROM threads WHERE author_user_id IS NULL);
DELETE FROM threads WHERE author_user_id IS NULL;
ALTER TABLE threads ALTER COLUMN author_user_id SET NOT NULL;
ALTER TABLE threads DROP COLUMN author_user_id_old;
COMMIT;
