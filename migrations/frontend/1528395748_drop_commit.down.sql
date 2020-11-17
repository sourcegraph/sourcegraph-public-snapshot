BEGIN;

ALTER TABLE lsif_nearest_uploads ADD COLUMN "commit" TEXT;

COMMIT;
