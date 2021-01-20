BEGIN;

ALTER TABLE lsif_nearest_uploads ADD COLUMN commit_bytea bytea;
CREATE INDEX lsif_nearest_uploads_repository_id_commit_bytea ON lsif_nearest_uploads USING btree (repository_id, commit_bytea);

-- Mark all repositories as dirty so that we will refresh them
UPDATE lsif_dirty_repositories SET dirty_token = dirty_token + 1;

COMMIT;
