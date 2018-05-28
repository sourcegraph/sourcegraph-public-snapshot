ALTER TABLE users RENAME COLUMN "provider" TO external_provider;
ALTER TABLE users ALTER COLUMN external_provider DROP NOT NULL;
ALTER TABLE users ALTER COLUMN external_provider DROP DEFAULT;

ALTER TABLE users DROP CONSTRAINT users_external_id_key;
CREATE UNIQUE INDEX users_external_id ON users(external_id, external_provider) WHERE external_provider IS NOT NULL;
