DROP INDEX users_external_id;
ALTER TABLE users ADD CONSTRAINT users_external_id_key UNIQUE (external_id);

ALTER TABLE users RENAME COLUMN external_provider TO "provider";
UPDATE users SET provider='' WHERE provider IS NULL;
ALTER TABLE users ALTER COLUMN "provider" SET NOT NULL;
ALTER TABLE users ALTER COLUMN "provider" SET DEFAULT '';
