BEGIN;

TRUNCATE lsif_nearest_uploads;

ALTER TABLE lsif_nearest_uploads ADD COLUMN uploads jsonb NOT NULL;
ALTER TABLE lsif_nearest_uploads DROP COLUMN upload_id;
ALTER TABLE lsif_nearest_uploads DROP COLUMN distance;

COMMIT;
