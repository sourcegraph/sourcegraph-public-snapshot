ALTER TABLE users RENAME COLUMN auth_id TO external_id;
ALTER TABLE users ALTER COLUMN external_id DROP NOT NULL;
ALTER TABLE users RENAME CONSTRAINT users_auth_id_key TO users_external_id_key;
