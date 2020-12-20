BEGIN;

TRUNCATE lsif_nearest_uploads;

ALTER TABLE lsif_nearest_uploads DROP COLUMN uploads;
ALTER TABLE lsif_nearest_uploads ADD COLUMN upload_id integer NOT NULL;
ALTER TABLE lsif_nearest_uploads ADD COLUMN distance integer NOT NULL;

COMMIT;
