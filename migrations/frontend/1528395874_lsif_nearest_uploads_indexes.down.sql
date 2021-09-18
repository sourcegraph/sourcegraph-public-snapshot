BEGIN;

DROP INDEX lsif_nearest_uploads_uploads;
DROP INDEX lsif_nearest_uploads_links_repository_id_ancestor_commit_bytea;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
