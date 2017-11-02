BEGIN;
ALTER TABLE org_settings RENAME COLUMN author_uid TO author_auth0_id;
ALTER TABLE users RENAME CONSTRAINT users_uid_key TO users_auth0_id_key;
ALTER TABLE users RENAME COLUMN uid TO auth0_id;
COMMIT;
