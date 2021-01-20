BEGIN;

UPDATE lsif_nearest_uploads SET "commit" = encode(commit_bytea, 'hex') WHERE "commit" IS NULL;
ALTER TABLE lsif_nearest_uploads ALTER COLUMN "commit" SET NOT NULL;
ALTER TABLE lsif_nearest_uploads ALTER COLUMN commit_bytea DROP NOT NULL;

COMMIT;
