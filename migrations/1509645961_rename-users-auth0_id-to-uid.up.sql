BEGIN;
ALTER TABLE users RENAME COLUMN auth0_id TO uid;
ALTER TABLE users RENAME CONSTRAINT users_auth0_id_key TO users_uid_key;
ALTER TABLE org_settings RENAME COLUMN author_auth0_id TO author_uid;
COMMIT;
