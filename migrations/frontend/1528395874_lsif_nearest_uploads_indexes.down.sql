BEGIN;

DROP INDEX lsif_nearest_uploads_uploads;
DROP INDEX lsif_nearest_uploads_links_repository_id_ancestor_commit_bytea;

COMMIT;
