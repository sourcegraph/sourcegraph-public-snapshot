BEGIN;

DROP INDEX IF EXISTS gauge_events_repo_id_btree;
DROP INDEX IF EXISTS gauge_events_repo_name_id_btree;
DROP INDEX IF EXISTS gauge_events_original_repo_name_id_btree;
DROP TABLE IF EXISTS gauge_events;

DROP INDEX IF EXISTS repo_names_name_unique_idx;
DROP INDEX IF EXISTS repo_names_name_trgm;
DROP TABLE IF EXISTS repo_names;

DROP INDEX IF EXISTS metadata_metadata_unique_idx;
DROP INDEX IF EXISTS metadata_metadata_gin;
DROP TABLE IF EXISTS metadata;

COMMIT;
