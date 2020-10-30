BEGIN;

UPDATE lsif_nearest_uploads SET commit_bytea = decode("commit", 'hex') WHERE commit_bytea IS NULL;
ALTER TABLE lsif_nearest_uploads ALTER COLUMN "commit" DROP NOT NULL;
ALTER TABLE lsif_nearest_uploads ALTER COLUMN commit_bytea SET NOT NULL;

COMMIT;
