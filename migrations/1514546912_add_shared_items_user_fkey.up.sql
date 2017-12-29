BEGIN;
ALTER TABLE shared_items RENAME COLUMN author_user_id TO author_user_id_old;
ALTER TABLE shared_items ADD COLUMN author_user_id integer REFERENCES users (id) ON DELETE RESTRICT;
UPDATE shared_items SET author_user_id=(SELECT users.id FROM users WHERE users.auth_id=shared_items.author_user_id_old);
CREATE TABLE shared_items_bkup_1514546912 AS (SELECT * FROM shared_items WHERE author_user_id IS NULL);
DELETE FROM shared_items WHERE author_user_id IS NULL;
ALTER TABLE shared_items ALTER COLUMN author_user_id SET NOT NULL;
ALTER TABLE shared_items DROP COLUMN author_user_id_old;
COMMIT;
