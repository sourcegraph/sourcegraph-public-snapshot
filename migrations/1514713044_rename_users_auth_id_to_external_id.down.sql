BEGIN;
ALTER TABLE users RENAME COLUMN external_id TO auth_id;
ALTER TABLE users ALTER COLUMN auth_id SET NOT NULL;
ALTER TABLE users RENAME CONSTRAINT users_external_id_key TO users_auth_id_key;
COMMIT;
