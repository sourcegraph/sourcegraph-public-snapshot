BEGIN;

UPDATE lsif_nearest_uploads SET "commit"='' WHERE "commit" IS NULL;
ALTER TABLE lsif_nearest_uploads ALTER COLUMN "commit" SET NOT NULL;

COMMIT;
