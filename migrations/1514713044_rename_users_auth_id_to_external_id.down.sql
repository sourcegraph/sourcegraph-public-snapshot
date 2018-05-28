ALTER TABLE users RENAME COLUMN external_id TO auth_id;
UPDATE users SET auth_id='' WHERE auth_id IS NULL;
ALTER TABLE users ALTER COLUMN auth_id SET NOT NULL;
ALTER TABLE users RENAME CONSTRAINT users_external_id_key TO users_auth_id_key;
