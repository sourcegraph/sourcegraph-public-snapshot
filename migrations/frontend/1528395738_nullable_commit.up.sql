BEGIN;

ALTER TABLE lsif_nearest_uploads ALTER COLUMN "commit" DROP NOT NULL;

COMMIT;
