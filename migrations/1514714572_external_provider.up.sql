BEGIN;
ALTER TABLE users RENAME COLUMN "provider" TO external_provider;
ALTER TABLE users ALTER COLUMN external_provider DROP NOT NULL;
ALTER TABLE users ALTER COLUMN external_provider DROP DEFAULT;
COMMIT;
