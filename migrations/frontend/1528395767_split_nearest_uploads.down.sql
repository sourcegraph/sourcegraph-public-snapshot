BEGIN;

TRUNCATE lsif_nearest_uploads;

ALTER TABLE lsif_nearest_uploads ADD COLUMN ancestor_visible boolean NOT NULL;
ALTER TABLE lsif_nearest_uploads ADD COLUMN overwritten boolean NOT NULL;

DROP TABLE IF EXISTS lsif_nearest_uploads_links;

DROP INDEX lsif_uploads_visible_at_tip_repository_id_upload_id;
CREATE INDEX lsif_uploads_visible_at_tip_repository_id ON lsif_uploads_visible_at_tip(repository_id);

COMMIT;
