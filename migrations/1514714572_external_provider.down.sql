BEGIN;
ALTER TABLE users RENAME COLUMN external_provider TO "provider";
ALTER TABLE users ALTER COLUMN "provider" SET NOT NULL;
ALTER TABLE users ALTER COLUMN "provider" SET DEFAULT '';
COMMIT;
