BEGIN;

TRUNCATE lsif_nearest_uploads;

ALTER TABLE lsif_nearest_uploads DROP COLUMN ancestor_visible;
ALTER TABLE lsif_nearest_uploads DROP COLUMN overwritten;

CREATE TABLE IF NOT EXISTS lsif_nearest_uploads_links (
    repository_id int NOT NULL,
    commit_bytea bytea NOT NULL,
    ancestor_commit_bytea bytea NOT NULL,
    distance int NOT NULL
);
CREATE INDEX lsif_nearest_uploads_links_repository_id_commit_bytea ON lsif_nearest_uploads_links USING btree (repository_id, commit_bytea);

DROP INDEX lsif_uploads_visible_at_tip_repository_id;
CREATE INDEX lsif_uploads_visible_at_tip_repository_id_upload_id ON lsif_uploads_visible_at_tip(repository_id, upload_id);

COMMIT;
